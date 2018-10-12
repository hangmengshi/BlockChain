package Tx

import (
    "blockchain/BLC"
    "bytes"
    "encoding/gob"
    "log"
    "github.com/boltdb/bolt"
    "encoding/hex"
    "fmt"
    "blockchain/Wallet"
)

//持久化

//utxo表名
const utxoTableName = "utxoTable"

//保存指定区块链所有的UTXO
type UTXOSet struct {
    BlockChain *BLC.BlockChain
}

//将utxo集合序列化
func (tx *AllUTXO) Seriallize() []byte {
    var result bytes.Buffer
    encoder := gob.NewEncoder(&result)
    err := encoder.Encode(tx)
    if err != nil {
        log.Printf("serialize the utxo failed! %v\n", err)
    }
    return result.Bytes()
}

//反序列化
func Deserialize(allBytes []byte) *AllUTXO {

    var all AllUTXO
    decoder := gob.NewDecoder(bytes.NewReader(allBytes))
    err := decoder.Decode(&all)
    if err != nil {
        log.Printf("deserialize the struct of AllUTXO! %v\n ", err)

    }
    return &all

}

//重置UTXO
func (utxoSet *UTXOSet) ResetUTXOSet() {
    //采用覆盖的方式更新utxo table
    err := utxoSet.BlockChain.Db.Update(func(tx *bolt.Tx) error {
        //查找utxo表
        b := tx.Bucket([]byte(utxoTableName))
        if b != nil {
	tx.DeleteBucket([]byte(utxoTableName))
        }
        c, _ := tx.CreateBucket([]byte(utxoTableName))
        if c != nil {
	//查找所有未花费的输出
	all := utxoSet.BlockChain.FindUTXOMap()
	for keyHash, output := range all {
	    txHash, _ := hex.DecodeString(keyHash)
	    //存入表
	    err := c.Put(txHash, output.Seriallize())
	    if err != nil {
	        log.Panicf("put the utxo into table failed! %v\n", err)
	    }

	}
        }
        return nil
    })
    if err != nil {
        log.Panicf("updata the db of utxoset failed! %v\n", err)
    }

}

//获取余额
func (utxoset *UTXOSet) GetBalance(address string) int64 {
    //获取指定地址的UTXO
    UTXOS := utxoset.FindUTXOWithAddress(address)
    var amount int64
    for _, utxo := range UTXOS {
        fmt.Printf("\tutxo-hash : %x\n", utxo.TxHash)
        fmt.Printf("\tutxo-index : %d\n", utxo.Index)
        fmt.Printf("\t\tutxo-Ripemd160Hash : %x\n", utxo.Output.Ripemd160Hash)
        fmt.Printf("\t\tutxo-value : %d\n", utxo.Output.Value)
        amount += utxo.Output.Value

    }
    return amount
}

//查找指定地址的UTXO
func (utxoSet *UTXOSet) FindUTXOWithAddress(address string) [] *UTXO {
    var utxos []*UTXO
    //查找数据库中的utxoTable表
    utxoSet.BlockChain.Db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(utxoTableName))
        if b != nil {
	// 游标
	c := b.Cursor()
	//遍历每一个UTXO
	for k, v := c.First(); k != nil; k, v = c.Next() {
	    // k -> 交易哈希
	    // v -> 输出结构的字节数组
	    allUTXO := Deserialize(v)
	    for _, utxo := range allUTXO.UTXO {
	        if utxo.Output.UnlockScriptPubkeyWithAddress(address) {
		utxos = append(utxos, utxo)
	        }

	    }
	}
        }
        return nil
    })
    return utxos
}

//查找可花费的UTXO
func (utxoSet *UTXOSet) FindSpendableUTXO(from string, amount int64, txs []*Transcation) (int64, map[string][]int) {
    //从未打包的交易中获取UTXO，如果足够，不再查询UTXO Tablie
    spendableUTXO := make(map[string][]int)
    //查找未打包交易中的的UTXO
    unPackagesUTXOs := utxoSet.FindUnPackageSpendableUTXOs(from,txs)
    var value  int64=0

    for _, utxo := range unPackagesUTXOs {
        value += utxo.Output.Value
        txHash := hex.EncodeToString(utxo.TxHash)
        spendableUTXO[txHash] = append(spendableUTXO[txHash], utxo.Index)
        if value >= amount {
	return value, spendableUTXO
        }
    }
    //  在获取到未打包交易中的UTXO集合之后，金额仍然不足，从UTXO集合表中获取
    utxoSet.BlockChain.Db.View(func(tx *bolt.Tx) error {
        // 先获取表
        b := tx.Bucket([]byte(utxoTableName))
        if b != nil {
	cursor := b.Cursor() // 有序遍历
        UTXOBREAK:
	for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
	    txOutputs := DeserializeTXOutputs(v)
	    for _, utxo := range txOutputs.UTXOS {
	        value += utxo.Output.Value
	        if value >= amount {
		txHash := hex.EncodeToString(utxo.TxHash)
		spendableUTXO[txHash] = append(spendableUTXO[txHash], utxo.Index)
		break UTXOBREAK
	        }
	    }
	}
        }
        return nil
    })

    if value < amount {
        log.Panic("余额不足......")
    }
    return value, spendableUTXO
}

