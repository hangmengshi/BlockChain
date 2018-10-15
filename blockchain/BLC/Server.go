package BLC

import (
    "bytes"
    "fmt"
    "io"
    "io/ioutil"
    "net"
    "log"
)

const PROTOCOL = "tcp"

const VERSION = "version"

const NODE_VERSION = 1

// 代表当前的区块版本信息(决定是否需要进行同步 )
type Version struct {
    Version    int    // 版本
    BestHeigth int    // 当前节点区块的高度
    AddrFrom   string // 当前节点的地址
}

// 3000作为主节点地址
var knowNodes = []string{"localhost:3000"}
// 服务处理文件
var nodeAddress string // 节点地址
// 启动服务器
func StartServer(nodeID string) {
    nodeAddress = fmt.Sprintf("localhost:%s", nodeID) // 服务节点地址
    fmt.Printf("服务节点 [%s] 启动...\n", nodeAddress)
    // 监听节点
    listen, err := net.Listen(PROTOCOL, nodeAddress)
    if nil != err {
        log.Panicf("listen address of %s failed! %v\n", nodeAddress, err)
    }
    defer listen.Close()
    bc := BlockchainObject(nodeID)
    if nodeAddress != knowNodes[0] {
        SendVersion(knowNodes[0], bc)
    }

    // 主节点接收请求
    for {
        conn, err := listen.Accept()
        if nil != err {
	log.Panicf("connect to node failed! %v\n", err)
        }
        request, err := ioutil.ReadAll(conn)
        if nil != err {
	log.Panicf("Receive a Message failed! %v\n", err)
        }
        cmd := bytesToCommand(request[:12])
        fmt.Printf("Receive a Message : %s\n", cmd)
    }
}

//"客户端(节点)"向服务器发送请求
func SendMessage(to string, msg []byte) {
    fmt.Println("向服务器发送请求...")
    conn, err := net.Dial(PROTOCOL, to)
    if nil != err {
        log.Panicf("connect to server failed! %v", conn)
    }
    defer conn.Close()
    // 要发送的数据添加到请求中
    io.Copy(conn, bytes.NewReader(msg))
    if nil != err {
        log.Panicf("add the data failed! %v\n", err)
    }
}

// 数据同步的函数
func SendVersion(toAddress string, bc *BlockChain) {
    // 在比特币中，消息是底层的比特序列，前12个字节指定命令名(verion)
    // 后面的字节包含的是gob编码过的消息结构
    bestHeigth := 1
    data := gobEncode(Version{NODE_VERSION, bestHeigth, nodeAddress})
    request := append(commandToBytes(VERSION), data...)
    SendMessage(toAddress, request)
}
