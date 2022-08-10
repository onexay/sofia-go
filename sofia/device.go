package sofia

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

/*
 *
 */
type Device struct {
	logger         *logrus.Entry // Device scoped logger
	rxBuf          *bytes.Buffer // Receive buffer
	txBuf          *bytes.Buffer // Transmit buffer
	host           string        // Host, IP address or hostname
	port           string        // Port
	connectTimeout time.Duration // Connect timeout
	connectRetries uint8         // Connect retries
	transport      net.Conn      // Transport connection
	sequence       *Sequence     // Sequence of indices
	sessions       []*Session    // Actual sessions
	tmpSessions    []*Session    // Temporary sessions (discarded after LOGIN_RSP)
	workerChan     chan error    // Device worker channel
}

/*
 *
 */
func (device *Device) WorkerChan() *chan error {
	return &device.workerChan
}

/*
 *
 */
func NewDevice(host string, port string, timeout uint16, retries uint8) *Device {
	// Allocate a new device
	var device *Device = new(Device)

	// Allocate Rx and Tx buffers
	{
		device.rxBuf = new(bytes.Buffer)
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
		newLogger.SetFormatter(&logrus.JSONFormatter{})
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
		device.workerChan = make(chan error)
	}

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
	// Make a reader
	reader := bufio.NewReader(device.transport)

	for {
		// Begin reading
		buf, err := reader.ReadBytes(0x0A)
		if err != nil && err == io.EOF {
			device.logger.Error("Connection closed [%s]", err.Error())
			//device.workerChan <- err
			break
		}

		// Decode message
		msg := device.DecodeMessage(buf)

		device.logger.Debug("Read ", len(buf), " bytes from transport")

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

		// Send message to session
		session.workerBus <- msg
	}
}

/*
 *
 */
func (device *Device) EncodeMessage(msg *DeviceMessage) {
	// Always reset the Tx buffer
	device.txBuf.Reset()

	// Encode message
	{
		encMsgId := make([]byte, 2)
		encDataLen := make([]byte, 4)
		binary.LittleEndian.PutUint16(encMsgId, uint16(msg.msgId))
		binary.LittleEndian.PutUint32(encDataLen, msg.dataLen)

		buf := device.txBuf
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
func (device *Device) DecodeMessage(buf []byte) DeviceMessage {
	// Decode message
	var msg DeviceMessage
	{
		msg.version = buf[DeviceMessageOffsetVersion]                              // Version
		msg.sessionId = buf[DeviceMessageOffsetSessionId]                          // Session ID
		msg.seqNum = buf[DeviceMessageOffsetSeqNum]                                // Sequence number
		msg.msgId = binary.LittleEndian.Uint16(buf[DeviceMessageOffsetMsgId:])     // Message ID
		msg.dataLen = binary.LittleEndian.Uint32(buf[DeviceMessageOffsetDataLen:]) // Data length
		msg.data = buf[20 : len(buf)-DeviceMessageTrailerLen]                      // Truncate the message trailer

		// Extract login correlation id
		if msg.msgId == LOGIN_RSP {
			msg.opaqueId = buf[DeviceMessageOffsetOpaqueId] // Opaque ID
		}
	}

	return msg
}

/*
 *
 */
func (device *Device) SendMessage(msg *DeviceMessage) error {
	// Encode message
	device.EncodeMessage(msg)

	// Send message
	writeLen, err := device.transport.Write(device.txBuf.Bytes())

	device.logger.Debug("Wrote ", writeLen, " bytes to transport")

	return err
}
