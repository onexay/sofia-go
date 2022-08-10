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
	"github.com/teris-io/shortid"
)

/*
 *
 */
type Device struct {
	logger         *logrus.Logger      // Device scoped logger
	rxBuf          *bytes.Buffer       // Receive buffer
	txBuf          *bytes.Buffer       // Transmit buffer
	host           string              // Host, IP address or hostname
	port           string              // Port
	connectTimeout time.Duration       // Connect timeout
	connectRetries uint8               // Connect retries
	transport      net.Conn            // Transport connection
	idGen          *shortid.Shortid    // ID generator
	sessions       map[byte]string     // Mapping session IDs
	localSessions  map[string]*Session // Sessions
	workerChan     chan error          // Device worker channel
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

	// Initialize and setup device logger
	{
		newLogger := logrus.New()
		newLogger.SetFormatter(&logrus.JSONFormatter{})
		newLogger.SetOutput(os.Stdout)
		newLogger.SetLevel(logrus.DebugLevel)
		device.logger = newLogger
	}

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

	// Initialize sessions
	{
		device.idGen = shortid.GetDefault()
		device.sessions = make(map[byte]string, 16)
		device.localSessions = make(map[string]*Session, 16)
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
				device.logger.WithFields(
					logrus.Fields{
						"host": device.host + ":" + device.port,
					}).Debug("Connected successfully in %d tries", try)

				// Start worker
				go device.worker()

				break
			}

			device.logger.WithFields(
				logrus.Fields{
					"host": device.host + ":" + device.port,
				}).Debug("Unable to connect [%s], tried %d time(s)", err.Error(), try)
		}
	}

	return err
}

/*
 *
 */
func (device *Device) NewSession(user string, password string) *Session {
	// Get new session
	session := NewSession(device, user, password)

	// Generate a new local session id
	localId, _ := device.idGen.Generate()

	// Add an entry
	device.localSessions[localId] = session

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
			device.logger.WithFields(
				logrus.Fields{
					"host": device.host + ":" + device.port,
				}).Error("Connection closed [%s]", err.Error())

			break
		}

		// Decode message
		msg := device.DecodeMessage(buf)

		device.logger.WithFields(
			logrus.Fields{
				"host":  device.host + ":" + device.port,
				"msgId": msg.msgId,
			}).Debug("Received from transport")

		// Find session local index
		localSessionId, found := device.sessions[msg.sessionId]
		if !found {
			device.logger.WithFields(
				logrus.Fields{
					"host": device.host + ":" + device.port,
				}).Info("Unexpected session ID [0x%X]", msg.sessionId)

			continue
		}

		// Find local session
		session, found := device.localSessions[localSessionId]
		if !found {
			device.logger.WithFields(
				logrus.Fields{
					"host": device.host + ":" + device.port,
				}).Info("Local session [%s] not found for session ID [0x%X]", localSessionId, msg.sessionId)

			continue
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
		buf.WriteByte(0xFF)                             // Header flag, always 0xFF
		buf.WriteByte(msg.version)                      // Version, usually 0
		buf.Write([]byte{0x00, 0x00})                   // Reserved field 1,2
		buf.WriteByte(msg.sessionId)                    // Session ID
		buf.Write([]byte{0x00, 0x00, 0x00})             // Unknown field 1
		buf.WriteByte(msg.seqNum)                       // Sequence number
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00}) // Unknown field 2
		buf.Write(encMsgId)                             // Message ID
		buf.Write(encDataLen)                           // Data length
		buf.Write(msg.data)                             // Data
		buf.Write([]byte{0x0A, 0x00})                   // Message trailer, always 0x0A,0x00
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

	device.logger.WithFields(
		logrus.Fields{
			"host": device.host + ":" + device.port,
		}).Debug("Wrote [%d] bytes to transport", writeLen)

	return err
}
