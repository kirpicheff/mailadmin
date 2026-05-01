//go:build !linux
package agent

import (
	"net"
)

func checkPeerCred(conn net.Conn) error {
	// На Windows и других ОС эта проверка не поддерживается или не требуется для разработки
	return nil
}

func SetSocketPermissions(path string) error {
	return nil
}