//从未打包的交易中进行查找
func (utxoset *UTXOSet) FindUnPackageSpendableUTXOs(from string, txs []*Transcation) []*UTXO {
    // 未打包交易中的UTXO
    var unUTXOs []*UTXO
    // 每个交易中的已花费输出(索引)
    spendTXOutputs := make(map[string][]int)
    for _, tx := range txs {
        //排队coinbase交易
        if !tx.IsCoinbaseTransaction() {
	for _, vin := range tx.Vins {
	    pubKeyHash := Wallet.Base58Decode([]byte(from))    // 解码，获取公钥哈希
	    ripemd160Hash := pubKeyHash[1 : len(pubKeyHash)-4] // 用户名
	    if vin.UnlockRipemd160Hash(ripemd160Hash) { // 解锁
	        key := hex.EncodeToString(vin.TxHash)
	        spendTXOutputs[key] = append(spendTXOutputs[key], vin.Vout)
	    }
	}
        }

    }
    for _, tx := range txs {
    UnUtxoLoop:
        for index, vout := range tx.Vouts {
	// 判断该vout是否属于from
	if vout.UnlockScriptPubkeyWithAddress(from) {
	    // 在没有包含已花费输出的情况
	    if len(spendTXOutputs) == 0 {
	        utxo := &UTXO{tx.TxHash, index, vout}
	        unUTXOs = append(unUTXOs, utxo)
	    } else {
	        for hash, indexArray := range spendTXOutputs {
		txHashStr := hex.EncodeToString(tx.TxHash)
		// 判断当前交易是否包含了已花费输出
		if hash == txHashStr {
		    var isUnpkgSpentUTXO bool // 判断该输出是否属于已花费输出
		    for _, idx := range indexArray {
		        if index == idx {
			isUnpkgSpentUTXO = true
			continue UnUtxoLoop
		        }
		    }
		    if isUnpkgSpentUTXO == false {
		        utxo := &UTXO{tx.TxHash, index, vout}
		        unUTXOs = append(unUTXOs, utxo)
		    }

		} else {
		    // 该交易没有包含已花费输出
		    utxo := &UTXO{tx.TxHash, index, vout}
		    unUTXOs = append(unUTXOs, utxo)
		}
	        }
	    }
	}
        }
    }
    return unUTXOs

}


// 实现 Utxo table实时更新
func (utxoSet *UTXOSet) Update() {
    // 找到需要删除的UTXO
    //  获取最新的区块
    latest_block := utxoSet.BlockChain.Iterator().Next()
    inputs := []*TxInput{} // 存放最新区块的所有输入
    // 获取需要存入utxo table中的UTXO
    outsMap := make(map[string]*AllUTXO)

    // 查找需要删除的数据
    for _, tx := range latest_block.Txs {
        // 遍历输入
        for _, vin := range tx.Vins {
	inputs = append(inputs, vin)
        }
    }
    // 获取当前最新区块的所有UTXO
    for _, tx := range latest_block.Txs {
        utxos := []*UTXO{}
        for index, out := range tx.Vouts {
	isSpent := false
	for _, in := range inputs {
	    if in.Vout == index && bytes.Compare(tx.TxHash, in.TxHash) == 0 {
	        if bytes.Compare(out.Ripemd160Hash, Wallet.Ripemd160Hash(in.PublicKey)) == 0 {
		isSpent = true
		continue
	        }
	    }
	}
	if isSpent == false {
	    utxo := &UTXO{tx.TxHash, index, out}
	    utxos = append(utxos, utxo)
	}
        }
        if len(utxos) > 0 {
	txHash := hex.EncodeToString(tx.TxHash)
	outsMap[txHash] = &AllUTXO{utxos}
        }
    }

    // 更新
    err := utxoSet.BlockChain.Db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(utxoTableName))
        if nil != b {
	// 删除已花费输出
	for _, in := range inputs {
	    txOutputsBytes := b.Get(in.TxHash) // 查找当前input所引用的交易哈希
	    if len(txOutputsBytes) == 0 {
	        continue
	    }
	    UTXOS := []*UTXO{}
	    // 反序列化
	    txOutpus := Deserialize(txOutputsBytes)
	    isNeedToDel := false
	    for _, utxo := range txOutpus.UTXO{
	        // 判断是哪一个输出被引用
	        if in.Vout == utxo.Index {
		if bytes.Compare(utxo.Output.Ripemd160Hash, Wallet.Ripemd160Hash(in.PublicKey)) == 0 {
		    isNeedToDel = true // 该输出已经被引用，需要删除
		}
	        } else {
		UTXOS = append(UTXOS, utxo)
	        }
	    }

	    if isNeedToDel {
	        // 先删除输出
	        b.Delete(in.TxHash)
	        if len(UTXOS) > 0 {
		preTXOutputs := outsMap[hex.EncodeToString(in.TxHash)]
		preTXOutputs.UTXO = append(preTXOutputs.UTXO, UTXOS...)
		// 更新
		outsMap[hex.EncodeToString(in.TxHash)] = preTXOutputs
	        }
	    }

	}

	for hash, outputs := range outsMap {
	    hashBytes, _ := hex.DecodeString(hash)
	    err := b.Put(hashBytes, outputs.Seriallize())
	    if nil != err {
	        log.Panicf("put the utxo to table failed! %v\n", err)
	    }
	}
        }

        return nil
    })

    if nil != err {
        log.Printf("update the UTXOMap to utxo table failed! %v", err)
    }
}