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
func (wallets *Wallets) CreateWallet(nodeID string) {
    // 新建钱包对象
    wallet := NewWallet()
    wallets.Wallets[string(wallet.GetAddress())] = wallet
    // 把钱包存储到文件中
    wallets.SaveWallets(nodeID)
}
// 持久化钱包信息（写入文件）
func (w *Wallets) SaveWallets(nodeID string) {
    var content bytes.Buffer
    // 注册
    gob.Register(elliptic.P256())
    encoder := gob.NewEncoder(&content)
    // 序列化钱包数据
    err := encoder.Encode(&w)
    if nil != err {
        log.Panicf("encode the struct of wallets failed! %v\n", err)
    }
    // 清空文件再云存储(此处只保存了一条数据，但该条数据会存储到目前为止所有地址的集合)
    walletFile := fmt.Sprintf(walletFile, nodeID)
    err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
    if nil != err {
        log.Panicf("write the content of wallets to file [%s] failed! %v\n", walletFile, err)
    }
}

