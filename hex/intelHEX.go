package hex

import (
	"fmt"
	"os"
	"encoding/hex"
	"strings"
	"log"
)

// 

//type IntelHEXfileIterator
// http://microsym.com/editor/assets/intelhex.pdf

func assert (err os.Error) {
	if err != nil {
		panic(err)
	}
}

func assertWhere (err os.Error, where string) {
	if err != nil {
		log.Panic(where, err)
	}
}

func main() {
	test ("intelHEXexample.txt")
	return
}

func test(filename string) {
	for _, line := range intelHEXfileUnpack(filename) {
		fmt.Printf ("str %s\n", line)
		fmt.Printf ("len %d\n", len (line))
		IntelHEXreadline ([]byte(line))
	}
	return
}

func intelHEXfileUnpack (filename string) []string {
	file, err := os.Open (filename, os.O_RDONLY, 0666)
	assert(err)
	defer file.Close()
	fileInfo, err := file.Stat()
	assert(err)
	buffer := make([]byte, fileInfo.Size)
	amt, err := file.Read (buffer)
	assert (err)
	ls := strings.Split(string(buffer[:amt]), ":", -1)
	return ls[1:len(ls)-1]
}

func popHexB (p []byte, len uint) ([]byte, []byte) {
	st, err := hex.DecodeString(string(p[:len]))
	assertWhere (err, "popHexB")
	return st, p[len:]
}

func popHexN (p []byte, ln uint) (int, []byte) {
	st, err := hex.DecodeString(string(p[:ln]))
	assertWhere(err, "popHexN: hex.DecodeString")
	return int(st[0]), p[ln:]
}

/* RECLEN | LOAD_OFFSET | RECTYP | DATA | CHKSUM
    1(n)         2          1       n       1 */
func IntelHEXreadline (p []byte) []byte {
	RECLEN, p := popHexN(p, 2)
	LOAD_OFFSET, p := popHexN(p, 4)
	RECTYP, p := popHexN(p, 2)
	DATA, p := popHexB(p, uint(2*RECLEN)) // 2 0x ASCII to one byte
	CHECKSUM, p := popHexN(p, 2) 
	fmt.Printf("%s: %d\n", "reclen",RECLEN)
	fmt.Printf("%s: %d\n", "load_offset", LOAD_OFFSET)
	fmt.Printf("%s: %d\n", "rectyp", RECTYP)
	fmt.Printf("%s: %s\n", "data", DATA)
	fmt.Printf("%s: %d\n", "checksum", CHECKSUM)
	return DATA
}
