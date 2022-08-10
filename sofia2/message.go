package sofia2

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
)

/*
 *
 */
type LoginReq struct {
	EncryptType string // Encryption type, always MD5
	LoginType   string // Client identifier
	PassWord    string // Password, default is empty
	UserName    string // Username, default is admin
}

/*
 *
 */
type LoginRes struct {
	AliveInterval uint32 `json:"AliveInterval"`
	ChannelNum    int    `json:"ChannelNum"`
	DeviceType    string `json:"DeviceType "` // Notice the extra space before closing ", ate my whole day!
	ExtraChannel  int    `json:"ExtraChannel"`
	Ret           uint32 `json:"Ret"`
	SessionID     string `json:"SessionID"`
}

/*
 *
 */
type CmdReq struct {
	Name      string // Command name
	SessionID string // Session ID
}

type KeepAliveReq CmdReq

type CmdRes struct {
	Name      string `json: "Name"`      // Command name
	Ret       uint32 `json: "Ret"`       // Return code
	SessionID string `json: "SessionID"` // Session ID
}

type KeepAliveRes CmdRes

type SysInfo struct {
	Name       string // Command name
	Ret        uint32 // Return code
	SessionID  string // Session ID
	SystemInfo struct {
		AlarmInChannel  int
		AlarmOutChannel int
		AudioInChannel  int
		BuildTime       string
		CombineSwitch   int
		DeviceModel     string
		DeviceRunTime   string
		DeviceType      int
		DigChannel      int
		EncryptVersion  string
		ExtraChannel    int
		HardWare        string
		HardWareVersion string
		SerialNo        string
		SoftWareVersion string
		TalkInChannel   int
		TalkOutChannel  int
		UpdataTime      string
		UpdataType      string
		VideoInChannel  int
		VideoOutChannel int
	}
}

type SysAbility struct {
}

type SysOEMInfo struct {
}

type SysConfig struct {
}

/*
 *
 */
func EncodeMessage(msg *bytes.Buffer, sessionID byte, msgId uint16, data []byte) {
	msg.Reset()

	PrintMessageAsJSON(msgId, data)

	encMsgId := make([]byte, 2)
	encMsgLen := make([]byte, 4)

	// Encode message ID in little endian
	binary.LittleEndian.PutUint16(encMsgId, msgId)

	// Encode message length in little endian
	binary.LittleEndian.PutUint32(encMsgLen, uint32(len(data)))

	msg.WriteByte(0xFF)                             // Header flag
	msg.WriteByte(0x0)                              // Version
	msg.WriteByte(0x0)                              // Reserved field 1
	msg.WriteByte(0x0)                              // Reserved filed 2
	msg.WriteByte(sessionID)                        // Session ID
	msg.Write([]byte{0x00, 0x00, 0x00})             // Unknown field 1
	msg.WriteByte(0x00)                             // Sequence Number
	msg.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00}) // Unknown field 2
	msg.Write(encMsgId)                             // Message ID
	msg.Write(encMsgLen)                            // Data length
	msg.Write(data)                                 // Data
	msg.WriteByte(0x0A)                             // ASCII LF as terminator
}

/*
 *
 */
func DecodeMessage(msg *bytes.Buffer) (byte, uint16, uint16, []byte) {
	// Duplicate buffer
	rawBytes := msg.Bytes()

	// Session ID is at offset +4
	sessionId := rawBytes[4]

	// Message ID is at offset +14
	msgId := binary.LittleEndian.Uint16(rawBytes[14:])

	// Data length is at offset +15
	dataLen := binary.LittleEndian.Uint16(rawBytes[15:])

	// Data is at offset +20 with last 2 bytes truncated
	data := rawBytes[20 : len(rawBytes)-2]

	PrintMessageAsJSON(msgId, data)

	return sessionId, msgId, dataLen, data
}

/*
 *
 */
func PrintMessageAsJSON(msgID uint16, data []byte) {
	// Pretty-format
	var prettyStr bytes.Buffer
	json.Indent(&prettyStr, data, "", "    ")
	fmt.Printf("[%d]\n%s\n", msgID, &prettyStr)
}
