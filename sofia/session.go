package sofia

import (
	"encoding/json"
	"fmt"
)

/* Session
 *
 */
type Session struct {
	id        byte               // Session id as received from device
	idStr     string             // Session id as string
	opaqueId  uint8              // Opaque id (used as a correlation id)
	user      string             // Username
	password  string             // Password
	workerBus chan DeviceMessage // Message bus for device worker
	device    *Device            // Device instance
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
		session.workerBus = make(chan DeviceMessage, 100)
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
	msg := DeviceMessage{
		msgId:     LOGIN_REQ2,
		opaqueId:  session.opaqueId,
		version:   0,
		sessionId: 0,
		seqNum:    0,
		dataLen:   uint32(len(mdata)),
		data:      mdata,
	}

	// Send message to device
	session.device.SendMessage(&msg)

	// Receive message from device
	resMsg := <-session.workerBus

	// Unmarshall response data
	var resData LoginResData
	json.Unmarshal(resMsg.data, &resData)

	// Save params
	session.idStr = resData.SessionID

	return nil
}

// System Info
func (session *Session) SysInfo() error {
	// Data for sysinfo
	data := CmdReqData{
		Name:      "SysInfo",
		SessionID: session.idStr,
	}

	// Marshall data as JSON
	mdata, _ := json.Marshal(data)

	// Build message
	msg := DeviceMessage{
		msgId:     SYSINFO_REQ,
		opaqueId:  0,
		version:   0,
		sessionId: session.id,
		seqNum:    0,
		dataLen:   uint32(len(mdata)),
		data:      mdata,
	}

	// Send message to device
	session.device.SendMessage(&msg)

	// Receive message from device
	resMsg := <-session.workerBus

	fmt.Printf("SysInfo %d bytes\n", resMsg.dataLen)

	return nil
}
