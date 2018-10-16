package BLC

import (
    "github.com/boltdb/bolt"
    "os"
    "fmt"
    "log"
    "math/big"
    "encoding/hex"
    "bytes"
    "blockchain/Wallet"
    "crypto/ecdsa"
    "strconv"
    "encoding/gob"
)

const dbName = "bc_%s.db"       //存储区块数据的数据库文件
const blockTableName = "blocks" //表名称

//定义区块链结构
type BlockChain struct {
    Db  *bolt.DB
    Tip []byte
}

//判断数据库文件是否存在
func DbExists(nodeID string) bool {
    Name := fmt.Sprintf(dbName, nodeID)
    if _, err := os.Stat(Name); os.IsNotExist(err) {
        return false
    }

    return true
}

//初始化区块链
func CreateBloxkChainWithGenesisBlock(address string, nodeID string) *BlockChain {
    if DbExists(nodeID) {
        fmt.Println("创世区块已经存在")
        os.Exit(1)
    }
    dbName := fmt.Sprintf(dbName, nodeID)
    // 创建或者打开数据
    db, err := bolt.Open(dbName, 0600, nil)
    if nil != err {
        log.Panicf("open the db failed! %v\n", err)
    }
    var blockHash []byte // 需要存储到数据库中的区块哈希
    err = db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blockTableName))
        if nil == b {
	// 添加创世区块
	b, err = tx.CreateBucket([]byte(blockTableName))
	if nil != err {
	    log.Panicf("create the bucket [%s] failed! %v\n", blockTableName, err)
	}
        }
        if nil != b {
	// 生成交易
	txCoinbase := NewCoinbaseTra(address)
	// 生成创世区块
	genesisBlock := CreateGenesisBlock([]*Transcation{txCoinbase})
	err = b.Put(genesisBlock.Hash, genesisBlock.Serialize())
	if nil != err {
	    log.Panicf("put the data of genesisBlock to db failed! %v\n", err)
	}
	// 存储最新区块的哈希
	err = b.Put([]byte("l"), genesisBlock.Hash)
	if nil != err {
	    log.Panicf("put the hash of latest block to db failed! %v\n", err)
	}
	blockHash = genesisBlock.Hash
        }
        return nil
    })
    if nil != err {
        log.Panicf("update the data of genesis block failed! %v\n", err)
    }
    return &BlockChain{db, blockHash}
}

// 返回Blockchain 对象
func BlockchainObject(nodeID string) *BlockChain {
    dbName := fmt.Sprintf(dbName, nodeID)
    // 读取数据库
    db, err := bolt.Open(dbName, 0600, nil)
    if nil != err {
        log.Panicf("get the object of blockchain failed! %v\n", err)
    }
    var tip []byte // 最新区块的哈希值
    err = db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blockTableName))
        if nil != b {
	tip = b.Get([]byte("l"))
        }
        return nil
    })
    return &BlockChain{db, tip}
}

//添加新的区块到区块链中
func (blockchain *BlockChain) AddBlock(txs []*Transcation) {
    //更新数据
    err := blockchain.Db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blockTableName))
        if b != nil {
	blockBytes := b.Get(blockchain.Tip)
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
	blockchain.Tip = block.Hash
        }
        return nil
    })
    if nil != err {
        log.Panicf("update the db of block failed! %v\n", err)
    }
}

// 通过接收交易，进行打包确认，最终生成新的区块
func (blockchain *BlockChain) MineNewBlock(from []string, to []string, amount []string, nodeID string) {
    fmt.Printf("\tFROM:[%s]\n", from)
    fmt.Printf("\tTO:[%s]\n", to)
    fmt.Printf("\tAMOUNT:[%s]\n", amount)
    // 接收交易
    var txs []*Transcation
    for index, address := range from {
        fmt.Printf("\tfrom:[%s], to[%s], amount:[%s]\n", address, to[index], amount[index])
        value, _ := strconv.Atoi(amount[index])
        utxoSet := &UTXOSet{blockchain}
        tx := NewTransaction(address, to[index], value, blockchain, txs, utxoSet, nodeID)
        txs = append(txs, tx)
        fmt.Printf("\ttx-hash:%x, tx-vouts:%v, tx-vins:%v\n", tx.TxHash, tx.Vouts, tx.Vins)
    }
    // 给矿工一定的奖励
    // 默认情况下，设置地址列表中的第一个地址为矿工奖励地址
    tx := NewCoinbaseTra(from[0])
    txs = append(txs, tx)
    // 打包交易
    // 生成新的区块
    var block *Block
    // 从数据库中获取最新区块
    blockchain.Db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blockTableName))
        if nil != b {
	hash := b.Get([]byte("l"))           // 获取最新区块哈希值(当作新生区块的prevHash)
	blockBytes := b.Get(hash)            // 得到最新区块(为了获取区块高度)
	block = DeserializeBlock(blockBytes) // 反序列化
        }
        return nil
    })
    // 在生成新区块之前，对交易签名进行验证
    // 在这里验证一下交易签名
    _txs := []*Transcation{} // 未打包的关联交易
    for _, tx := range txs {
        // 验证每一笔交易
        // 第二笔交易引用了第一笔交易的UTXO作为输入
        // 第一笔交易还没有被打包到区块中，所以添加到缓存列表中
        fmt.Printf("txHash : %v\n", tx.TxHash)
        if !blockchain.VerifyTransaction(tx, _txs) {
	log.Panic("ERROR : tx [%x] verify failed!")
        }
        _txs = append(_txs, tx)
    }
    // 生成新的区块
    block = NewBlock(block.Hash, txs, block.Index+1)
    // 持久化新区块
    blockchain.Db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blockTableName))
        if nil != b {
	err := b.Put(block.Hash, block.Serialize())
	if nil != err {
	    log.Panicf("update the new block to db failed! %v\n", err)
	}
	b.Put([]byte("l"), block.Hash) // 更新数据库中的最新哈希值
	blockchain.Tip = block.Hash
        }
        return nil
    })
}

