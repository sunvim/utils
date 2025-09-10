package mmap

import (
	"syscall"
	"unsafe"
)

func Alloc(eltsize uintptr, size int) ([]byte, error) {
	var errno syscall.Errno
	fd := -1
	data, _, errno := syscall.Syscall6(
		syscall.SYS_MMAP,
		0, // address
		eltsize*uintptr(size),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANON|syscall.MAP_PRIVATE,
		uintptr(fd), // No file descriptor
		0,           // offset
	)

	if errno != 0 {
		return nil, errno
	}

	return unsafe.Slice((*byte)(unsafe.Pointer(data)), size*int(eltsize)), nil
}

// Free releases resources allocated via Alloc
func Free(data unsafe.Pointer, size uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_MUNMAP, uintptr(data), size, 0)
	if errno != 0 {
		return errno
	}
	return nil
}
