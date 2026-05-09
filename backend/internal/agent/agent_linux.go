//go:build linux
package agent

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

func SetSocketPermissions(path string) error {
	// Устанавливаем права 0660 (владелец и группа смогут читать/писать)
	if err := os.Chmod(path, 0660); err != nil {
		return err
	}

	// Ищем пользователя mailadmin
	u, err := user.Lookup("mailadmin")
	if err == nil {
		uid, _ := strconv.Atoi(u.Uid)
		// Меняем владельца файла на mailadmin
		if err := os.Chown(path, uid, -1); err != nil {
			return fmt.Errorf("failed to chown socket to user mailadmin: %v", err)
		}
	}
	return nil
}

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
