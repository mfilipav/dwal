package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

// Binary encoding
var enc = binary.BigEndian

// Bytes used to store length of the record
const lenWidth = 8

// Store struct is a wrapper around a file with two methods
// to Append to and Read bytes from file
type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

// Creates a store for some file `f`
func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := uint64(fi.Size())
	return &store{
		File: f,
		buf:  bufio.NewWriter(f),
		size: size,
	}, nil
}

// Appends given bytes to the store. Returns: bytes written,
// position where store holds the file
func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pos = s.size
	// write to buffered writer
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}
	bytesWritten, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}
	bytesWritten += lenWidth
	s.size += uint64(bytesWritten)
	return uint64(bytesWritten), pos, nil
}

// Reads from store at given pos and returns the record
func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// flush writer buffer, before attempting to read
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}
	// calculate record size
	size := make([]byte, lenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}
	b := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}
	return b, nil
}

// Reads into file at offset, returns bytes read
func (s *store) ReadAt(p []byte, offset int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, offset)
}

// Flush buffered data before closing file
func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.buf.Flush()
	if err != nil {
		return err
	}
	return s.File.Close()
}
