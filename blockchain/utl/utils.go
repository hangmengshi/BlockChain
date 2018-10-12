package utl

import (
    "bytes"
    "encoding/binary"
    "log"
)

//类型转换
func Int64ToSting(data int64) []byte {
    buffer := new(bytes.Buffer)
    err := binary.Write(buffer, binary.BigEndian, data)
    if err!=nil{
        log.Panic(err)
    }
    return buffer.Bytes()
}
