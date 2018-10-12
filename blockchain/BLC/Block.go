package BLC

import (
    "time"
    "bytes"
    "encoding/gob"
    "log"
    "crypto/sha256"
    "blockchain/Tx"
)

//区块基本结构
type Block struct {
    Index     int64 //区块高度
    TimeStamp int64
    PreHash   []byte
    Hash      []byte
    //Data      []byte
    Txs   []*Tx.Transcation //交易数据
    Nonce int64
}

//创建区块
func NewBlock(pre []byte, txs []*Tx.Transcation, i int64) *Block {
    block := new(Block)
    block.TimeStamp = time.Now().Unix()
    block.PreHash = []byte(pre)
    //block.Data = []byte(data)
    block.Txs = txs
    block.Index = i
    pow := NewPow(block)
    //pow计算区块哈希
    hash, nonce := pow.Run()
    block.Hash = hash
    block.Nonce = nonce
    return block
}

//生成创世区块
func CreateGenesisBlock(txs []*Tx.Transcation) *Block {
    return NewBlock(nil, txs, 1)
}

//序列化，将区块结构序列化
func (block *Block) Serialize() []byte {
    var result bytes.Buffer
    encoder := gob.NewEncoder(&result)
    if err := encoder.Encode(block); err != nil {
        log.Panicf("serialize the block to byte failed! %v\n", err)
    }
    return result.Bytes()
}

//反序列化，将字节数组转化为区块
func DeserializeBlock(blockBytes []byte) *Block {
    var block Block
    decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
    if err := decoder.Decode(&block); nil != err {
        log.Panicf("deserialize the []byte to block failed! %v\n", err)
    }
    return &block
}

//把区块中的所有交易结构转换成[]byte
func (block *Block) HashTransactuions() []byte {
    var txHashes [][]byte
    for _, tx := range block.Txs {
        txHashes = append(txHashes, tx.TxHash)
    }

    txHash := sha256.Sum256(bytes.Join(txHashes, []byte{}))
    return txHash[:]
}
