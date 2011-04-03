package usb

// #cgo CFLAGS: -I/usr/include/libusb-1.0
// #cgo LDFLAGS: -lusb-1.0
// #include <libusb.h>
// #include <stdio.h>
import "C"
import "reflect"
import "unsafe"


type UsbError struct{
	Text string
}

/////////////////// Basic types
type Context struct {
	initialized bool
	ctx *C.struct_libusb_context
}

var DefaultContext *Context

var (
	UsbSuccess = &UsbError{"SUCCESS"}
	UsbErrorIO = &UsbError{"EIO"}
	UsbErrorInvalidParam = &UsbError{"EINVALID"}
	UsbErrorAccess = &UsbError{"EACCESS"}
	UsbErrorNoDevice = &UsbError{"ENODEV"}
	UsbErrorNotFound = &UsbError{"ENOENT"}
	UsbErrorBusy = &UsbError{"EBUSY"}
	UsbErrorTimeout = &UsbError{"ETIMEDOUT"}
	UsbErrorOverflow = &UsbError{"EOVERFLOW"}
	UsbErrorPipe = &UsbError{"EPIPE"}
	UsbErrorInterrupted = &UsbError{"EINTR"}
	UsbErrorNoMem = &UsbError{"ENOMEM"}
	UsbErrorNotSupported = &UsbError{"ENOTSUP"}
	UsbErrorMisc = &UsbError{"EIEIO"} // Old McDonald had a flash drive...
)

const (
	DEBUG_SILENT = iota
	DEBUG_ERROR
	DEBUG_WARN
	DEBUG_INFO
)

var UsbErrorMap = map[int]*UsbError{
	-0: UsbSuccess,
	-1: UsbErrorIO,
	-2: UsbErrorInvalidParam,
	-3: UsbErrorAccess,
	-4: UsbErrorNoDevice,
	-5: UsbErrorNotFound,
	-6: UsbErrorBusy,
	-7: UsbErrorTimeout,
	-8: UsbErrorOverflow,
	-9: UsbErrorPipe,
	-10: UsbErrorInterrupted,
	-11: UsbErrorNoMem,
	-12: UsbErrorNotSupported,
	-99: UsbErrorMisc,
}

func decodeUsbError(errno C.int) (int,*UsbError) {
	if errno >= 0 {
		return int(errno), nil
	}
	err, ok := UsbErrorMap[int(errno)]
	if !ok {
		err = UsbErrorMisc
	} 
	return int(errno), err
}

func returnUsbError(errno C.int) *UsbError {
	_, err := decodeUsbError(errno)
	return err
}

func (err *UsbError) String() string {
	return err.Text
}

//////////////////////// Basic lifecycle support...

// Automatically called when necessary
func (ctx *Context) doinit() *UsbError {
	if !ctx.initialized {
		_, err := decodeUsbError(C.libusb_init(&ctx.ctx))
		if err == nil {
			ctx.initialized = true
		}
		return err
	}
	return nil
}

// Can be called to shut down. Probably not required
func (ctx *Context) destroy() {
	if ctx.initialized {
		ctx.initialized = false
		C.libusb_exit(ctx.ctx)
		ctx.ctx = nil
	}
}

func (ctx *Context) SetDebug(level int) {
	ctx.doinit()
	C.libusb_set_debug(ctx.ctx, C.int(level))
}


func init() {
	_, err := decodeUsbError(C.libusb_init(nil))
	if err != nil {
		panic(err)
	}
	DefaultContext = &Context{initialized: true}
}
//////////////////////// DEVICE SUPPORT
// memory management
type Device struct {
	ctx *Context
	device *C.struct_libusb_device
}

type DeviceHandle struct {
	ctx *Context
	handle *C.struct_libusb_device_handle
	interfaces map[int]*Interface
}

func (ctx *Context) wrapDevice(dev *C.struct_libusb_device) *Device {
	C.libusb_ref_device(dev)
	return &Device{ctx, dev}
}
func (dev *Device) destroy() {
	C.libusb_unref_device(dev.device)
	dev.device = nil
	dev.ctx = nil
}

func (handle *DeviceHandle) destroy() {
	if handle.handle != nil {
		C.libusb_close(handle.handle)
		handle.handle = nil
	}
	handle.ctx = nil
}

