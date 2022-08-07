package sofia

/* Device
 *
 * A device is a resource which supports multiple user accounts and thus by
 * extension can support multiple active sessions. However, the sematics of
 * SOFIA server running on device allocates a session index for each unique
 * login instance.
 */

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

var msgEnd []byte = []byte{0x0A, 0x00}

// Instance of device
type Device struct {
	conn     net.Conn          // Network connection
	reqBuf   *bytes.Buffer     // Tx message buffer
	resBuf   *bytes.Buffer     // Rx message buffer
	sessions map[byte]*Session // Sessions
}

/*
 *
 */
func NewDevice() *Device {
	// Allocate a new device
	device := new(Device)

	// Allocate Tx message buffer
	device.reqBuf = new(bytes.Buffer)

	// Allocate Rx message buffer
	device.resBuf = new(bytes.Buffer)

	// Allocate map for sessions
	device.sessions = make(map[byte]*Session, 0xFF)

	// Return device
	return device
}

/*
 *
 */
func DeleteDevice(device *Device) {

}

/*
 *
 */
func (dev *Device) Connect(host string, port string) error {
	var err error = nil

	// Try connecting to device
	dev.conn, err = net.Dial("tcp", host+":"+port)

	return err
}

/*
 *
 */
func (dev *Device) Disconnect() {
	// Try disconnecting from device
	dev.conn.Close()
}

/*
 *
 */
func (dev *Device) ReadMessage() (byte, uint16, uint16, []byte) {
	// Clear response buffer
	dev.resBuf.Reset()

	// We read in chunks of 2 bytes (!TODO optimize)
	tmp := make([]byte, 2)

	for {
		// Read 2 bytes
		if _, err := dev.conn.Read(tmp); err != nil {
			if err == io.EOF {
				break
			}
		}

		// Append read bytes to main buffer
		dev.resBuf.Write(tmp)

		// Read bytes until message terminator is seen
		if bytes.Equal(tmp, msgEnd) {
			break
		}

		// Clear read bytes for next loop iteration
		tmp = []byte{0x00, 0x00}
	}

	// Decode message and return
	return DecodeMessage(dev.resBuf)
}

/*
 *
 */
func (dev *Device) Login(username string, password string) (*Session, error) {
	// Check user for default username
	if len(username) == 0 {
		username = "admin"
	}

	// Check password for default password
	if len(password) == 0 {
		password = "tlJwpbo6"
	}

	// Build login message
	loginData := LoginReq{
		EncryptType: "MD5",
		LoginType:   "Sofia-Go",
		UserName:    username,
		PassWord:    password,
	}

	// Marshal message to JSON
	data, _ := json.Marshal(loginData)

	// Build device message
	EncodeMessage(dev.reqBuf, data, LOGIN_REQ2)

	// Send message
	if _, err := dev.conn.Write(dev.reqBuf.Bytes()); err != nil {
		return nil, err
	}

	return dev.LoginResponse(), nil
}

/*
 *
 */
func (dev *Device) LoginResponse() *Session {
	var res LoginRes

	// Read response
	sessionId, msgId, _, data := dev.ReadMessage()

	// Unmarshall data to JSON
	if err := json.Unmarshal(data, &res); err != nil {
		return nil
	}

	if msgId != LOGIN_RSP {
		return nil
	}

	if res.Ret != RetOK {
		return nil
	}

	// Lookup session
	session, found := dev.sessions[sessionId]
	if !found {
		session = NewSesion(sessionId, res.SessionID)
	}

	return session
}

/*
 *
 */
func (device *Device) Logout() {

}

func (dev *Device) SystemInfo(s *Session) (SysInfo, error) {
	// Clear buffer
	dev.reqBuf.Reset()

	data := CmdReq{
		Name:      "SystemInfo",
		SessionID: *s.IDStr(),
	}

	// Marshal data to JSON
	encodedData, _ := json.Marshal(data)

	// Encode message
	EncodeMessage(dev.reqBuf, encodedData, SYSINFO_REQ)

	// Send message
	if _, err := dev.conn.Write(dev.reqBuf.Bytes()); err != nil {
		return SysInfo{}, err
	}

	return dev.SystemInfoResponse(), nil
}

func (dev *Device) SystemInfoResponse() SysInfo {
	// Read response
	_, msgId, _, data := dev.ReadMessage()

	// Create a map
	var res SysInfo

	// Unmarshall data to JSON
	if err := json.Unmarshal(data, &res); err != nil {
		fmt.Printf("1 %s\n", err.Error())
		return SysInfo{}
	}

	if msgId != SYSINFO_RSP {
		fmt.Printf("2 %d\n", msgId)
		return SysInfo{}
	}

	return res
}
