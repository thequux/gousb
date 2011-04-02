package usb
// #cgo CFLAGS: -I/usr/include/libusb-1.0
// #cgo LDFLAGS: -lusb-1.0
// #include <libusb.h>
// #include <malloc.h>
import "C"
import "reflect"
import "unsafe"

type (
	DescriptorType int
	ClassCode      int
)

const (
	DT_DEVICE DescriptorType = iota + 1
	DT_CONFIG
	DT_STRING
	DT_INTERFACE
	DT_ENDPOINT

	// These don't actually show up in libusb, but are in the spec
	DT_DEVICE_QUALIFIER
	DT_OTHER_SPEED_CONFIG
	DT_INTERFACE_POWER

	// These are in the spec, but not libusb
	DT_HID          DescriptorType = 0x21
	DT_HID_REPORT   DescriptorType = 0x22
	DT_HID_PHYSICAL DescriptorType = 0x23
	DT_HUB          DescriptorType = 0x29
)

const (
	CLASS_PER_INTERFACE ClassCode = iota
	CLASS_AUDIO
	CLASS_COMM
	CLASS_HID
	CLASS_PTP
	CLASS_PRINTER
	CLASS_MASS_STORAGE
	CLASS_HUB
	CLASS_DATA
	CLASS_VENDOR ClassCode = 0xff
)

const (
	ISO_SYNC_TYPE_NONE     = iota << 2 //
	ISO_SYNC_TYPE_ASYNC                //
	ISO_SYNC_TYPE_ADAPTIVE             //
	ISO_SYNC_TYPE_SYNC                 //
	ISO_SYNC_TYPE_MASK     = 0x0C
)

const (
	ISO_USAGE_TYPE_DATA     = iota << 4 // 
	ISO_USAGE_TYPE_FEEDBACK             // 
	ISO_USAGE_TYPE_IMPLICIT             //
	ISO_USAGE_TYPE_MASK     = 0x30
)

const (
	TRANSFER_TYPE_CONTROL     = iota // Control endpoint
	TRANSFER_TYPE_ISOCHRONOUS        // Isochronous endpoint
	TRANSFER_TYPE_BULK               // Bulk endpoint
	TRANSFER_TYPE_INTERRUPT          // Interrupt endpoint
	TRANSFER_TYPE_MASK        = 0x03
)

type (
	DeviceDescriptor struct {
		BLength             byte
		BDescriptorType     DescriptorType
		BcdUSB              uint16
		BDeviceClass        ClassCode
		BDeviceSubClass     byte
		BDeviceProtocol     byte
		BMaxPacketSize0     byte
		IdVendor, IdProduct uint16
		BcdDevice           uint16
		IManufacturer       byte
		IProduct            byte
		ISerialNumber       byte
		BNumConfigurations  byte
	}

	ConfigDescriptor struct {
		BLength             byte
		BDescriptorType     DescriptorType
		WTotalLength        uint16
		BConfigurationValue byte
		IConfiguration      byte
		BmAttributes        byte
		MaxPower            byte
		Interfaces          [][]InterfaceDescriptor
		Extra               []byte
	}

	EndpointDescriptor struct {
		BLength          byte
		BDescriptorType  DescriptorType
		BEndpointAddress byte
		BmAttributes     byte
		WMaxPacketSize   uint16
		BInterval        byte
		BRefresh         byte
		BSynchAddress    byte
		Extra            []byte
	}

	InterfaceDescriptor struct {
		BLength            byte
		BDescriptorType    DescriptorType
		BInterfaceNumber   byte
		BAlternateSetting  byte
		BInterfaceClass    ClassCode
		BInterfaceSubClass byte
		BInterfaceProtocol byte
		IInterface         byte
		Endpoints          []EndpointDescriptor
		Extra              []byte
	}
)

// Usage:
// given: foo_array *Foo := native array of Foo
//        foo_len   int  := Number of Foos in foo_array
// 
// res := sliceify(foo_array, []Foo(nil), foo_len).([]Foo)
func sliceify(arg interface{}, sample interface{}, length int) interface{} {
	val := reflect.NewValue(arg).(*reflect.PtrValue)
	target := unsafe.Pointer(&reflect.SliceHeader{
		Data: val.Get(),
		Len:  length,
		Cap:  length,
	})
	return unsafe.Unreflect(reflect.Typeof(sample), target)
}

func cloneSlice(arg interface{}) interface{} {
	oldval := reflect.NewValue(arg).(*reflect.SliceValue)
	val := reflect.MakeSlice(oldval.Type().(*reflect.SliceType), oldval.Len(), oldval.Len())
	reflect.Copy(val, oldval)
	return val.Interface()
}

