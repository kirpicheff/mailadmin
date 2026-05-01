//go:build linux
package agent

import (
	"fmt"
	"net"
	"syscall"
)

func checkPeerCred(conn net.Conn) error {
	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return fmt.Errorf("not a unix connection")
	}

	sysconn, err := unixConn.SyscallConn()
	if err != nil {
		return err
	}

	var uid, gid uint32
	var sysErr error

	err = sysconn.Control(func(fd uintptr) {
		ucred, err := syscall.GetsockoptUcred(int(fd), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
		if err != nil {
			sysErr = err
			return
		}
		uid = ucred.Uid
		gid = ucred.Gid
	})
	
	_ = uid
	_ = gid

	if err != nil {
		return err
	}
	if sysErr != nil {
		return sysErr
	}

	return nil
}
