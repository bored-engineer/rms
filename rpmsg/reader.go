package rpmsg

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// https://docs.microsoft.com/en-us/previous-versions/windows/internet-explorer/ie-developer/platform-apis/aa767786(v=vs.85)?redirectedfrom=MSDN#prefix-and-save-the-file
var magicBytes = []byte{0x76, 0xe8, 0x04, 0x60, 0xc4, 0x11, 0xe3, 0x86}

// https://docs.microsoft.com/en-us/previous-versions/windows/internet-explorer/ie-developer/platform-apis/aa767786(v=vs.85)?redirectedfrom=MSDN#compress-the-resulting-compound-file
var segmentBytes = []byte{0xA0, 0x0F, 0x00, 0x00}

// segmentReader reads segments from an rpmsg
type segmentReader struct {
	// underlying io.Reader of rpmsg data
	r io.Reader
	// segment header, re-used for memory usage
	header [12]byte
	// buf is data left over from one segment between reads
	buf []byte
	// expected is the expected size of the gzip decompressed data
	expected uint64
}

// readSegment reads a segment from the io.Reader into the internal buffer
func (r *segmentReader) readSegment() error {
	// Read in the header and make sure it has the right prefix
	if _, err := io.ReadFull(r.r, r.header[:]); err != nil {
		// EOF reading a header is expected, return it as-is
		if err == io.EOF {
			return io.EOF
		}
		return fmt.Errorf("failed to read segment header: %v", err)
	}
	if !bytes.Equal(r.header[0:4], segmentBytes) {
		return errors.New("failed to match magic segment prefix")
	}
	originalSize := binary.LittleEndian.Uint32(r.header[4:8])
	compressedSize := binary.LittleEndian.Uint32(r.header[8:12])
	r.buf = make([]byte, compressedSize)
	r.expected += uint64(originalSize)
	if _, err := io.ReadFull(r.r, r.buf); err != nil {
		return fmt.Errorf("failed to read compressed segment: %v", err)
	}
	return nil
}

// Read reads segments until p is full or an error occurs
func (r *segmentReader) Read(p []byte) (int, error) {
	// Loop until we have filled p with data from segments
	copied := 0
	size := len(p)
	for copied < size {
		// If we have data in the buffer, copy that first
		if avail := len(r.buf); avail > 0 {
			// If the data doesn't fit in p, truncate
			if rem := (size - copied); rem < avail {
				copied += copy(p[copied:], r.buf[:rem])
				// TODO: This will cost us memory as go retains the full array
				r.buf = r.buf[rem:]
				return copied, nil
			}
			// Copy the full buffer and release it
			copied += copy(p[copied:], r.buf)
			r.buf = nil
		}
		// Read the next segment
		if err := r.readSegment(); err == io.EOF {
			// EOF is a special case and is not wrapped
			return copied, io.EOF
		} else if err != nil {
			return copied, fmt.Errorf("failed to read segment: %v", err)
		}
	}
	return copied, nil
}

// zlibReader converts the io.ReadCloser to io.Reader
type zlibReader struct {
	r    *segmentReader
	zr   io.ReadCloser
	read uint64
}

func (r *zlibReader) Read(p []byte) (int, error) {
	// If no zlib reader yet create one
	if r.zr == nil {
		var err error
		r.zr, err = zlib.NewReader(r.r)
		if err != nil {
			return 0, fmt.Errorf("failed to read zlib header: %v", err)
		}
	}
	// Read directly into p
	n, err := io.ReadFull(r.zr, p)
	r.read += uint64(n)
	if err != nil {
		// TODO: Hacky af
		if err == io.EOF || err.Error() == "unexpected EOF" {
			if r.read == r.r.expected {
				return n, io.EOF
			} else {
				return n, errors.New("unexpected EOF")
			}
		}
		return n, fmt.Errorf("failed to read zlib data: %v", err)
	}
	return n, nil
}

// NewReader reads the prefix and reads
func NewReader(r io.Reader) (io.Reader, error) {
	// Read in the prefix and make sure it's the magic bytes
	var prefix [8]byte
	if _, err := io.ReadFull(r, prefix[:]); err != nil {
		return nil, fmt.Errorf("failed to read magic bytes: %v", err)
	}
	if !bytes.Equal(prefix[:], magicBytes) {
		return nil, errors.New("failed to match magic prefix")
	}
	// Return a zlib reader wrapping the segment reader
	return &zlibReader{r: &segmentReader{r: r}}, nil
}
