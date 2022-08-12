package sofia

import (
	"encoding/json"
	"fmt"
	"time"
)

/* Session
 *
 */
type Session struct {
	id         byte               // Session id as received from device
	idStr      string             // Session id as string
	opaqueId   uint8              // Opaque id (used as a correlation id)
	seqNum     uint8              // Sequence number
	kaInterval uint32             // Keepalive interval
	user       string             // Username
	password   string             // Password
	rxChan     chan DeviceMessage // Channel for receiving messages
	device     *Device            // Device instance
}

/*
 *
 */
func NewSession(device *Device, localId uint8, user string, password string) *Session {
	// Allocate new session
	var session *Session = new(Session)

	// Play with ids
	{
		session.id = 0
		session.opaqueId = localId
		session.seqNum = 0
		session.kaInterval = 0
	}

	// Save username and password
	{
		if session.user = user; len(user) == 0 {
			session.user = "admin"
		}

		if session.password = password; len(password) == 0 {
			session.password = "tlJwpbo6"
		}
	}

	// Initialize worker message bus
	{
		session.rxChan = make(chan DeviceMessage)
	}

	// Save device context
	{
		session.device = device
	}

	return session
}

/*
 *
 */
func DeleteSession(session *Session) {
	// Close channel
	close(session.rxChan)
}

// Keepalive task
func (session *Session) keepAliveTask() {
	session.device.logger.Info("Starting KA task for session ", session.idStr)

	// Create a ticker
	ticker := time.NewTicker(time.Second * time.Duration(session.kaInterval))

	for {
		select {
		case <-ticker.C:
			// Send a keep alive message
			if err := session.KeepAlive(); err != nil {
				return
			}
		}
	}
}

// Build message
func (session *Session) BuildMessage(msgId uint16, data []byte) DeviceMessage {
	return DeviceMessage{
		msgId:     msgId,
		opaqueId:  session.opaqueId,
		version:   0,
		sessionId: session.id,
		seqNum:    session.seqNum,
		dataLen:   uint32(len(data)),
		data:      data,
	}
}

// Login to device
func (session *Session) Login() error {
	// Data for login
	data := LoginReqData{
		EncryptType: "MD5",
		LoginType:   "Sofia-Go",
		PassWord:    session.password,
		UserName:    session.user,
	}

	// Marshall data as JSON
	mdata, _ := json.Marshal(data)

	// Build message
	msg := session.BuildMessage(LOGIN_REQ2, mdata)

	// Send message to device
	session.device.SendMessage(&msg)

	// Receive message from device
	resMsg := <-session.rxChan

	// Unmarshall response data
	var resData LoginResData
	if err := json.Unmarshal(resMsg.data, &resData); err != nil {
		fmt.Printf("Unable to parse response for session 0x%X, message %d [%s]\n", resMsg.sessionId, resMsg.msgId, err.Error())
		return err
	}

	session.kaInterval = resData.AliveInterval
	session.idStr = resData.SessionID
	fmt.Printf("Login success for session %s\n", resData.SessionID)

	// Start KA task
	go session.keepAliveTask()

	return nil
}

// Session keep-alive
func (session *Session) KeepAlive() error {
	// Data for keepalive
	data := KeepAliveReqData{
		Name:      "KeepAlive",
		SessionID: session.idStr,
	}

	// Marshall data as JSON
	mdata, _ := json.Marshal(data)

	// Build message
	msg := session.BuildMessage(KEEPALIVE_REQ, mdata)

	// Send message to device
	if err := session.device.SendMessage(&msg); err != nil {
		return err
	}

	// Receive message from device
	resMsg := <-session.rxChan

	// Unmarshall response data
	var resData KeepAliveResData
	if err := json.Unmarshal(resMsg.data, &resData); err != nil {
		return err
	}

	return nil
}

// System Info
func (session *Session) SysInfo() error {
	// Data for sysinfo
	data := CmdReqData{
		Name:      "SystemInfo",
		SessionID: session.idStr,
	}

	// Marshall data as JSON
	mdata, _ := json.Marshal(data)

	// Build message
	msg := session.BuildMessage(SYSINFO_REQ, mdata)

	// Send message to device
	session.device.SendMessage(&msg)

	// Receive message from device
	resMsg := <-session.rxChan

	fmt.Printf("[%s] SysInfo %d bytes\n", session.idStr, resMsg.dataLen)

	var x map[string]interface{}
	json.Unmarshal(resMsg.data, &x)
	dumpJSON(data.Name, x, "")

	return nil
}

// System Abilities
func (session *Session) SysAbilities() error {
	// Data for sysinfo
	data := CmdReqData{
		Name:      "SystemFunction",
		SessionID: session.idStr,
	}

	// Marshall data as JSON
	mdata, _ := json.Marshal(data)

	// Build message
	msg := session.BuildMessage(ABILITY_REQ, mdata)

	// Send message to device
	session.device.SendMessage(&msg)

	// Receive message from device
	resMsg := <-session.rxChan

	fmt.Printf("[%s] System Abilities %d bytes\n", session.idStr, resMsg.dataLen)

	var x map[string]interface{}
	json.Unmarshal(resMsg.data, &x)
	dumpJSON(data.Name, x, "")

	return nil
}

// System OEM info
func (session *Session) SysOEMInfo() error {
	// Data for sysinfo
	data := CmdReqData{
		Name:      "OEMInfo",
		SessionID: session.idStr,
	}

	// Marshall data as JSON
	mdata, _ := json.Marshal(data)

	// Build message
	msg := session.BuildMessage(SYSINFO_REQ, mdata)

	// Send message to device
	session.device.SendMessage(&msg)

	// Receive message from device
	resMsg := <-session.rxChan

	fmt.Printf("[%s] System OEM info %d bytes\n", session.idStr, resMsg.dataLen)

	var x map[string]interface{}
	json.Unmarshal(resMsg.data, &x)
	dumpJSON(data.Name, x, "")

	return nil
}
