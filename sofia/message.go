package sofia

// Message types
const (
	LOGIN_REQ1  = 999
	LOGIN_REQ2  = 1000
	LOGIN_RSP   = 1001
	LOGOUT_REQ  = 1001
	LOGOUT_RSP  = 1002
	SYSINFO_REQ = 1020
	SYSINFO_RSP = 1021
)

// Message format particulars
const (
	DeviceMessageHdeaderLen      = 18
	DeviceMessageTrailerLen      = 2
	DeviceMessageOffsetVersion   = 1
	DeviceMessageOffsetSessionId = 4
	DeviceMessageOffsetSeqNum    = 8
	DeviceMessageOffsetOpaqueId  = 9
	DeviceMessageOffsetMsgId     = 14
	DeviceMessageOffsetDataLen   = 15
)

// Message for internal consumption (!wire format)
type DeviceMessage struct {
	msgId     uint16 // Message ID
	opaqueId  uint8  // Opaque ID (meant to be sent back by device)
	version   byte   // Version
	sessionId byte   // Session ID
	seqNum    byte   // Sequence number
	dataLen   uint32 // Payload length
	data      []byte // Payload
}

// Login request data
type LoginReqData struct {
	EncryptType string // Encryption type, always MD5
	LoginType   string // Client identifier
	PassWord    string // Password, default is empty
	UserName    string // Username, default is admin
}

// Login response data
type LoginResData struct {
	AliveInterval uint32 `json:"AliveInterval"`
	ChannelNum    int    `json:"ChannelNum"`
	DeviceType    string `json:"DeviceType "` // Notice the extra space before closing ", ate my whole day!
	ExtraChannel  int    `json:"ExtraChannel"`
	Ret           uint32 `json:"Ret"`
	SessionID     string `json:"SessionID"`
}

// Generic request data
type CmdReqData struct {
	Name      string // Command name
	SessionID string // Session ID
}

// Generic response data
type CmdResData struct {
	Name      string `json: "Name"`      // Command name
	Ret       uint32 `json: "Ret"`       // Return code
	SessionID string `json: "SessionID"` // Session ID
}

type KeepAliveReqData CmdReqData
type KeepAliveResData CmdResData
