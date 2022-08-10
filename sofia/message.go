package sofia

// Message types
const (
	LOGIN_REQ1 = 999
	LOGIN_REQ2 = 1000
	LOGIN_RSP  = 1001
	LOGOUT_REQ = 1001
	LOGOUT_RSP = 1002
)

// Message format particulars
const (
	DeviceMessageHdeaderLen      = 18
	DeviceMessageTrailerLen      = 2
	DeviceMessageOffsetVersion   = 1
	DeviceMessageOffsetSessionId = 4
	DeviceMessageOffsetSeqNum    = 9
	DeviceMessageOffsetMsgId     = 14
	DeviceMessageOffsetDataLen   = 15
)

// Message for internal consumption (!wire format)
type DeviceMessage struct {
	msgId     uint16 // Message ID
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
