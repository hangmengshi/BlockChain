package Wallet

import (
    "math/big"
    "fmt"
    "bytes"
)

var base58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmopqrstuvwxyz")

//编码函数
func Base58Encode(input []byte) []byte {
    var result []byte
    x := big.NewInt(0).SetBytes(input)
    fmt.Println("x:%v\n", x)
    //设置一个base58基数（进制数）
    base := big.NewInt(int64(len(base58Alphabet)))
    zero := big.NewInt(0)
    mod := &big.Int{} //余数
    for x.Cmp(zero) != 0 {
        x.DivMod(x, base, mod) //求余
        //以余数为下表，取值
        result = append(result, base58Alphabet[mod.Int64()])
    }
    //反转切片
    Reverse(result)
    for b := range input {
        if b == 0x00 {
	result = append([]byte{base58Alphabet[0]}, result...)
        } else {
	break
        }
    }
    fmt.Printf("result : %s\n", result)
    return result
}

//解码函数
func Base58Decode(input []byte)[]byte  {
    result := big.NewInt(0)
    zeroBytes:=0
    for b:=range input{
        if b==0x00 {
            zeroBytes++
        }
    }
    fmt.Println(zeroBytes)
    data:=input[zeroBytes:]
    for _, b := range data {
        // 获取bytes数组中指定数字第一次出现的索引
        charIndex := bytes.IndexByte(base58Alphabet, b)
        result.Mul(result, big.NewInt(58))
        result.Add(result, big.NewInt(int64(charIndex)))
    }
    decoded := result.Bytes()
    decoded = append(bytes.Repeat([]byte{byte(0x00)}, zeroBytes), decoded...)
    return  decoded
}