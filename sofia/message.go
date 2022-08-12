package sofia

// Message types
const (
	LOGIN_REQ1    = 999
	LOGIN_REQ2    = 1000
	LOGIN_RSP     = 1001
	LOGOUT_REQ    = 1001
	LOGOUT_RSP    = 1002
	SYSINFO_REQ   = 1020
	SYSINFO_RSP   = 1021
	ABILITY_REQ   = 1360
	ABILITY_RSP   = 1361
	KEEPALIVE_REQ = 1006 // 1005 on some devices
	KEEPALIVE_RSP = 1007 // 1006 on some devices
)

/*
	<-1----------------->|<-2-------------->|...
     0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+...+-+-+
	|A|B| C |D|  E  |F|    G    | H |   I   | J | K |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+...+-+-+
	 | |  |  |   |   |     |      |     |     |   |
	 | |  |  |   |   |     |      |     |     |   +--> Trailer, 2B
	 | |  |  |   |	 |	   |      |     |     +--> Data, @Data Length B
	 | |  |  |   |   | 	   |      |     +--> Data Length, 4B
	 | |  |  |   |   |     |      +--> Message ID, 2B
	 | |  |  |   |   |     +--> Unknown, 5B
	 | |  |  |   |   +--> Sequence Number, 1B
	 | |  |  |   +--> Unknown, 3B
	 | |  |  +--> Session ID, 1B
	 | |  +--> Reserved, 2B
	 | +--> Version, 1B
	 +--> Header flag, 1B, always 0xFF
*/

// Message format particulars
const (
	DeviceMessageHeaderLen       = 20
	DeviceMessageTrailerLen      = 2
	DeviceMessageOffsetVersion   = 1
	DeviceMessageOffsetSessionId = 4
	DeviceMessageOffsetSeqNum    = 8
	DeviceMessageOffsetOpaqueId  = 9
	DeviceMessageOffsetMsgId     = 14
	DeviceMessageOffsetDataLen   = 16
	DeviceMessageOffsetData      = 20
)

// Message header for internal consumption (!wire format)
type DeviceMessageHeader struct {
	msgId     uint16 // Message ID
	opaqueId  uint8  // Opaque ID (meant to be sent back by device)
	version   byte   // Version
	sessionId byte   // Session ID
	seqNum    byte   // Sequence number
	dataLen   uint32 // Payload length
}

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

type SystemInfo struct {
	SystemInfo struct {
		UpdataType      string
		VideoInChannel  uint32
		AlarmOutChannel uint32
		AudioInChannel  uint32
		BuildTime       string
		ExtraChannel    uint32
		TalkInChannel   uint32
		UpdataTime      string
		CombineSwitch   uint32
		DeviceModel     string
		DeviceType      uint32
		EncryptVersion  string
		SerialNo        string
		AlarmInChannel  uint32
		DigChannel      uint32
		HardWareVersion string
		TalkOutChannel  uint32
		DeviceRunTime   string
		HardWare        string
		SoftWareVersion string
		VideoOutChannel uint32
	}

	Name      string
	Ret       uint32
	SessionID string
}

type SysAbilitiesData struct {
	Name           string
	Ret            float64
	SessionID      string
	SystemFunction struct {
		EncodeFunction struct {
			DoubleStream bool
			SmartH264    bool
			SmartH264V2  bool
			SnapStream   bool
		}

		NetServerFunction struct {
			NetPMSV2             bool
			NetAlarmCenter       bool
			NetDAS               bool
			NetEmail             bool
			NetFTP               bool
			NetNat               bool
			NetDDNS              bool
			NetIPFilter          bool
			NetNTP               bool
			NetWifi              bool
			IPAdaptive           bool
			Net3G                bool
			NetDHCP              bool
			NetMutlicast         bool
			OnvifPwdCheckout     bool
			WifiModeSwitch       bool
			WifiRouteSignalLevel bool
			NetDNS               bool
			NetPMS               bool
			NetPPPoE             bool
			NetRTSP              bool
			NetUPNP              bool
		}

		OtherFunction struct {
			SupportRPSVideo              bool
			SupportSetVolume             bool
			SupportTextPassword          bool
			SuppportChangeOnvifPort      bool
			SupportCfgCloudupgrade       bool
			SupportPWDSafety             bool
			SupportNetWorkMode           bool
			SupportPTZTour               bool
			SupportSetPTZPresetAttribute bool
			SupportTimingSleep           bool
			NOHDDRECORD                  bool
			SupportDoubleLightBoxCamera  bool
			SupportPtz360Spin            bool
			SupportShowH265X             bool
			SupportSoftPhotosensitive    bool
			SupportAlarmVoiceTips        bool
			SupportCloseVoiceTip         bool
			SupportMusicBulb433Pair      bool
			SupportOneKeyMaskVideo       bool
			SupportParkingGuide          bool
			SupportSetHardwareAbility    bool
			SupportCameraWhiteLight      bool
			SupportCommDataUpload        bool
			SupportTimeZone              bool
			SupportWebRTCModule          bool
			SupportDoubleLightBulb       bool
			SupportElectronicPTZ         bool
			SupportOSDInfo               bool
			SupportStatusLed             bool
			SupportWriteLog              bool
			SupportDNChangeByImage       bool
			SupportMailTest              bool
			SupportDimenCode             bool
			SupportMusicLightBulb        bool
			SupportSnapCfg               bool
			SupportBT                    bool
			SupportCloudUpgrade          bool
			SupportFTPTest               bool
			SupportAppBindFlag           bool
			SupportCamareStyle           bool
		}

		PreviewFunction struct {
			Talk bool
			Tour bool
		}

		TipShow struct {
			NoBeepTipShow bool
		}

		AlarmFunction struct {
			NetAbort          bool
			NetAlarm          bool
			StorageLowSpace   bool
			StorageNotExist   bool
			HumanDection      bool
			BlindDetect       bool
			HumanPedDetection bool
			LossDetect        bool
			MotionDetect      bool
			NetIpConflict     bool
			StorageFailure    bool
			AlarmConfig       bool
		}

		CommFunction struct {
			CommRS232 bool
			CommRS485 bool
		}
	}
}

// Autogenerated struct, don't modify manually!
type OEMInfo struct {
	Name    string
	OEMInfo struct {
		Address   string
		Name      string
		OEMID     uint32
		Telephone string
	}

	Ret       uint32
	SessionID string
}

type SysIPSetData struct {
	HostIP        string
	HttpPort      uint32
	MAC           string
	TCPPort       uint32
	TransferPlan  uint32
	HostName      string
	UseHSDownLoad bool
	TCPMaxConn    uint32
	DvrMac        string
	EncryptType   uint32
	GateWay       string
	MaxBps        uint32
	MonMode       uint32
	Submask       string
	Password      string
	SSLPort       uint32
	UDPPort       uint32
	Username      string
}
