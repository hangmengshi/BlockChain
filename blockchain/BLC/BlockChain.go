package BLC

import (
    "github.com/boltdb/bolt"
    "os"
    "fmt"
    "log"

    "math/big"
    "blockchain/Tx"
    "encoding/hex"
    "bytes"
    "blockchain/Wallet"
    "crypto/ecdsa"
)

const dbName = "bc.db"          //存储区块数据的数据库文件
const blockTableName = "blocks" //表名称

//定义区块链结构
type BlockChain struct {
    Db  *bolt.DB
    Tip []byte
}

//判断数据库文件是否存在
func dbExists() bool {
    if _, err := os.Stat(dbName); os.IsNotExist(err) {
        return false
    }
    return true
}

//初始化区块链
func CreateBloxkChainWithGenesisBlock(address string) *BlockChain {
    if dbExists() {
        fmt.Println("创世区块已经存在")
        os.Exit(1)
    }
    //创建或者打开数据
    db, e := bolt.Open(dbName, 0600, nil)
    if e != nil {
        log.Panicf("open the db failed! %v\n", e)
    }
    var blockHash []byte
    db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blockTableName))
        if b == nil {
	b, e = tx.CreateBucket([]byte(blockTableName))
	if e != nil {
	    log.Panicf("create the bucket [%s] failed! %v\n", blockTableName, e)
	}
        }
        if b != nil {
	//生成交易
	txCoinbase := NewCoinbaseTransaction(address)
	//生成创世区块
	block := CreateGenesisBlock([]*Tx.Transcation{txCoinbase})
	err := b.Put(block.Hash, block.Serialize())
	if nil != err {
	    log.Panicf("put the data of genesisBlock to db failed! %v\n", err)
	}
	//存储最新区块的哈希
	err = b.Put([]byte("1"), block.Hash)
	if err != nil {
	    log.Panicf("put the hash of latest block to db failed! %v\n", err)
	}
	blockHash = block.Hash
        }
        return nil
    })
    if nil != e {
        log.Panicf("update the data of genesis block failed! %v\n", e)
    }
    return &BlockChain{db, blockHash}
}

//添加新的区块到区块链中
func (bc *BlockChain) AddBlock(txs []*Tx.Transcation) {
    //更新数据
    err := bc.Db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blockTableName))
        if b != nil {
	blockBytes := b.Get(bc.Tip)
	last_block := DeserializeBlock(blockBytes)
	block := NewBlock(last_block.Hash, txs, last_block.Index+1)
	err := b.Put(block.Hash, block.Serialize())
	if err != nil {
	    log.Panicf("put the data of new block into db failed! %v\n", err)
	}
	// 6. 更新最新区块的哈希
	err = b.Put([]byte("l"), block.Hash)
	if nil != err {
	    log.Panicf("put the hash of the newest block into db failed! %v\n", err)
	}
	bc.Tip = block.Hash
        }
        return nil
    })
    if nil != err {
        log.Panicf("update the db of block failed! %v\n", err)
    }
}

// 遍历输出区块链所有区块的信息
func (bc *BlockChain) PrintChain() {
    fmt.Println("区块链完整信息...")
    var block *Block
    //创建迭代器对象
    iterator := bc.Iterator()
    for {
        fmt.Printf("-----------------------------------------\n")
        block = iterator.Next()
        fmt.Printf("\t Height: %d\n", block.Index)
        fmt.Printf("\t TimeStamp: %d\n", block.TimeStamp)
        fmt.Printf("\t PreHash: %x\n", block.PreHash)
        fmt.Printf("\t hash: %v\n", block.Hash)
        fmt.Printf("\t Transaction: %v\n", block.Txs)
        //打印交易信息
        for _, tx := range block.Txs {
	fmt.Printf("\t\t tx_hash: %x\n", tx.TxHash)
	fmt.Println("\t\t 交易输入。。。")
	//for _,vin := range tx.Vins {
	//  fmt.Printf("\t\t\t in-txhash:%x\n",vin.Txhash)
	//}
        }
        //判断是否已经遍历到创世区块
        if big.NewInt(0).Cmp(big.NewInt(0).SetBytes(block.PreHash)) == 0 {
	break
        }
    }

}

//返回Blockchain对象
func BlockchainObjiect(nodeID string) *BlockChain {

}

