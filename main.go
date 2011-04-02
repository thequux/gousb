package main

import (
	"usb"
	"fmt"
)

func main() {
	ctx := new(usb.Context)
	ctx.SetDebug(3)
	dev, err := ctx.Open(0x18d1,  0x4e22)
	if err != nil {
		panic(err)
	}
	desc, _ := dev.GetDevice().GetDeviceDescriptor()
	fmt.Printf("%#v\n", desc)
	println(ctx, dev)
}