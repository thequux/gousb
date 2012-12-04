package usb

// #cgo CFLAGS: -I/usr/include/libusb-1.0
// #cgo LDFLAGS: -lusb-1.0
// #include <libusb.h>
// #include <malloc.h>
import "C"
import (
	"io"
	"os"
	"unsafe"
)

type Endpoint interface {
	GetDescriptor() EndpointDescriptor
	Readable() bool
	Writable() bool
	io.Reader
	io.Writer
}
/*
type EndpointReader interface {
	Endpoint
	io.Reader
}
type EndpointWriter interface {
	Endpoint
	io.Writer
}
*/
type EndpointHandle struct {
	handle     *DeviceHandle
	descriptor *EndpointDescriptor
	readable   bool
	ep         byte // endpoint number
	transfer   func(ep *EndpointHandle, p []byte, for_read bool) (n int, err os.Error)
}

func (ep *EndpointHandle) Write(p []byte) (n int, err os.Error) {
	if ep.readable {
		return 0, os.EBADF
	}
	n, err = ep.transfer(ep, p, false)
	if err == nil && n < len(p) {
		err = os.EAGAIN
	}
	return
}


func (ep *EndpointHandle) Read(p []byte) (n int, err os.Error) {
	if !ep.readable {
		return 0, os.EBADF
	}
	return ep.transfer(ep, p, true)
}

func (ep *EndpointHandle) Readable() bool {
	return ep.readable
}
func (ep *EndpointHandle) Writable() bool {
	return !ep.readable
}
func (ep *EndpointHandle) GetDescriptor() EndpointDescriptor {
	return *ep.descriptor
}

type TwoWayStream struct {
	R io.Reader
	W io.Writer
}

func (s TwoWayStream) Read(p []byte) (n int, err os.Error) {
	return s.R.Read(p)
}
func (s TwoWayStream) Write(p []byte) (n int, err os.Error) {
	return s.W.Write(p)
}

func interruptTransfer(ep *EndpointHandle, p []byte, _ bool) (n int, err os.Error) {
	var transferred C.int
	err0 := returnUsbError(C.libusb_interrupt_transfer(
		ep.handle.handle,
		C.uchar(ep.ep),
		(*C.uchar)(unsafe.Pointer(&p[0])),
		C.int(len(p)),
		&transferred,
		0))
	if err0 != nil {
		err = err0
	}
	return int(transferred), err
}

func bulkTransfer(ep *EndpointHandle, p []byte, _ bool) (n int, err os.Error) {
	var transferred C.int
	err0 := returnUsbError(C.libusb_bulk_transfer(
		ep.handle.handle,
		C.uchar(ep.ep),
		(*C.uchar)(unsafe.Pointer(&p[0])),
		C.int(len(p)),
		&transferred,
		0))
	if err0 != nil {
		err = err0
	}

	return int(transferred), err
}

