include $(GOROOT)/src/Make.inc

DEPS=\
	usb\
	hex

export GOFLAGS=-N
export GOLDFLAGS=-e

TARG=bsl
GOFILES=\
	bsl.go

include $(GOROOT)/src/Make.cmd
