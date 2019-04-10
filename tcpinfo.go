package gremgoser

import (
	"errors"
	"net"
	"syscall"
	"unsafe"
)

type TCPInfo syscall.TCPInfo

func getsockopt(s int, level int, name int, val uintptr, vallen *uint32) (err error) {
	_, _, e1 := syscall.Syscall6(syscall.SYS_GETSOCKOPT, uintptr(s), uintptr(level), uintptr(name), uintptr(val), uintptr(unsafe.Pointer(vallen)), 0)
	if e1 != 0 {
		err = e1
	}
	return
}

func GetsockoptTCPInfo(conn *net.Conn) (*TCPInfo, error) {
	tcpConn, ok := (*conn).(*net.TCPConn)
	if !ok {
		return nil, errors.New("not a TCPConn")
	}

	file, err := tcpConn.File()
	if err != nil {
		return nil, err
	}

	fd := file.Fd()
	tcpInfo := TCPInfo{}
	size := uint32(unsafe.Sizeof(tcpInfo))
	err = getsockopt(int(fd), syscall.SOL_TCP, syscall.TCP_INFO, uintptr(unsafe.Pointer(&tcpInfo)), &size)
	if err != nil {
		return nil, err
	}

	return &tcpInfo, nil
}
