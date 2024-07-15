package log

import (
	"io"
	"os"

	"github.com/tysonmote/gommap" // use https://pkg.go.dev/github.com/go-mmap/mmap
)

var (
	offWidth   uint64 = 4
	posWidth   uint64 = 8
	entryWidth        = offWidth + posWidth
)

// index file, containing a mapping between record's offset
// and its position in the store file
type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64 // where to write next entry
}

// Maps file f into memory; returns index struct which contains the file
func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	idx.size = uint64(fi.Size())

	if err = os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}
	// memory map entire idx file, use read/write flags,
	if idx.mmap, err = gommap.Map(
		idx.file.Fd(),
		gommap.PROT_READ|gommap.PROT_WRITE,
		gommap.MAP_SHARED,
	); err != nil {
		return nil, err
	}
	return idx, nil
}

// Ensures that memory-mapped index file syncs its data to persisted file
// and that persisted file has flushed its contents to stable storage.
// Also, truncates persisted file to the amount of data that's
// actually in it and closes the file.
func (i *index) Close() error {
	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}
	if err := i.file.Sync(); err != nil {
		return err
	}
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}
	return i.file.Close()
}

// For a given offset, returns associated record's position in the store
func (i *index) Read(off int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}
	if off == -1 {
		out = uint32((i.size / entryWidth) - 1)
	} else {
		out = uint32(off)
	}
	pos = uint64(out) * entryWidth
	// validate that enough space to write
	if i.size < pos+entryWidth {
		return 0, 0, io.EOF
	}
	out = enc.Uint32(i.mmap[pos : pos+offWidth])
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entryWidth])
	return out, pos, nil
}

// Appends the given offset and position to the index
func (i *index) Write(off uint32, pos uint64) error {
	if uint64(len(i.mmap)) < i.size+entryWidth {
		return io.EOF
	}
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entryWidth], pos)
	i.size += uint64(entryWidth)
	return nil
}

func (i *index) Name() string {
	return i.file.Name()
}
