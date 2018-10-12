package Tx

type UTXO struct {
    //UTXO的交易哈希
    TxHash []byte
    //UTXO所属交易中的索引
    Index int
    //Output
    Output *TxOutput
}

//所有的UTXO
type AllUTXO struct {
    UTXO []*UTXO
}
