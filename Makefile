GOROOT=/usr/local/google/home/thequux/go
include $(GOROOT)/src/Make.inc

DEPS=\
	usb

export GOFLAGS=-N
export GOLDFLAGS=-e

TARG=usb_t
GOFILES=\
	main.go

include $(GOROOT)/src/Make.cmd
