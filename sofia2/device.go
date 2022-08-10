package sofia2

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
	"time"

	"github.com/sirupsen/logrus"
)

/*
 *
 */
type Device struct {
	Logger    logrus.Logger       // Device scoped logger
	RxBuf     bytes.Buffer        // Receive buffer
	TxBuf     bytes.Buffer        // Transmit buffer
	Transport net.Conn            // Transport connection
	Sessions  map[string]*Session // Device sessions
}

var msgEnd []byte = []byte{0x0A, 0x00}

// Instance of device
type Device struct {
	addr     string            // Host address
	port     string            // Host port, default is 34567
	username string            // Username (default is 'admin')
	password string            // Password
	conn     net.Conn          // Network connection
	reqBuf   *bytes.Buffer     // Tx message buffer
	resBuf   *bytes.Buffer     // Rx message buffer
	sessions map[byte]*Session // Sessions
}

/*
 *
 */
func NewDevice(addr string, port string, username string, password string) *Device {
	// Allocate a new device
	device := new(Device)

	// Save host address
	device.addr = addr

	// Check for default port
	if device.port = port; len(device.port) == 0 {
		device.port = "34567"
	}

	// Check user for default username
	if device.username = username; len(device.username) == 0 {
		device.username = "admin"
	}

	// Check password for default password
	if device.password = password; len(device.password) == 0 {
		device.password = "tlJwpbo6"
	}

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
func DeleteDevice(dev *Device) {
	dev.Disconnect()
}

/*
 *
 */
func (dev *Device) KeepAliveTask(session *Session) {
	fmt.Printf("Starting KA task of %d seconds for session %s\n", session.kaInterval, session.idStr)

	// Create timer
	ticker := time.NewTicker(time.Second * time.Duration(session.kaInterval))

	for {
		select {
		case val := <-session.kaChan:
			// Check for values sent on channel
			if val == 0 {
				fmt.Printf("Stopping KA task for session %s\n", session.idStr)

				// Stop timer and exit
				ticker.Stop()
			} else {
				// Value is for modifying timer
				ticker.Reset(time.Second * time.Duration(val))
			}
		case <-ticker.C:
			fmt.Printf("Ticking KA task for session %s\n", session.idStr)

			// Send keep alive
			if dev.KeepAlive(session) != nil {
				ticker.Stop()
				fmt.Printf("Terminating KA task for session %s\n", session.idStr)
			}
		}
	}
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
	EncodeMessage(dev.reqBuf, 0, LOGIN_REQ2, data)

	// Send message
	if _, err := dev.conn.Write(dev.reqBuf.Bytes()); err != nil {
		return nil, err
	}

	return dev.LoginResponse(LOGIN_RSP), nil
}

/*
 *
 */
func (dev *Device) LoginResponse(expMsgId uint16) *Session {
	var res LoginRes

	// Read response
	sessionId, msgId, _, data := dev.ReadMessage()

	// Unmarshall data to JSON
	if err := json.Unmarshal(data, &res); err != nil {
		return nil
	}

	if msgId != expMsgId {
		return nil
	}

	if res.Ret != RetOK {
		return nil
	}

	// Lookup session
	session, found := dev.sessions[sessionId]
	if !found {
		// Create new session
		session = NewSesion(sessionId, res.AliveInterval, res.SessionID)

		// Start keepalive task
		go dev.KeepAliveTask(session)
	}

	return session
}

/*
 *
 */
func (dev *Device) KeepAlive(session *Session) error {
	// Clear buffer
	dev.reqBuf.Reset()

	data := KeepAliveReq{
		Name:      "KeepAlive",
		SessionID: *session.IDStr(),
	}

	// Encode as JSON
	encodedData, _ := json.Marshal(data)

	// Encode message
	EncodeMessage(dev.reqBuf, session.id, KEEPALIVE_REQ, encodedData)

	// Send message
	if _, err := dev.conn.Write(dev.reqBuf.Bytes()); err != nil {
		return err
	}

	err := dev.KeepAliveResponse(KEEPALIVE_RSP)

	return err
}

/*
 *
 */
func (dev *Device) KeepAliveResponse(expMsgId uint16) error {
	// Read message
	_, msgId, _, data := dev.ReadMessage()

	// Check message id
	if msgId != expMsgId {
		//return nil
	}

	// Unmarshall data to JSON
	var res KeepAliveRes
	json.Unmarshal(data, res)

	return nil
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
	EncodeMessage(dev.reqBuf, s.id, SYSINFO_REQ, encodedData)

	// Send message
	if _, err := dev.conn.Write(dev.reqBuf.Bytes()); err != nil {
		return SysInfo{}, err
	}

	return dev.SystemInfoResponse(SYSINFO_RSP), nil
}

func (dev *Device) SystemInfoResponse(expMsgId uint16) SysInfo {
	// Read response
	_, msgId, _, data := dev.ReadMessage()

	// Create a map
	var res SysInfo

	// Unmarshall data to JSON
	if err := json.Unmarshal(data, &res); err != nil {
		fmt.Printf("1 %s\n", err.Error())
		return SysInfo{}
	}

	if msgId != expMsgId {
		return SysInfo{}
	}

	return res
}

func (dev *Device) SystemAbility(s *Session) (SysAbility, error) {
	// Clear buffer
	dev.reqBuf.Reset()

	data := CmdReq{
		Name:      "SystemFunction",
		SessionID: *s.IDStr(),
	}

	// Marshal data to JSON
	encodedData, _ := json.Marshal(data)

	// Encode message
	EncodeMessage(dev.reqBuf, s.id, ABILITY_GET, encodedData)

	// Send message
	if _, err := dev.conn.Write(dev.reqBuf.Bytes()); err != nil {
		return SysAbility{}, err
	}

	return dev.SystemAbilityResponse(ABILITY_GET_RSP), nil
}

func (dev *Device) SystemAbilityResponse(expMsgId uint16) SysAbility {
	// Read response
	_, msgId, _, data := dev.ReadMessage()

	// Create a map
	var res SysAbility

	// Unmarshall data to JSON
	if err := json.Unmarshal(data, &res); err != nil {
		fmt.Printf("1 %s\n", err.Error())
		return SysAbility{}
	}

	if msgId != expMsgId {
		fmt.Printf("2 %d\n", msgId)
		return SysAbility{}
	}

	return res
}

func (dev *Device) SystemOEMInfo(s *Session) (SysOEMInfo, error) {
	// Clear buffer
	dev.reqBuf.Reset()

	data := CmdReq{
		Name:      "OEMInfo",
		SessionID: *s.IDStr(),
	}

	// Marshal data to JSON
	encodedData, _ := json.Marshal(data)

	// Encode message
	EncodeMessage(dev.reqBuf, s.id, SYSINFO_REQ, encodedData)

	// Send message
	if _, err := dev.conn.Write(dev.reqBuf.Bytes()); err != nil {
		return SysOEMInfo{}, err
	}

	return dev.SystemOEMInfoResponse(SYSINFO_RSP), nil
}

func (dev *Device) SystemOEMInfoResponse(expMsgId uint16) SysOEMInfo {
	// Read response
	_, msgId, _, data := dev.ReadMessage()

	// Create a map
	var res SysOEMInfo

	// Unmarshall data to JSON
	if err := json.Unmarshal(data, &res); err != nil {
		fmt.Printf("1 %s\n", err.Error())
		return SysOEMInfo{}
	}

	if msgId != expMsgId {
		fmt.Printf("2 %d\n", msgId)
		return SysOEMInfo{}
	}

	return res
}

func (dev *Device) SystemConfig(s *Session, what string) (SysConfig, error) {
	// Clear buffer
	dev.reqBuf.Reset()

	data := CmdReq{
		Name:      what,
		SessionID: *s.IDStr(),
	}

	// Marshal data to JSON
	encodedData, _ := json.Marshal(data)

	// Encode message
	EncodeMessage(dev.reqBuf, s.id, CONFIG_GET, encodedData)

	// Send message
	if _, err := dev.conn.Write(dev.reqBuf.Bytes()); err != nil {
		return SysConfig{}, err
	}

	return dev.SystemConfigResponse(CONFIG_GET_RSP), nil
}

func (dev *Device) SystemConfigResponse(expMsgId uint16) SysConfig {
	// Read response
	_, msgId, _, data := dev.ReadMessage()

	// Create a map
	var res SysConfig

	// Unmarshall data to JSON
	if err := json.Unmarshal(data, &res); err != nil {
		fmt.Printf("1 %s\n", err.Error())
		return SysConfig{}
	}

	if msgId != expMsgId {
		fmt.Printf("2 %d\n", msgId)
		return SysConfig{}
	}

	return res
}
