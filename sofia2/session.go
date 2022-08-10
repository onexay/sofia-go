package sofia2

import (
	"bytes"
	"net"

	"github.com/teris-io/shortid"
)

/* Session
 *
 *
 */
type Session struct {
	ID        shortid.Shortid // Unique session ID
	Host      string          // Host, IP address or hostname
	Port      string          // Port
	User      string          // Username
	Password  string          // Password
	TheDevice *Device // Device instance
}

/*
 *
 */
func (s *Session)
