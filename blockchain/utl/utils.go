package utl

import (
    "bytes"
    "encoding/binary"
    "log"
    "encoding/json"
    "encoding/gob"
    "fmt"
)

//类型转换
func Int64ToSting(data int64) []byte {
    buffer := new(bytes.Buffer)
    err := binary.Write(buffer, binary.BigEndian, data)
    if err != nil {
        log.Panic(err)
    }
    return buffer.Bytes()
}

// int64转换成字节数组
func IntToHex(data int64) []byte {
    buffer := new(bytes.Buffer) // 新建一个buffer
    err := binary.Write(buffer, binary.BigEndian, data)
    if nil != err {
        log.Panicf("int to []byte failed! %v\n", err)
    }
    return buffer.Bytes()
}

func JSONToArray(jsonString string) []string {
    var strArr []string
    if err := json.Unmarshal([]byte(jsonString), &strArr); err != nil {
        log.Panicf("json to []string failed! %v\n", err)
    }
    return strArr
}

// 将结构体序列化为字节数组
func GobEncode(data interface{}) []byte {
    var buff bytes.Buffer
    enc := gob.NewEncoder(&buff)
    err := enc.Encode(data)
    if nil != err {
        log.Panicf("encode the data failed! %v\n", err)
    }
    return buff.Bytes()
}


// 将命令转C为字节数组
// 指令长度最长为12位
func CommandToBytes(command string) []byte {
    var bytes [12]byte // 命令长度
    for i, c := range command {
        bytes[i] = byte(c) // 转换
    }
    return bytes[:]
}

// 将字节数组转成cmd
func BytesToCommand(bytes []byte) string  {
    var command []byte // 接收命令
    for _, b := range bytes {
        if b != 0x0 {
            command = append(command, b)
        }
    }
    return fmt.Sprintf("%s", command)
}