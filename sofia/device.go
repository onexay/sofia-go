package sofia

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

var msgEnd []byte = []byte{0x0A, 0x00}

/*
 *
 */
type Device struct {
	host     string            // Hostname or IP address
	port     string            // Port
	proto    string            // Protocol
	conn     net.Conn          // Network connection
	username string            // Username
	password string            // Password
	reqData  *bytes.Buffer     // Request data buffer
	resData  *bytes.Buffer     // Response data buffer
	sessions map[byte]*Session // Sessions
}

/*
 *
 */
func NewDevice(host string, port string, proto string, username string, password string) *Device {
	device := new(Device)

	device.host = host
	device.port = port
	device.proto = proto

	if len(username) == 0 {
		device.username = "admin"
	} else {
		device.username = username
	}

	if len(password) == 0 {
		device.password = "tlJwpbo6"
	} else {
		device.password = password
	}

	device.reqData = new(bytes.Buffer)
	device.resData = new(bytes.Buffer)
	device.sessions = make(map[byte]*Session)

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
func ReadDevice(device *Device) {
	tmp := make([]byte, 2)

	for {
		// Read 2 bytes
		_, err := device.conn.Read(tmp)

		if err != nil {
			if err == io.EOF {
				break
			}
		}

		// Append read bytes to main buffer
		device.resData.Write(tmp)

		// Read bytes until message terminator is seen
		if bytes.Equal(tmp, msgEnd) {
			// Handle message
			HandleMessage(device)

			// Clear main buffer once one whole message is read
			device.resData.Reset()
		}

		// Clear read bytes for next loop iteration
		tmp = []byte{0x00, 0x00}
	}
}

/*
 *
 */
func HandleMessage(device *Device) {
	// Take message bytes
	raw := device.resData.Bytes()

	// Session ID
	sessionID := raw[4]

	// Message ID
	msgID := binary.LittleEndian.Uint16(raw[14:])

	// Read message data
	data := raw[20 : len(raw)-2]

	switch msgID {
	case LOGIN_RSP:
		device.LoginRes(sessionID, data)

	case SYSINFO_RSP:
		if session, found := device.sessions[sessionID]; found {
			session.SysInfoRes(data)
		}
	}
}

/*
 *
 */
func (device *Device) Connect() error {
	var err error = nil

	// Try connecting to device
	device.conn, err = net.Dial(device.proto, device.host+":"+device.port)

	return err
}

/*
 *
 */
func (device *Device) Disconnect() error {
	// Disconnect device connection
	return device.conn.Close()
}

/*
 *
 */
func (device *Device) Login() error {
	// Build login message
	loginData := LoginReq{
		EncryptType: "MD5",
		LoginType:   "Sofia-Go",
		UserName:    device.username,
		PassWord:    device.password,
	}

	// Marshal message to JSON
	data, _ := json.Marshal(loginData)

	// Build device message
	MakeMessage(device.reqData, data, LOGIN_REQ2)

	// Send message
	_, err := device.conn.Write(device.reqData.Bytes())

	return err
}

func (device *Device) LoginRes(sessionId byte, data []byte) {
	var res LoginRes

	// Unmarshall
	json.Unmarshal(data, &res)

	if res.Ret != 100 {
		return
	}

	// Try to find a session
	_, found := device.sessions[sessionId]
	if found {
		fmt.Printf("Session found, unexpected!")
		return
	}

	device.sessions[sessionId] = NewSesion(sessionId)

	sysInfo := CmdReq{
		Name:      "ConfigGet",
		SessionID: res.SessionID,
	}

	nData, _ := json.Marshal(sysInfo)

	MakeMessage(device.reqData, nData, CONFIG_GET)

	device.conn.Write(device.reqData.Bytes())
}

/*
 *
 */
func (device *Device) Logout() {

}
