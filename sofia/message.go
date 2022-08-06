package sofia

import (
	"bytes"
	"encoding/binary"
)

/*
 *
 */
type LoginMsg struct {
	EncryptType string
	LoginType   string
	PassWord    string
	UserName    string
}

type LoginRes struct {
	AliveInterval int    `json:"AliveInterval"`
	ChannelNum    int    `json:"ChannelNum"`
	DeviceType    string `json:"DeviceType"`
	ExtraChannel  int    `json:"ExtraChannel"`
	Ret           int    `json:"Ret"`
	SessionID     string `json:"SessionID"`
}

/*
 *
 */
func MakeMessage(msg *bytes.Buffer, data []byte, msgId uint16) {
	msg.Reset()

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
	msg.WriteByte(0x0)                              // Session ID
	msg.Write([]byte{0x00, 0x00, 0x00})             // Unknown field 1
	msg.WriteByte(0x00)                             // Sequence Number
	msg.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00}) // Unknown field 2
	msg.Write(encMsgId)                             // Message ID
	msg.Write(encMsgLen)                            // Data length
	msg.Write(data)                                 // Data
	msg.WriteByte(0x0A)                             // ASCII LF as terminator
}