// 遍历输出区块链所有区块的信息
func (blockchain *BlockChain) PrintChain() {
    fmt.Println("区块链完整信息...")
    var block *Block
    //创建迭代器对象
    iterator := blockchain.Iterator()
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
        }
        //判断是否已经遍历到创世区块
        if big.NewInt(0).Cmp(big.NewInt(0).SetBytes(block.PreHash)) == 0 {
	break
        }
    }

}

//查找所有UTXO
func (blockchain *BlockChain) FindUTXOMap() map[string]*AllUTXO {
    iterator := blockchain.Iterator()
    //存储已花费的UTXO的信息
    //key：代表指定的交易哈希
    //value：代表所有引用了该交易output的输入
    spentUTXOMap := make(map[string][]*TxInput)
    //UTXO集合
    //Key：指定交易哈希
    //value：该交易中所有的未花费输出
    utxoMap := make(map[string]*AllUTXO)
    for {
        block := iterator.Next()
        //遍历每个区块中的交易
        for i := len(block.Txs) - 1; i >= 0; i-- {
	//保存输出的列表
	all := &AllUTXO{[]*UTXO{}}
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
		utxo := UTXO{tx.TxHash, index, out}
		all.UTXO = append(all.UTXO, &utxo)
	        }

	    } else {
	        //如果没有input，都是未花费的输出
	        utxo := UTXO{tx.TxHash, index, out}
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
func (blockchain *BlockChain) FindTransaction(id []byte, txs []*Transcation) Transcation {
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

    return Transcation{}
}

//区块交易签名
func (blockchain *BlockChain) SignTransaction(tx *Transcation, privateKey ecdsa.PrivateKey, txs []*Transcation) {
    //coinbase交易不需要签名
    if tx.IsCoinbaseTransaction() {
        return
    }
    //处理input，查找tx的input所引用的vout所属的交易
    prevTXs := make(map[string]Transcation)
    for _, vin := range tx.Vins {
        //查找所引用的每一个交易
        fmt.Sprintf("hash;[%x]\n", vin.TxHash)
        prevTX := blockchain.FindTransaction(vin.TxHash, txs)

        prevTXs[hex.EncodeToString(prevTX.TxHash)] = prevTX

    }
    //实现签名函数
    tx.Sign(privateKey, prevTXs)

}

// 验证签名
func (blockchain *BlockChain) VerifyTransaction(tx *Transcation, txs []*Transcation) bool {
    // 查找指定交易的关联交易
    prevTxs := make(map[string]Transcation)
    for _, vin := range tx.Vins {
        prevTx := blockchain.FindTransaction(vin.TxHash, txs)
        prevTxs[hex.EncodeToString(prevTx.TxHash)] = prevTx
    }

    return tx.Verify(prevTxs)
}

// 将字节数组转成cmd
func bytesToCommand(bytes []byte) string {
    var command []byte // 接收命令
    for _, b := range bytes {
        if b != 0x0 {
	command = append(command, b)
        }
    }
    return fmt.Sprintf("%s", command)
}

// 将结构体序列化为字节数组
func gobEncode(data interface{}) []byte {
    var buff bytes.Buffer
    enc := gob.NewEncoder(&buff)
    err := enc.Encode(data)
    if nil != err {
        log.Panicf("encode the data failed! %v\n", err)
    }
    return buff.Bytes()
}

// 将命令转为字节数组
// 指令长度最长为12位
func commandToBytes(command string) []byte {
    var bts [12]byte // 命令长度
    for i, c := range command {
        bts[i] = byte(c) // 转换
    }
    return bts[:]
}
