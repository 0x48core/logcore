package store

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"sync"
)

// enc is the byte order used for all numeric encoding.
// BigEndian is chosen as it is the agreed-upon standard for network applications.
var enc = binary.BigEndian

// lenWidth is the number of bytes reserved to store the length of each message.
// Using 8 bytes (uint64) makes reads fully predictable:
//
//	read 8 bytes → get length N
//	read next N bytes → actual message
const lenWidth = 8

// store is an append-only binary file.
// It knows nothing about offsets — only byte positions.
//
// On-disk layout (repeating):
//
//	[ 8 bytes: uint64 message length ][ N bytes: raw message data ]
type store struct {
	file *os.File
	mu   sync.Mutex
	buf  *bufio.Writer // write buffer - must be flushed before any Read
	size uint64        // total bytes written so far; also the next append position
}

// newStore wraps an existing (or newly created) file as a store.
// It reads the current file size so that appends resume correctly after a restart.
func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, fmt.Errorf("error to stat file, %w", err)
	}
	return &store{
		file: f,
		size: uint64(fi.Size()),
		buf:  bufio.NewWriter(f),
	}, nil
}

// Append writes data to the store.
// It first writes an 8-byte length header, then the raw data.
// Returns:
//
//	n   — total bytes written (lenWidth + len(data))
//	pos — byte position at which this entry starts (used by the index)
//	err — any I/O error
func (s *store) Append(data []byte) (n, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pos = s.size // current end of file = start of new entry

	// Write the 8-byte length header.
	if err = binary.Write(s.buf, enc, uint64(len(data))); err != nil {
		return 0, 0, fmt.Errorf("error to write binary data, %w", err)
	}

	// Write the actual message bytes
	w, err := s.buf.Write(data)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to write content, %w", err)
	}

	w += lenWidth
	s.size += uint64(w)
	return uint64(w), pos, nil
}

// Read returns the message stored at the given byte position.
// It flushes the write buffer first because data may still be in memory.
func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Flush so any buffered-but-not-yet-writtten data is on disk
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}

	// Read the 8-byte length header.
	size := make([]byte, lenWidth)
	if _, err := s.file.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}

	// Read the message using the length we just decoded
	b := make([]byte, enc.Uint64(size))
	if _, err := s.file.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}

	return b, nil
}

// ReadAt is a raw pass-through to the underlying file's ReadAt.
// Used internally when rebuilding state from disk.
func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return 0, err
	}

	return s.file.ReadAt(p, off)
}

// Close flushes any buffered writes and closes the underlying file.
func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return err
	}

	return s.file.Close()
}