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
    "time"
    "blockchain/utl"
    "crypto/rand"
    "crypto/elliptic"
    "math/big"
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
    //添加时间戳标识， 没有时间标识会导致所有的coinbase交易的哈希完全一样
    tm := time.Now().Unix()
    join := bytes.Join([][]byte{result.Bytes(), utl.IntToHex(tm)}, []byte{})
    hash := sha256.Sum256(join)
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
    for txHash, indexArrey := range spendableUTXO {
        //fmt.Printf("indexArray: %v\n",indexArrey)
        txHashBytes, _ := hex.DecodeString(txHash)
        for _, index := range indexArrey {
	// 此处的输出是需要消费的，必然会被其它的交易的输入所引用
	txInput := &TxInput{txHashBytes, index, nil, wallet.PublicKey}
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
    tx := &Transcation{nil, txInputs, txOutputs}
    tx.TranscationHash()
    //对交易进行签名，参数主要tx，wallet，PrivateKey
    for _, vin := range tx.Vins {
        //查找所引用的每一个交易
        fmt.Printf("transaction.go  hash : [%x]\n", vin)
    }
    //交易签名
    block.SignTransaction(tx, wallet.PrivateKey, txs)
    return tx

}

//判断交易是否是一个coinbase交易
func (tx *Transcation) IsCoinbaseTransaction() bool {

    return len(tx.Vins[0].TxHash) == 0 && tx.Vins[0].Vout == -1
}

//交易签名
func (tx *Transcation) Sign(privKey ecdsa.PrivateKey, prevTxs map[string]Transcation) {
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
    txCopy := tx.TrimmedCopy()
    for id, vin := range txCopy.Vins {
        prevTx := prevTxs[hex.EncodeToString(vin.TxHash)]
        txCopy.Vins[id].Signature = nil
        txCopy.Vins[id].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash
        txCopy.TxHash = txCopy.Hash()
        txCopy.Vins[id].PublicKey = nil
        r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.TxHash)
        if nil != err {
	log.Panicf("sign to tx %x failed! %v\n", tx.TxHash, err)
        }
        signature := append(r.Bytes(), s.Bytes()...)
        tx.Vins[id].Signature = signature
    }
}

//设置用于签名的数据哈希
func (tx *Transcation) Hash() []byte {
    txCopy := tx
    txCopy.TxHash = []byte{}
    hash := sha256.Sum256(txCopy.Serialize())
    return hash[:]
}

// 序列化
func (tx *Transcation) Serialize() []byte {
    var result bytes.Buffer
    encoder := gob.NewEncoder(&result) //新建eoncode对象
    if err := encoder.Encode(tx); nil != err { // 编码
        log.Panicf("serialize the tx to byte failed! %v\n", err)
    }
    return result.Bytes()
}

//添加一个交易的拷贝，用于交易签名，返回需要进行签名的交易
func (tx *Transcation) TrimmedCopy() Transcation {
    var inputs []*TxInput
    var outputs []*TxOutput
    for _, vin := range tx.Vins {
        inputs = append(inputs, &TxInput{vin.TxHash, vin.Vout, nil, nil})
    }
    for _, vout := range tx.Vouts {
        outputs = append(outputs, &TxOutput{vout.Value, vout.Ripemd160Hash})
    }
    txCopy := Transcation{tx.TxHash, inputs, outputs}
    return txCopy

}

//交易验证
func (tx *Transcation) Verify(prevTxs map[string]Transcation) bool {
    if tx.IsCoinbaseTransaction() {
        return true
    }
    // 查找每个vin所引用的交易hash是否包含在prevTxs
    for _, vin := range tx.Vins {
        if prevTxs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
	log.Panic("ERROR: Tx is Incorrect")
        }
    }
    txCopy := tx.TrimmedCopy()
    // 使用相同的椭圆获取密钥对
    curve := elliptic.P256()
    for vinId, vin := range tx.Vins {
        prevTx := prevTxs[hex.EncodeToString(vin.TxHash)]
        txCopy.Vins[vinId].Signature = nil
        txCopy.Vins[vinId].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash
        // 上面是生成哈希的数据，所以要与签名时的数据完全一致
        txCopy.TxHash = txCopy.Hash() // 要验证的数据
        txCopy.Vins[vinId].PublicKey = nil
        // r, s代表签名
        r := big.Int{}
        s := big.Int{}
        sigLen := len(vin.Signature)
        r.SetBytes(vin.Signature[:(sigLen / 2)])
        s.SetBytes(vin.Signature[(sigLen / 2):])
        // 生成x,y(首先，签名是一个数字对，公钥是X,Y坐标组合，
        // 在生成公钥时，需要将X Y坐标组合到一起，在验证的时候，需要将
        x := big.Int{}
        y := big.Int{}
        pubkeyLen := len(vin.PublicKey)
        x.SetBytes(vin.PublicKey[:(pubkeyLen / 2)])
        y.SetBytes(vin.PublicKey[(pubkeyLen / 2):])
        // 生成验证签名所需的公钥
        rawPublicKey := ecdsa.PublicKey{curve, &x, &y}

        // 验证签名
        if !ecdsa.Verify(&rawPublicKey, txCopy.TxHash, &r, &s) {
	return false
        }
    }

    return true
}
