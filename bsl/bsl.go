package main

import (
	"gopkg.thequux.com/usb"
	"log"
)

type Bsl struct {
	usb.TwoWayStream
}


var DefaultPassword = [32]byte{
	0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff}

func (bsl Bsl) sendCommand(buf []byte) []byte {
	if len(buf) >= 256 {
		panic("Buf too long")
	}
	packet := append([]byte{0x3f, byte(len(buf))}, buf...)
	_, err := bsl.Write(packet)
	if err != nil {
		log.Fatal("Writing command: ", err)
	}
	resp := make([]byte, 64)
	n, err := bsl.Read(resp)
	if n < 64 && err != nil {
		log.Fatal("Reading reply: ", err)
	}
	if n < 2 {
		log.Print("Short response")
	}
	if resp[0] != 0x3f {
		log.Print("Malformed packet")
	}
	return resp[2 : 2+resp[1]]
}

func (bsl Bsl) RxPassword(password [32]byte) []byte {
	msg := make([]byte, 33)
	copy(msg[1:], password[:])
	msg[0] = 0x11
	return bsl.sendCommand(msg)
}


func main() {
	ctx := new(usb.Context)
	ctx.SetDebug(3)
	dev, err := ctx.Open(0x2047, 0x0200)
	if err != nil {
		panic(err)
	}
	/*curr_cfg, err := dev.GetConfiguration()
	if err != nil {
		panic(err)
	}*/
	desc, _ := dev.GetDevice().GetConfigDescriptor(0)
	if err := dev.SetConfiguration(-1); err != nil {
		log.Print(err)
	}
	if err := dev.SetConfiguration(desc.BConfigurationValue); err != nil {
		log.Print(err)
	}
	idesc := desc.Interfaces[0][0]
	iface := dev.GetInterface(idesc.BInterfaceNumber)
	if active, _ := iface.IsKernelDriverActive(); active {
		if err := iface.DetachKernelDriver(); err != nil {
			panic(err)
		}
	}

	if err := iface.Claim(); err != nil {
		panic(err)
	}
	defer iface.Release()

	stream := Bsl{}
	for _, ep_desc := range idesc.Endpoints {
		ep, err := dev.OpenEndpoint(ep_desc)
		if err != nil {
			panic(err)
		}
		if ep.Readable() {
			stream.R = ep
		} else if ep.Writable() {
			stream.W = ep
		}
		ep.ClearHalt()
	}
	log.Print(stream.RxPassword(DefaultPassword))

}
