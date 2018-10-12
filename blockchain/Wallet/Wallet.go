package Wallet

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "log"
    "crypto/sha256"
    "bytes"
    "golang.org/x/crypto/ripemd160"
)

//钱包相关
//版本
const Version = byte(0x00)

//checksum 长度
const AddressChecksumLen = 4

//钱包结构(存储键值对)
type Wallet struct {
    //1.私钥
    PrivateKey ecdsa.PrivateKey
    //2.公钥
    PublicKey []byte
}

//创建钱包
func NewWallet() *Wallet {
    privateKey, pubkey := newKeyPair()
    return &Wallet{PrivateKey: privateKey, PublicKey: pubkey}
}

//生成公钥私钥对
func newKeyPair() (ecdsa.PrivateKey, []byte) {
    curve := elliptic.P256()
    //椭圆加密
    priv, err := ecdsa.GenerateKey(curve, rand.Reader)
    if err != nil {
        log.Printf("ecdsa generate key failed %v\n", err)
    }
    pubKey := append(priv.PublicKey.X.Bytes(), priv.PublicKey.Y.Bytes()...)
    return *priv, pubKey
}

//对公钥进行双哈希
func Ripemd160Hash(puKey []byte) []byte {
    //sha256
    hash256 := sha256.New()
    hash256.Write(puKey)
    hash := hash256.Sum(nil)

    //ripemad160
    rmd160 := ripemd160.New()
    rmd160.Write(hash)
    return rmd160.Sum(nil)
}

//通过钱包获取地址
func (w *Wallet) GetAddress() []byte {
    //获取公钥哈希
    ripemd160Hash := Ripemd160Hash(w.PublicKey)
    //将生成的Version并加入到hash中
    VerHash := append([]byte{Version}, ripemd160Hash...)
    //生成校验和数据
    checkSum := CheckSum(VerHash)
    //拼接校验和
    bytes := append(VerHash, checkSum...)
    //调用base58Encode生成地址
    base58 := Base58Encode(bytes)
    return base58
}

//生成校验和
func CheckSum(payload []byte) []byte {
    first_hash := sha256.Sum256(payload)
    second_hash := sha256.Sum256(first_hash[:])
    return second_hash[:AddressChecksumLen]
}

// 判断地址有效性
func IsValidForAddress(address []byte) bool {
    // 1. 地址通过base58Decode进行解码
    version_pubkey_checksumBytes := Base58Decode(address) // 25位
    // 2. 拆开，进行校验和的校验
    checkSumBytes := version_pubkey_checksumBytes[len(version_pubkey_checksumBytes)-AddressChecksumLen:]
    version_ripemd160 := version_pubkey_checksumBytes[:len(version_pubkey_checksumBytes)-AddressChecksumLen]
    checkBytes := CheckSum(version_ripemd160)
    if bytes.Compare(checkSumBytes, checkBytes) == 0 {
        return true
    }
    return false
}