func (ctx *Context) GetDeviceList() (dev []*Device, err *UsbError) {
	var (
		baseptr **C.struct_libusb_device
		devlist []*C.struct_libusb_device
	)
	count, err := decodeUsbError(C.int(C.libusb_get_device_list(ctx.ctx, &baseptr)))
	if err != nil {
		dev = nil
		return
	}

	hdr := &reflect.SliceHeader{Data: uintptr(unsafe.Pointer(baseptr)), Len: count, Cap: count}
	devlist = unsafe.Unreflect(unsafe.Typeof(devlist), unsafe.Pointer(&hdr)).([]*C.struct_libusb_device)
	dev = make([]*Device, count)
	for i := 0; i < count; i++ {
		dev[i] = ctx.wrapDevice(devlist[i])
	}
	devlist = nil
	C.libusb_free_device_list(baseptr, 1)
	return dev, nil
}

func (dev *Device) GetDeviceAddress() (bus,addr int) {
	bus = int(C.libusb_get_bus_number(dev.device))
	addr = int(C.libusb_get_device_address(dev.device))
	return
}
	
func (dev *Device) GetMaxPacketSize(endpoint int) (int,*UsbError) {
	sz, err := decodeUsbError(C.libusb_get_max_packet_size(dev.device, C.uchar(endpoint)))
	return sz, err
}

func (dev *Device) GetMaxIsoPacketSize(endpoint int) (sz int, err *UsbError) {
	sz, err = decodeUsbError(C.libusb_get_max_iso_packet_size(dev.device, C.uchar(endpoint)))
	return
}
	

func (dev *Device) Open() (handle *DeviceHandle, err *UsbError) {
	handle = &DeviceHandle{dev.ctx, nil, make(map[int]*Interface, 0)}
	_, err = decodeUsbError(C.libusb_open(dev.device, &handle.handle))
	if err != nil {
		handle = nil
	}
	return
}

func (ctx *Context) Open(vendor, product int) (*DeviceHandle,*UsbError) {
	
	handle := &DeviceHandle{ctx, nil, make(map[int]*Interface, 0)}
	dev := C.libusb_open_device_with_vid_pid(ctx.ctx, C.uint16_t(vendor), C.uint16_t(product))

	if dev == nil {
		return nil, UsbErrorMisc
	}
	handle.handle = dev
	return handle, nil
}

func (h *DeviceHandle) GetDevice() *Device {
	return h.ctx.wrapDevice(C.libusb_get_device(h.handle))
}

func (h *DeviceHandle) GetConfiguration() (int, *UsbError) {
	var res C.int
	if err := returnUsbError(C.libusb_get_configuration(h.handle, &res)); err != nil {
		return 0, err
	}
	return int(res), nil
}

func (h *DeviceHandle) SetConfiguration(config int) *UsbError {
	return returnUsbError(C.libusb_set_configuration(h.handle, C.int(config)))
}

//////////////////////// Interfaces
type Interface struct {
	handle *DeviceHandle
	num C.int
	claimed int
}

func (h *DeviceHandle) GetInterface(iface_no int) *Interface {
	if iface, ok := h.interfaces[iface_no]; ok {
		return iface
	}
	iface := &Interface{handle: h, num: C.int(iface_no)}
	h.interfaces[iface_no] = iface
	return iface
}

func (i *Interface) Claim() *UsbError {
	if err := returnUsbError(C.libusb_claim_interface(i.handle.handle, i.num)); err != nil {
		return err
	}
	i.claimed++
	return nil
}

func (i *Interface) Release() *UsbError {
	i.claimed--
	if i.claimed <= 0 {
		return returnUsbError(C.libusb_release_interface(i.handle.handle, i.num))
	}
	return nil
}

func (i *Interface) SetAlternate(alt int)  *UsbError {
	return returnUsbError(C.libusb_set_interface_alt_setting(i.handle.handle, i.num, C.int(alt)))
}

func (i *Interface) IsKernelDriverActive() (bool, *UsbError) {
	v, err := decodeUsbError(C.libusb_kernel_driver_active(i.handle.handle, i.num))
	if err != nil {
		return false, err
	}
	return (v == 1), nil
}

func (i *Interface) AttachKernelDriver() *UsbError {
	return returnUsbError(C.libusb_attach_kernel_driver(i.handle.handle, i.num))
}

func (i *Interface) DetachKernelDriver() *UsbError {
	return returnUsbError(C.libusb_detach_kernel_driver(i.handle.handle, i.num))
}

func (h *DeviceHandle) ClearHalt(endpoint int) *UsbError {
	return returnUsbError(C.libusb_clear_halt(h.handle, C.uchar(endpoint)))
}

func (h *DeviceHandle) Reset() *UsbError {
	return returnUsbError(C.libusb_reset_device(h.handle))
}



///////////////////////
