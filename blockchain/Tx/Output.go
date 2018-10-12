package Tx

import (
    "blockchain/Wallet"
    "bytes"
)

type TxOutput struct {
    //金额
    Value int64
    //用户（公钥哈希）
    Ripemd160Hash []byte
}

//身份验证
func (out *TxOutput) UnlockScriptPubkeyWithAddress(address string) bool {
    b := Lock(address)
    return bytes.Compare(out.Ripemd160Hash, b) == 0
}

//解码地址相当于锁定用户
func Lock(address string) []byte {
    publicKeyHash := Wallet.Base58Decode([]byte(address))
    hash160 := publicKeyHash[1 : len(publicKeyHash)-Wallet.AddressChecksumLen]
    return hash160
}

//创建output对象
func NewTxOutput(value int64,addres string) *TxOutput {
    txOutput := new(TxOutput)
    lock := Lock(addres)
    txOutput.Value=value
    txOutput.Ripemd160Hash=lock
    return txOutput
}