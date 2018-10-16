package BLC

import (
    "math/big"
    "bytes"
    "crypto/sha256"
    "fmt"
    "blockchain/utl"
)

//工作量
type Pow struct {
    block  *Block   // 对指定的区块进行验证
    target *big.Int // 大数据存储(用于验证hash)
}

//定位常量,代表生成的hash前边0的个数
const targetBit = 20

//创建Pow对象

func NewPow(block *Block) *Pow {
    newInt := big.NewInt(1)
    lsh := newInt.Lsh(newInt, 256-targetBit)
    return &Pow{block, lsh}
}

// 开始工作量证明
func (p *Pow) Run() ([]byte, int64) {
    var nonce = 0       //碰撞系数
    var hashInt big.Int // 存储哈希转换之后生成的数据，最终和target数据进行比较
    for {
        nonce++
        hash := p.GetHash(int64(nonce))
        hashint := hashInt.SetBytes(hash)
        fmt.Println("hash : \r%x", hash)
        if p.target.Cmp(hashint) == 1 {
	fmt.Printf("\n碰撞次数: %d\n", nonce)
	return hash, int64(nonce)
        }
    }

}

//哈希
func (p *Pow) GetHash(n int64) []byte {
    block := p.block
    data := bytes.Join([][]byte{
        block.HashTransactuions(),
        block.PreHash,
        utl.Int64ToSting(block.TimeStamp),
        utl.Int64ToSting(block.Index),
        utl.Int64ToSting(n),
        utl.Int64ToSting(targetBit),
    }, []byte{})
    hash := sha256.Sum256(data)

    return hash[:]
}
