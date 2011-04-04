package hex

type Record struct {
	Address int
	Data    []byte
}

// Exactly what it says on the tin. Iterates over records of a hex file
//
// Expected usage:
//    iter := file.Iterate()
//    for record := iter.Next(); record != nil; record = iter.Next() {
//            // do something with Record
//            log.Printf("%v/%v bytes processed", iter.Progress())
//    }
type HexfileIterator interface {
	// Return the next record, or nil if none
	Next() *Record
	// Get the current offset in the stream.
	//     total is the total number of bytes, or -1 if not known
	//     cur is the number of bytes seen, including the record most recently fetched by Next()
	Progress() (cur, total int)
}

type Hexfile interface {
	Iterate() HexfileIterator
}


type ReblockedIterator struct {
	stream   <-chan *Record
	max, pos int
}

// Reblocks a HexfileIterator. Record_size is the desired record size,
// and preserve_splits specifies whether to preserver record
// boundaries in the original stream. Chances are, you will get far
// fewer records if you don't try to preserve breaks.
func Reblock(iter HexfileIterator, record_size int, preserve_splits bool) HexfileIterator {
	pos, max := iter.Progress()
	stream := make(chan *Record)

	go func() {
		defer close(stream)
		buffer = make([]byte, record_size*2)
		addr := -1
		drain := func(flush bool) {
			for len(buffer) >= record_size {
				stream <- &Record{
					Address: addr,
					Data:    buffer[0:record_size],
				}
				buffer = buffer[record_size:]
				addr += record_size
			}
			if !flush {
				return
			}
			if len(buffer > 0) {
				stream <- &Record{
					Address: addr,
					Data:    buffer,
				}
				addr += len(buffer)
			}
			buffer = make([]byte, record_size*2)
		}
		for record := iter.Next(); iter != nil; record = iter.Next() {
			if addr+len(buffer) != record.Address {
				// drain the thing...
				drain(true)
				addr = record.Address
			}

			buffer := append(buffer, record.Data)
			// No need to completely flush it, unless we
			// want to preserve breaks in the original.
			drain(preserve_splits)
		}
		// No more records... flush completely.
		drain(true)
	}()

	return &ReblockedIterator{
		stream: stream,
		max:    max,
		pos:    pos,
	}
}

func (iter *ReblockedIterator) Next() *Record {
	next, ok := <-iter.stream
	if next == nil || !ok {
		return nil
	}
	iter.pos += len(next.Data)
	return next
}

func (iter *ReblockedIterator) Progress() (pos, max int) {
	pos, pax = iter.pos, iter.max
	return
}