func parseDeviceDescriptor(desc *C.struct_libusb_device_descriptor) DeviceDescriptor {
	return DeviceDescriptor{
		BLength:            byte(desc.bLength),
		BDescriptorType:    DescriptorType(desc.bDescriptorType),
		BcdUSB:             uint16(desc.bcdUSB),
		BDeviceClass:       ClassCode(desc.bDeviceClass),
		BDeviceSubClass:    byte(desc.bDeviceSubClass),
		BDeviceProtocol:    byte(desc.bDeviceProtocol),
		BMaxPacketSize0:    byte(desc.bMaxPacketSize0),
		IdVendor:           uint16(desc.idVendor),
		IdProduct:          uint16(desc.idProduct),
		BcdDevice:          uint16(desc.bcdDevice),
		IManufacturer:      byte(desc.iManufacturer),
		IProduct:           byte(desc.iProduct),
		ISerialNumber:      byte(desc.iSerialNumber),
		BNumConfigurations: byte(desc.bNumConfigurations),
	}
}

func parseConfigDescriptor(desc *C.struct_libusb_config_descriptor) ConfigDescriptor {
	ret := ConfigDescriptor{
		BLength:             byte(desc.bLength),
		BDescriptorType:     DescriptorType(desc.bDescriptorType),
		WTotalLength:        uint16(desc.wTotalLength),
		BConfigurationValue: byte(desc.bConfigurationValue),
		IConfiguration:      byte(desc.iConfiguration),
		BmAttributes:        byte(desc.bmAttributes),
		MaxPower:            byte(desc.MaxPower),
		Interfaces:          make([][]InterfaceDescriptor, int(desc.bNumInterfaces)),
		Extra:               cloneSlice(sliceify(desc.extra, []byte{}, int(desc.extra_length))).([]byte),
	}

	iface_list := sliceify(desc._interface, []C.struct_libusb_interface(nil), int(desc.bNumInterfaces)).([]C.struct_libusb_interface)
	for i := 0; i < int(desc.bNumInterfaces); i++ {
		iface := iface_list[i]
		alts := sliceify(iface.altsetting, []C.struct_libusb_interface_descriptor{}, int(iface.num_altsetting)).([]C.struct_libusb_interface_descriptor)
		parsed := make([]InterfaceDescriptor, len(alts), len(alts))
		ret.Interfaces[i] = parsed
		for j := range alts {
			parsed[j] = parseInterfaceDescriptor(&alts[j])
		}
	}
	return ret
}

func parseEndpointDescriptor(desc *C.struct_libusb_endpoint_descriptor) EndpointDescriptor {
	return EndpointDescriptor{
		BLength:          byte(desc.bLength),
		BDescriptorType:  DescriptorType(desc.bDescriptorType),
		BEndpointAddress: byte(desc.bEndpointAddress),
		BmAttributes:     byte(desc.bmAttributes),
		WMaxPacketSize:   uint16(desc.wMaxPacketSize),
		BInterval:        byte(desc.bInterval),
		BRefresh:         byte(desc.bRefresh),
		BSynchAddress:    byte(desc.bSynchAddress),
		Extra:            cloneSlice(sliceify(desc.extra, []byte{}, int(desc.extra_length))).([]byte),
	}
}

func parseInterfaceDescriptor(desc *C.struct_libusb_interface_descriptor) InterfaceDescriptor {
	ret := InterfaceDescriptor{
		BLength:            byte(desc.bLength),
		BDescriptorType:    DescriptorType(desc.bDescriptorType),
		BInterfaceNumber:   byte(desc.bInterfaceNumber),
		BAlternateSetting:  byte(desc.bAlternateSetting),
		BInterfaceClass:    ClassCode(desc.bInterfaceClass),
		BInterfaceSubClass: byte(desc.bInterfaceSubClass),
		BInterfaceProtocol: byte(desc.bInterfaceProtocol),
		IInterface:         byte(desc.iInterface),
		Endpoints:          make([]EndpointDescriptor, int(desc.bNumEndpoints)),
		Extra:              cloneSlice(sliceify(desc.extra, []byte{}, int(desc.extra_length))).([]byte),
	}

	ep_list := sliceify(desc.endpoint, []C.struct_libusb_endpoint_descriptor(nil), int(desc.bNumEndpoints)).([]C.struct_libusb_endpoint_descriptor)
	for i := 0; i < int(desc.bNumEndpoints); i++ {
		ret.Endpoints[i] = parseEndpointDescriptor(&ep_list[i])
	}
	return ret
}

func (dev *Device) GetDeviceDescriptor() (DeviceDescriptor, *UsbError) {
	//var desc = (*C.struct_libusb_device_descriptor)(&C.devdesc)
	var desc C.struct_libusb_device_descriptor
	//println(desc)
	err := returnUsbError(C.libusb_get_device_descriptor(dev.device, &desc))
	if err != nil {
		return DeviceDescriptor{}, err
	}
	return parseDeviceDescriptor(&desc), err
}
