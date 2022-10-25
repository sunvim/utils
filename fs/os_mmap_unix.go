//go:build !windows

package fs

import (
	"os"
	"syscall"
	"unsafe"
)

func Mmap(f *os.File, mappingSize int64) ([]byte, error) {
	p, err := syscall.Mmap(int(f.Fd()), 0, int(mappingSize), syscall.PROT_READ, syscall.MAP_SHARED|syscall.MAP_POPULATE)
	return p, err
}

func Munmap(data []byte) error {
	return syscall.Munmap(data)
}

func MadviceRandom(data []byte) error {
	_, _, errno := syscall.Syscall(syscall.SYS_MADVISE, uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)), uintptr(syscall.MADV_RANDOM))
	if errno != 0 {
		return errno
	}
	return nil
}

func (f *osMMapFile) Truncate(size int64) error {
	if err := f.File.Truncate(size); err != nil {
		return err
	}
	f.size = size
	return f.mremap()
}
