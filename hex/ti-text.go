package hex

import (
	"io"
	"os"
	"encoding/line"
	"strings"
)


// File format used mostly by MSP430 programming tools, but notable for its simplicity.
// Each line consists of either an atsign immediately followed by an address, or a set of hex pairs.
// The file is terminated by a line containing a single 'q'
//
// Address lines set the current address; hex pairs are written to
// sequential addresses starting with the current address.
//
// An example TiText file containing "GO" at address 0x200 follows:
//     @200
//     47 4f
//     q
type TiText struct{}

// Iplements half of HexFormat; q.v.
func (TiText) ReadHex(r io.Reader) (Hexfile, os.Error) {
	resp := RecordSequence{}

	// 16 bytes per line ought to be enough for anybody.
	line_reader := line.NewReader(r, 64)
	addr := 0

	for {
		line, is_prefix, err := line_reader.ReadLine()
		if line == nil && err == os.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		for is_prefix {
			var line_part []byte
			line_part, is_prefix, err = line_reader.ReadLine()
			if line == nil && err == os.EOF {
				break
			} else if err != nil {
				return nil, err
			}
			line = append(line, line_part...)
		}
		line_s := string(line)
		reader := strings.NewReader(line_s)
		switch line_s[0] {
		case '@':
			reader.ReadByte()
			if len(line) < 2 {
				return nil, StrError("Format error: short address")
			}
			addr, err = decodeInt(reader, -1)
			if err != nil {
				return nil, err
			}
			// TODO(thequux): Check for trailing junk
		case 'q':
			return resp, nil
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			fallthrough
		case 'a', 'b', 'c', 'd', 'e', 'f':
			fallthrough
		case 'A', 'B', 'C', 'D', 'E', 'F':
			buf := make([]byte, 0, len(line)/3)
			for {

				if n, err := decodeInt(reader, -1); err != nil {
					if err == os.EOF {
						break
					}
					return nil, err
				} else {
					buf = append(buf, byte(n))
				}
			}
			resp = append(resp, Record{addr, buf})
			addr += len(buf)
		default:
			return nil, StrError("Invalid format")
		}
	}
	return resp, nil
}
