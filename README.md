Gousb
=====

This provides some basic bindings to [libusb][]-1.0. It is currently
rather incomplete, but progressing quickly.

Status
------

So far, lots of stuff has been written, but not much has been tested.
Interrupt transfers appear to work, but control/isochronous/bulk
transfers definitely don't.

It includes a [MSP430 bsl](http://focus.ti.com/lit/ug/slau319a/slau319a.pdf) client as a demo.

Installation
------------

To install, you have two options:

### Goinstall 

    $ goinstall github.com/thequux/gousb/usb
    
### Manually

    $ git clone git://github.com/thequux/gousb.git
    $ cd gousb/usb
    $ gomake install

Credits
-------

This was written by Dan Hirsch, aka TheQuux, with assistance from Glen Lenker.

Flames/rants/suggestions/patches to <thequux@gmail.com>, or <glen.lenker@gmail.com>.
  
[libusb]: http://libusb.org
