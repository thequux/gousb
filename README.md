Gousb
=====

This provides some basic bindings to [libusb][]-1.0. It ir currently
rather incomplete, but progressing quickly.

Status
------

So far, only the bare minimum to implement lsusb is implemented.
Expect transfers to come sometime in the coming few weeks.

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

This was written by Dan Hirsch, aka TheQuux.
Flames/rants/suggestions/patches to <thequux@gmail.com>.
  
[libusb]: http://libusb.org
