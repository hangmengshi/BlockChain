package Tx

import (
    "bytes"
    "encoding/gob"
    "log"
    "crypto/sha256"
    "blockchain/BLC"
    "blockchain/Wallet"
    "fmt"
    "encoding/hex"
    "crypto/ecdsa"
)

//交易相关
type Transcation struct {
    //交易的唯一标识
    TxHash []byte
    //输入
    Vins []*TxInput
    //输出
    Vouts []*TxOutput
}

//生成交易哈希
func (tx *Transcation) TranscationHash() {
    var result bytes.Buffer
    encoder := gob.NewEncoder(&result)
    err := encoder.Encode(tx)
    if err != nil {
        log.Printf("tx hash generate failed! %v\n", err)
    }
    hash := sha256.Sum256(result.Bytes())
    tx.TxHash = hash[:]
}

//生成coinbase交易(交易奖励币)
func NewCoinbaseTra(address string) *Transcation {
    //输入
    txInput := &TxInput{[]byte{}, -1, nil, nil}
    //输出
    txOutput := NewTxOutput(10, address)
    //交易
    txCoinbase := &Transcation{nil, []*TxInput{txInput}, []*TxOutput{txOutput}}
    //交易hash
    txCoinbase.TranscationHash()
    return txCoinbase
}

//生成转账交易
func NewTransaction(from string, to string, amount int, block *BLC.BlockChain, txs []*Transcation, utxoSet *UTXOSet, nodeID string) *Transcation {

    var txInputs []*TxInput
    var txOutputs []*TxOutput
    //查找指定地址的可用UTXO
    money, spendableUTXO := utxoSet.FindSpendableUTXO(from, int64(amount), txs)
    //fmt.Printf("money: %v\n", money)
    //获取钱包集合
    wallets, err := Wallet.NewWallets(nodeID)
    if err != nil {
        log.Printf("NewWallets failed %v\n", err)
        }
        //指定地址对应的钱包结构
    wallet := wallets.Wallets[from]
    fmt.Printf("spendableUTXODic: %v\n", spendableUTXO)
    for txHash,indexArrey := range spendableUTXO {
        //fmt.Printf("indexArray: %v\n",indexArrey)
        txHashBytes, _ := hex.DecodeString(txHash)
        for _,index := range indexArrey {
            // 此处的输出是需要消费的，必然会被其它的交易的输入所引用
            txInput := &TxInput{txHashBytes,index, nil, wallet.PublicKey}
            txInputs = append(txInputs, txInput)
        }

    }
    //转账
    txOutput := NewTxOutput(int64(amount), to)
    txOutputs = append(txOutputs, txOutput)
    //找零
    Output := NewTxOutput(money-int64(amount), from)
    txOutputs = append(txOutputs, Output)
    //生成交易
    tx:=&Transcation{nil,txInputs,txOutputs}
    tx.TranscationHash()
    //对交易进行签名，参数主要tx，wallet，PrivateKey
    for _,vin := range tx.Vins {
        //查找所引用的每一个交易
        fmt.Printf("transaction.go  hash : [%x]\n", vin)
    }
    //交易签名
    block.

}

//判断交易是否是一个coinbase交易
func (tx *Transcation) IsCoinbaseTransaction() bool {

    return len(tx.Vins[0].TxHash) == 0 && tx.Vins[0].Vout == -1
}

//交易签名
func (tx *Transcation)Sign(privKey ecdsa.PrivateKey,prevTxs map[string]Transcation)  {
    //判断是否是挖矿交易
    if tx.IsCoinbaseTransaction() {
        return
    }
    for _, vin := range tx.Vins {
        if prevTxs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
            log.Panicf("ERROR:Prev transaction is not correct\n")
        }
    }
    //提取需要签名属性
    tx.TrimmedCopy()
}

//添加一个交易的拷贝，用于交易签名，返回需要进行签名的交易
func ()  {
    
}