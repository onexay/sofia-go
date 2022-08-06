package sofia

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

var msgEnd []byte = []byte{0x0A, 0x00}

/*
 *
 */
type Device struct {
	host     string           // Hostname or IP address
	port     string           // Port
	proto    string           // Protocol
	conn     net.Conn         // Network connection
	username string           // Username
	password string           // Password
	reqData  *bytes.Buffer    // Request data buffer
	resData  *bytes.Buffer    // Response data buffer
	sessions map[byte]Session // Sessions
}

/*
 *
 */
func NewDevice(host string, port string, proto string, username string, password string) *Device {
	device := new(Device)

	device.host = host
	device.port = port
	device.proto = proto

	if len(username) == 0 {
		device.username = "admin"
	} else {
		device.username = username
	}

	if len(password) == 0 {
		device.password = "tlJwpbo6"
	} else {
		device.password = password
	}

	device.reqData = new(bytes.Buffer)

	device.resData = new(bytes.Buffer)

	device.sessions = make(map[byte]Session)

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
func ReadDevice(device *Device) {
	tmp := make([]byte, 2)

	for {
		_, err := device.conn.Read(tmp)

		if err != nil {
			if err == io.EOF {
				fmt.Printf("Device [%s] closed connection.\n", device.conn.RemoteAddr().String())
				break
			}
		}

		device.resData.Write(tmp)

		if tmp[0] == msgEnd[0] && tmp[1] == msgEnd[1] {
			fmt.Printf("%s\n", hex.Dump(device.resData.Bytes()))
			//msg := device.resData.Bytes()[20 : len(device.resData.Bytes())-2]
			device.resData.Reset()
			tmp = []byte{0x00, 0x00}
		}
	}
}

/*
 *
 */
func (device *Device) Connect() error {
	var err error = nil

	device.conn, err = net.Dial(device.proto, device.host+":"+device.port)

	return err
}

/*
 *
 */
func (device *Device) Disconnect() {
	device.conn.Close()
}

/*
 *
 */
func (device *Device) Login() {
	loginData := LoginMsg{
		EncryptType: "MD5",
		LoginType:   "Sofia-Go",
		UserName:    device.username,
		PassWord:    device.password,
	}

	data, _ := json.Marshal(loginData)

	MakeMessage(device.reqData, data, 1000)

	fmt.Printf("Login message\n%s\n", hex.Dump(device.reqData.Bytes()))

	_, err := device.conn.Write(device.reqData.Bytes())

	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
}

/*
 *
 */
func (device *Device) Logout() {

}