//查找所有UTXO
func (blockchain *BlockChain) FindUTXOMap() map[string]*Tx.AllUTXO {
    iterator := blockchain.Iterator()
    //存储已花费的UTXO的信息
    //key：代表指定的交易哈希
    //value：代表所有引用了该交易output的输入
    spentUTXOMap := make(map[string][]*Tx.TxInput)
    //UTXO集合
    //Key：指定交易哈希
    //value：该交易中所有的未花费输出
    utxoMap := make(map[string]*Tx.AllUTXO)
    for {
        block := iterator.Next()
        //遍历每个区块中的交易
        for i := len(block.Txs) - 1; i >= 0; i-- {
	//保存输出的列表
	all := &Tx.AllUTXO{[]*Tx.UTXO{}}
	//获取每一个交易
	tx := block.Txs[i]
	//判断是否一个coinbase交易
	if tx.IsCoinbaseTransaction() == false {
	    fmt.Printf("tx-hash:%x\n", tx.TxHash)
	    //遍历交易中的每一个输入
	    for _, txInput := range tx.Vins {
	        // 当前输入所引用的输出所在的交易哈希
	        txHash := hex.EncodeToString(txInput.TxHash)
	        spentUTXOMap[txHash] = append(spentUTXOMap[txHash], txInput)
	    }
	} else {
	    fmt.Printf("coinbase tx-hash: %x\n", tx.TxHash)
	}
	//遍历输出
	txHash := hex.EncodeToString(tx.TxHash)
        WorkOutLoop:
	for index, out := range tx.Vouts {
	    // 查找指定哈希的关联输入
	    txInputs := spentUTXOMap[txHash]
	    if len(txInputs) > 0 {
	        //判断output是否已经被花费
	        isSpent := false
	        for _, in := range txInputs {
		outAddress := out.Ripemd160Hash
		inAddress := Wallet.Ripemd160Hash(in.PublicKey)
		// 检查input和output中的用户是否是同一个
		if bytes.Compare(outAddress, inAddress) == 0 {
		    if index == in.Vout {
		        isSpent = true //已花费
		        continue WorkOutLoop
		    }

		}

	        }
	        if isSpent == false {
		// isSpent为假，说明该交易相关的输入中没输入能够与当前判断的out相匹配
		utxo := Tx.UTXO{tx.TxHash, index, out}
		all.UTXO = append(all.UTXO, &utxo)
	        }

	    } else {
	        //如果没有input，都是未花费的输出
	        utxo := Tx.UTXO{tx.TxHash, index, out}
	        all.UTXO = append(all.UTXO, &utxo)
	    }
	}
	//该交易所有的UTXO
	utxoMap[txHash] = all
        }
        //退出条件
        var hashInt big.Int
        hashInt.SetBytes(block.PreHash)
        if hashInt.Cmp(big.NewInt(0)) == 0 {
	break
        }
    }
    return utxoMap
}

//查找指定的交易
func (blockchain *BlockChain) FindTransaction(id []byte, txs []*Tx.Transcation) Tx.Transcation {
    //查找缓存中是否有符合条件的关联交易
    for _, tx := range txs {
        if bytes.Compare(tx.TxHash, id) == 0 {
	return *tx
        }

    }
    bcit := blockchain.Iterator()
    for {
        block := bcit.Next()
        for _, tx := range block.Txs {
	//判断交易哈希是否相等
	if bytes.Compare(tx.TxHash, id) == 0 {
	    return *tx
	}

        }
        var hashInt big.Int
        hashInt.SetBytes(block.PreHash)
        if big.NewInt(0).Cmp(&hashInt) == 0 {
	break
        }
    }

    return Tx.Transcation{}
}

//区块交易签名
func (blockchain *BlockChain) SignTransaction(tx *Tx.Transcation, privateKey ecdsa.PrivateKey, txs []*Tx.Transcation) {
    //coinbase交易不需要签名
    if tx.IsCoinbaseTransaction() {
        return
    }
    //处理input，查找tx的input所引用的vout所属的交易
    prevTXs := make(map[string]Tx.Transcation)
    for _, vin := range tx.Vins {
        //查找所引用的每一个交易
        fmt.Sprintf("hash;[%x]\n", vin.TxHash)
        prevTX := blockchain.FindTransaction(vin.TxHash, txs)

        prevTXs[hex.EncodeToString(prevTX.TxHash)] = prevTX

    }
    //实现签名函数

}
