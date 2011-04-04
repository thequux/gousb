package main

import (
	"usb"
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

	_ = usb.ConfigDescriptor{
		BLength:             0x9,
		BDescriptorType:     2,
		WTotalLength:        0x29,
		BConfigurationValue: 0x1,
		IConfiguration:      0x0,
		BmAttributes:        0x80,
		MaxPower:            0x32,
		Interfaces: [][]usb.InterfaceDescriptor{
			[]usb.InterfaceDescriptor{
				usb.InterfaceDescriptor{
					BLength:            0x9,
					BDescriptorType:    4,
					BInterfaceNumber:   0x0,
					BAlternateSetting:  0x0,
					BInterfaceClass:    3,
					BInterfaceSubClass: 0x0,
					BInterfaceProtocol: 0x0,
					IInterface:         0x0,
					Endpoints: []usb.EndpointDescriptor{
						usb.EndpointDescriptor{
							BLength:          0x7,
							BDescriptorType:  5,
							BEndpointAddress: 0x81,
							BmAttributes:     0x3,
							WMaxPacketSize:   0x40,
							BInterval:        0x1,
							BRefresh:         0x0,
							BSynchAddress:    0x0,
							Extra:            []byte{}},
						usb.EndpointDescriptor{
							BLength:          0x7,
							BDescriptorType:  5,
							BEndpointAddress: 0x1,
							BmAttributes:     0x3,
							WMaxPacketSize:   0x40,
							BInterval:        0x1,
							BRefresh:         0x0,
							BSynchAddress:    0x0,
							Extra:            []byte{}}},
					Extra: []byte{0x9,
						0x21,
						0x1,
						0x1,
						0x0,
						0x1,
						0x22,
						0x24,
						0x0}}}},
		Extra: []byte{}}
}
