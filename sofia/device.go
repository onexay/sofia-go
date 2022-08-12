package sofia

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

/*
 *
 */
type Device struct {
	logger         *logrus.Entry   // Device scoped logger
	txBuf          *bytes.Buffer   // Transmit buffer
	host           string          // Host, IP address or hostname
	port           string          // Port
	connectTimeout time.Duration   // Connect timeout
	connectRetries uint8           // Connect retries
	transport      net.Conn        // Transport connection
	sequence       *Sequence       // Sequence of indices
	sessions       []*Session      // Actual sessions
	tmpSessions    []*Session      // Temporary sessions (discarded after LOGIN_RSP)
	wg             *sync.WaitGroup // Wait groups for sessions
	rxChan         chan error      // Device receive channel
}

/*
 *
 */
func (device *Device) WorkerChan() *chan error {
	return &device.rxChan
}

/*
 *
 */
func NewDevice(host string, port string, timeout uint16, retries uint8) (*Device, error) {
	// Allocate a new device
	var device *Device = new(Device)

	// Allocate Tx buffers
	{
		device.txBuf = new(bytes.Buffer)
	}

	// Setup connection parameters
	{
		device.host = host
		device.port = port
		device.connectTimeout = time.Second * time.Duration(timeout)
		device.connectRetries = retries
	}

	// Initialize and setup device logger
	{
		newLogger := logrus.New()
		newLogger.SetFormatter(&logrus.TextFormatter{})
		newLogger.SetOutput(os.Stdout)
		newLogger.SetLevel(logrus.DebugLevel)
		device.logger = newLogger.WithFields(logrus.Fields{"remote": device.host + ":" + device.port})
	}

	// Initialize sessions
	{
		device.sequence = NewSequence(0xFF)
		device.sessions = make([]*Session, 0xFF)
		device.tmpSessions = make([]*Session, 0xFF)
	}

	// Setup worker channel
	{
		device.wg = new(sync.WaitGroup)
		device.rxChan = make(chan error)
	}

	return device, nil
}

/*
 *
 */
func DeleteDevice(device *Device) {

}

/*
 *
 */
func (device *Device) Connect() error {
	// Try to connect to device
	var err error
	{
		for try := 1; try <= int(device.connectRetries); try++ {
			if device.transport, err = net.DialTimeout("tcp", device.host+":"+device.port, device.connectTimeout); err == nil {
				device.logger.Debug("Connected successfully in try ", try)

				// Start worker
				go device.worker()

				break
			}

			device.logger.Debug("Unable to connect, try ", err.Error(), try)
		}
	}

	return err
}

/*
 *
 */
func (device *Device) NewSession(user string, password string) *Session {
	// Generate a new local session id
	localId := device.sequence.GetIndex()
	if localId == 0 {
		return nil
	}

	// Get new session
	session := NewSession(device, localId, user, password)

	device.logger.Info("Created new session with local ID ", localId)

	// Add an entry
	device.tmpSessions[localId] = session

	return session
}

/*
 *
 */
func (device *Device) DeleteSession() {

}

/*
 *
 */
func (device *Device) worker() {
	for {
		// Read message header
		hbuf := make([]byte, DeviceMessageHeaderLen)
		hrlen, err := io.ReadFull(device.transport, hbuf)
		if err != nil {
			device.logger.Error("Connection closed [reading header] ", err.Error())
			break
		}

		// Decode message header
		hdr := DecodeMessageHeader(hbuf)

		// Read rest of message
		dbuf := make([]byte, hdr.dataLen)
		drlen, err := io.ReadFull(device.transport, dbuf)
		if err != nil {
			device.logger.Error("Connection closed [reading data] ", err.Error())
			break
		}

		device.logger.Debug("Rx message [", hdr.msgId, "], length [", hrlen+drlen, "]")

		msg := DeviceMessage{
			msgId:     hdr.msgId,
			opaqueId:  hdr.opaqueId,
			version:   hdr.version,
			sessionId: hdr.sessionId,
			seqNum:    hdr.seqNum,
			dataLen:   hdr.dataLen,
			data:      dbuf[:hdr.dataLen-DeviceMessageTrailerLen],
		}

		// All messages for device require a valid session ID that we receive
		// in LOGIN_RSP. Since we create a temporary session even before we have
		// this session ID, we require it to be mapped to our internal ID.

		var session *Session
		{
			if msg.msgId == LOGIN_RSP {
				// Find session using the internal id
				if session = device.tmpSessions[msg.opaqueId]; session == nil {
					device.logger.Info("Unexpected local session ID ", msg.opaqueId)
					continue
				}

				// Update actual session ID, session and release temporary session
				{
					session.id = msg.sessionId
					device.sessions[session.id] = session
					device.tmpSessions[msg.opaqueId] = nil
					device.sequence.FreeIndex(msg.opaqueId)
				}
			} else {
				// Find session using device session id
				if session = device.sessions[msg.sessionId]; session == nil {
					device.logger.Info("Unexpected device session ID ", msg.sessionId)
					continue
				}
			}
		}

		// Increment sequence number
		session.seqNum = session.seqNum + 1

		// Send message to session
		session.rxChan <- msg
	}
}

