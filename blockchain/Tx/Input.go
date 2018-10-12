package Tx

import (
    "blockchain/Wallet"
    "bytes"
)

//交易输入
type TxInput struct {
    //交易哈希（不是当前交易的哈希，而是该输入所引用的交易的哈希）
    TxHash []byte
    //引用的上一笔交易的Output索引
    Vout int
    //数字签名
    Signature []byte
    //公钥
    PublicKey []byte
}

//公钥哈希
func (in *TxInput) UnlockRipemd160Hash(ripHash []byte) bool {
    //获取input的ripemd160哈希
    hash := Wallet.Ripemd160Hash(in.PublicKey)
    return bytes.Compare(hash, ripHash) == 0
}
