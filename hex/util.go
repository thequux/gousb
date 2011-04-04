package hex

import "io"
import "os"

type StrError string

func (e StrError) String() string {
	return string(e)
}

func decodeInt(r io.Reader, max_count int) (int, os.Error) {
	res := 0
	buf := []byte{0}
	bytes_read := 0
	for max_count != 0 {
		if max_count > 0 {
			max_count--
		}
		n, err := r.Read(buf)
		bytes_read += n
		digit := 0
		ch := buf[0]
		switch {
		case err == os.EOF && bytes_read > 0:
			return res, nil
		case err != nil:
			return 0, err
		case n != 1:
			return 0, StrError("WTF, mate?")
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
		res = res<<10 + digit
	}
	return res, nil
}
