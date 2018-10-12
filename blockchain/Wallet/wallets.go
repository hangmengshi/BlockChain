package Wallet

import (
    "fmt"
    "os"
    "io/ioutil"
    "log"
    "encoding/gob"
    "crypto/elliptic"
    "bytes"
)

//钱包集合的存储文件
const walletFile = "Wallets_%s.dat"

//钱包的集合结构
type Wallets struct {
    Wallets map[string]*Wallet
}

//初始化一个钱包集合
func NewWallets(nodeID string) (*Wallets,error) {
    file := fmt.Sprintf(walletFile, nodeID)
    //判断文件是否存在
   if _, err := os.Stat(file);os.IsNotExist(err){
       wallets := &Wallets{}
       wallets.Wallets = make(map[string]*Wallet)
       return wallets,err
   }
    //文件存在，读取内容
    readFile, err := ioutil.ReadFile(file)
    if err!=nil {
        log.Printf("get file content failed! %v\n",err)
    }
    var wallets  Wallets
    gob.Register(elliptic.P256())
    decoder := gob.NewDecoder(bytes.NewReader(readFile))
    err = decoder.Decode(&wallets)
    if err!=nil{
        log.Printf("decode readFile failed %v\n",err)
    }
    return &wallets,nil
}

//创建新的钱包，并且将其添加到集合
func (wallets *Wallets) CreateWallet() {
    wallet := NewWallet()
    wallets.Wallets[string(wallet.GetAddress())] = wallet
}