/*
 *
 */
func (device *Device) SendMessage(msg *DeviceMessage) error {
	// Always reset the Tx buffer
	device.txBuf.Reset()

	// Encode message
	EncodeMessage(msg, device.txBuf)

	// Send message
	writeLen, err := device.transport.Write(device.txBuf.Bytes())

	device.logger.Debug("Tx message [", msg.msgId, "], length [", writeLen, "]")

	return err
}

/*
 *
 */
func EncodeMessage(msg *DeviceMessage, bytesBuf *bytes.Buffer) {
	// Encode message
	{
		encMsgId := make([]byte, 2)
		encDataLen := make([]byte, 4)
		binary.LittleEndian.PutUint16(encMsgId, uint16(msg.msgId))
		binary.LittleEndian.PutUint32(encDataLen, msg.dataLen)

		buf := bytesBuf
		buf.WriteByte(0xFF)                 // Header flag, always 0xFF
		buf.WriteByte(msg.version)          // Version, usually 0
		buf.Write([]byte{0x00, 0x00})       // Reserved field 1,2
		buf.WriteByte(msg.sessionId)        // Session ID
		buf.Write([]byte{0x00, 0x00, 0x00}) // Unknown field 1
		buf.WriteByte(msg.seqNum)           // Sequence number

		// Use 2 bytes of Unknown field 2 to correlate local session with login
		if msg.msgId == LOGIN_REQ2 {
			buf.WriteByte(msg.opaqueId) // Unknown field 2 (first byte)
		} else {
			buf.Write([]byte{0x00}) // Unknown field 2 (first byte)
		}

		buf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // Unknown field 2 (last 4 bytes)
		buf.Write(encMsgId)                       // Message ID
		buf.Write(encDataLen)                     // Data length
		buf.Write(msg.data)                       // Data
		buf.Write([]byte{0x0A, 0x00})             // Message trailer, always 0x0A,0x00
	}
}

/*
 *
 */
func DecodeMessageHeader(buf []byte) DeviceMessageHeader {
	// Decode message
	var hdr DeviceMessageHeader
	{
		hdr.version = buf[DeviceMessageOffsetVersion]                              // Version
		hdr.sessionId = buf[DeviceMessageOffsetSessionId]                          // Session ID
		hdr.seqNum = buf[DeviceMessageOffsetSeqNum]                                // Sequence number
		hdr.msgId = binary.LittleEndian.Uint16(buf[DeviceMessageOffsetMsgId:])     // Message ID
		hdr.dataLen = binary.LittleEndian.Uint32(buf[DeviceMessageOffsetDataLen:]) // Data length

		// Extract login correlation id
		if hdr.msgId == LOGIN_RSP {
			hdr.opaqueId = buf[DeviceMessageOffsetOpaqueId] // Opaque ID
		}
	}

	return hdr
}

/*
 *
 */
func DecodeMessage(buf []byte) DeviceMessage {
	// Decode message
	var msg DeviceMessage
	{
		msg.version = buf[DeviceMessageOffsetVersion]                              // Version
		msg.sessionId = buf[DeviceMessageOffsetSessionId]                          // Session ID
		msg.seqNum = buf[DeviceMessageOffsetSeqNum]                                // Sequence number
		msg.msgId = binary.LittleEndian.Uint16(buf[DeviceMessageOffsetMsgId:])     // Message ID
		msg.dataLen = binary.LittleEndian.Uint32(buf[DeviceMessageOffsetDataLen:]) // Data length
		msg.data = buf[DeviceMessageOffsetData:]                                   // Truncate the message trailer

		// Extract login correlation id
		if msg.msgId == LOGIN_RSP {
			msg.opaqueId = buf[DeviceMessageOffsetOpaqueId] // Opaque ID
		}
	}

	return msg
}
