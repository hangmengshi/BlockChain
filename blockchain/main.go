package main

import "blockchain/BLC"

func main() {
    //wallet := Wallet.NewWallet()
    //address := wallet.GetAddress()
    //fmt.Printf("address : %s\n", address)
    //fmt.Printf("validation of address %s is %v\n", address, Wallet.IsValidForAddress([]byte(address)))
    BLC.CreateBloxkChainWithGenesisBlock("aaa")

}