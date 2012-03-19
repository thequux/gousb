package hex

import (
	"fmt"
	"os"
	"io"
	"flag"
	"encoding/line"
	"encoding/hex"
)

// http://microsym.com/editor/assets/intelhex.pdf
// RECLEN | LOAD_OFFSET | RECTYP | DATA | CHKSUM
//  1(n)         2          1       n       1

type IntelHEX struct {
	stream <-chan *Record
}

func (i *IntelHEX) Next() *Record {
	return <-i.stream
}

func (i *IntelHEX) Progress() (cur, total int) {
	return 0, 0
}

func main() {
	flag.Parse()
	var filename string
	if flag.NArg() > 0 {
		filename = flag.Arg(0)
	} else {
		filename = "intelHEXexample.txt"
	}
	test(filename)
	return
}

func test(filename string) {
	file, err := os.Open(filename, os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	intelHEX := newIntelHEX(file)
	for a := range intelHEX.stream {
		if a.Data == nil {
			break
		}
		fmt.Printf("%d %s <-\n", a.Address, string(a.Data))
	}
}

const (
	IH_DATA = iota
	IH_EOF
	IH_ESAR
	IH_SSAR
	IH_ELAR
	IH_SLAR
)

func newIntelHEX(r io.Reader) *IntelHEX {
	out := make(chan *Record)
	between := make(chan []byte)
	reader := line.NewReader(r, 64)
	decodeHex := func(b []byte, n int) (int, []byte) {
		res := 0
		for _, ch := range b[:n] {
			digit := int(0)
			switch {
			case '0' <= ch && ch <= '9':
				digit = int(ch - '0')
			case 'A' <= ch && ch <= 'F':
				ch |= 0x20
				fallthrough
			case 'a' <= ch && ch <= 'f':
				digit = int(ch - 'a' + 10)
			default:
				return res, nil
			}
			res = res<<4 + digit
		}
		return res, b[n:]
	}
	byte_addition := func(b []byte) (res int) {
		res = 0
		var c int
		for len(b) > 0 {
			c, b = decodeHex(b, 2)
			res += c
		}
		return
	}

	go func() {
		for {
			lp, is_prefix, err := reader.ReadLine()
			//fmt.Printf("lp is %s (%d)\n", lp, len(lp))
			if is_prefix {
				for is_prefix {
					var extra []byte
					cp := make([]byte, len(lp))
					copy(cp, lp)
					extra, is_prefix, err = reader.ReadLine()
					//fmt.Printf("extra \"%s\"\n", extra)
					if err != nil {
						goto cleanup // umad?
					}
					lp = append(cp, extra...)
					//fmt.Printf("extended: \"%s\"\n", lp)
				} 
			}
			if ! is_prefix {
				//fmt.Printf("sending \"%s\"\n", lp)
				between <- []byte(string(lp)) // truncate any left over space from []byte
			} 
		cleanup:
			if err != nil {
				if err == os.EOF {
					fmt.Printf("a exiting...\n")
					close(between)
					return
				}
				panic(err)
			}
		}
	}()

	go func() {
		addrMod := 0
		wordsz := 2
		for str := range between {
			if str == nil {
				fmt.Printf("b exiting...\n")
				out <- &Record{
					Address: 0,
					Data:    nil,
				}
				return
			}
			if str[0] != ':' {
				panic("not even the \":\"?\n")
			}
			//fmt.Printf("b recvd \"%s\"\n", str)
			os.Stdout.Sync()

			str = str[1:]
			//fmt.Printf("checksuming \"%s\"\n", str[:len(str)-2])
			xsum := 0x100 - (byte_addition(str[:len(str)-2]) & 0xFF)
			checksum, _ := decodeHex(str[len(str)-2:], 2)
			// checksum, _ := decodeHex(str[bytecnt*wordsz:], 2)
			if xsum != checksum {
				panic("checksum failed")
			}

			bytecnt, str := decodeHex(str, 2)
			addr, str := decodeHex(str, 4)
			rectyp, str := decodeHex(str, 2)
			// fmt.Printf("{%d %d %d (%x v %x)}\n", bytecnt, addr, rectyp, xsum, checksum)

			switch rectyp {
			case IH_DATA:
				data, err := hex.DecodeString(string(str[:bytecnt*wordsz]))
				if err != nil {
					panic(err)
				}

				addr += int(addrMod)
				out <- &Record{
					Address: int(addr),
					Data:    data,
				}
			case IH_ESAR: // switch wordsz to 4
				wordsz = 4
				mod, _ := decodeHex(str, bytecnt*wordsz)
				addrMod = mod << 4
			case IH_ELAR:
				mod, _ := decodeHex(str, bytecnt*wordsz)
				addrMod = mod << 16
			case IH_SSAR:
				continue // do we care about CS:IP registers?
			case IH_SLAR:
				continue // " " EIP?
			case IH_EOF:
				out <- &Record{
					Address: 0,
					Data:    nil,
				}
				return
			}


		}
	}()

	return &IntelHEX{
		stream: out,
	}
}
