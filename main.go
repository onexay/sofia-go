package main

import (
	"fmt"
	"sofia-go/sofia"
)

func main() {
	device := sofia.NewDevice("192.168.31.177", "34567", "tcp4", "", "")

	device.Connect()

	device.Login()

	go sofia.ReadDevice(device)

	fmt.Scanln()
}
