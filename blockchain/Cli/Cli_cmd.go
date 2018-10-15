package Cli

// 查询余额
import (
    "fmt"
    "os"
    "blockchain/BLC"
    "blockchain/Tx"
    "blockchain/Wallet"
)

// 创建区块链
func (cli *CLI) createBlockchainWithGenesis(address string, nodeID string) {
    blockchain := BLC.CreateBloxkChainWithGenesisBlock(address, nodeID)
    defer blockchain.Db.Close()

    // 设置utxoSet操作
    utxoSet := &Tx.UTXOSet{blockchain}
    utxoSet.ResetUTXOSet() // 重置数据库，主要是更新UTXO表
}

// 创建钱包集合
func (cli *CLI) CreateWallets(nodeID string) {
    fmt.Printf("nodeID : %v\n", nodeID)
    // 创建一个集合对象
    wallets, _ := Wallet.NewWallets(nodeID)
    wallets.CreateWallet(nodeID)
    fmt.Printf("wallets : %v\n", wallets)
}

func (cli *CLI) getAddressLists(nodeID string) {
    fmt.Println("打印所有钱包地址...")
    wallets, _ := Wallet.NewWallets(nodeID)
    for address, _ := range wallets.Wallets {
        fmt.Printf("address : [%s]\n", address)
    }
}

func (cli *CLI) getBalance(from string, nodeID string) {
    // 获取指定地址的余额
    blockchain := BLC.BlockchainObject(nodeID)
    defer blockchain.Db.Close()
    utxoSet := &Tx.UTXOSet{blockchain}
    amount := utxoSet.GetBalance(from)
    fmt.Printf("\t地址: %s的余额为:%d\n", from, amount)
}

// 输出区块链信息
func (cli *CLI) printchain(nodeID string) {
    if BLC.DbExists(nodeID) == false {
        fmt.Println("数据库不存在...")
        os.Exit(1)
    }
    blockchain := BLC.BlockchainObject(nodeID) // 获取区块链对象
    defer blockchain.Db.Close()
    blockchain.PrintChain()
}

// 发送交易
func (cli *CLI) send(from []string, to []string, amount []string, nodeID string) {
    // 检测数据库
    if BLC.DbExists(nodeID) == false {
        fmt.Println("数据库不存在...")
        os.Exit(1)
    }
    blockchain := BLC.BlockchainObject(nodeID) // 获取区块链对象
    defer blockchain.Db.Close()
    blockchain.MineNewBlock(from, to, amount, nodeID)

    utxoSet := &Tx.UTXOSet{blockchain}
    utxoSet.Update()
}

// 实现启动服务的功能

func (cli *CLI) startNode(nodeID string) {
    BLC.StartServer(nodeID)
}
