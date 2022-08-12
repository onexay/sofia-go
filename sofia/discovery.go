package sofia

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
)

/*
Broadcast UDP sent from factory reset devices on port 34569
{
	"Name" : "NetWork.NetCommon",
	"NetWork.NetCommon" :
	{
		"DeviceType" : 43,
		"GateWay" : "0x0101a8c0",
		"HostIP" : "0x0a01a8c0",
		"HostName" : "IPC_6a6b",
		"HttpPort" : 80,
		"MAC" : "00:12:31:09:b7:9e",
		"MaxBps" : 0,
		"MonMode" : "TCP",
		"SN" : "a5b23b431b14712e",
		"SSLPort" : 8443,
		"Submask" : "0x00ffffff",
		"TCPMaxConn" : 10,
		"TCPPort" : 34567,
		"TransferPlan" : "Quality",
		"UDPPort" : 34568,
		"UseHSDownLoad" : false,
		"Version" : "V4.02.R12.E7335520.12012.047502.00000",
		"BuildDate" : "2018-09-17 10:47:30" ,
		"OtherFunction": "D=2022-08-12 11:14:35 V=0832fc42ad047dc"
	},
	"Ret" : 100,
	"SessionID" : "0x00000000"
}
*/

type Discovery struct {
	logger *logrus.Entry // Contextual logger
	addr   *net.UDPAddr  // Address
	conn   *net.UDPConn  // UDP connection
}

func NewDiscovery(port uint16) (*Discovery, error) {
	// Allocate a new discovery object
	discovery := new(Discovery)

	// Initialize logger
	{

	}

	// Initialize address
	{
		discovery.addr = new(net.UDPAddr)
		discovery.addr.IP = net.IPv4zero
		discovery.addr.Port = int(port)
	}

	return discovery, nil
}

func dumpJSON(name string, amap map[string]interface{}, level string) {
	olevel := level
	level = level + "  "

	if len(olevel) == 0 {
		fmt.Printf("// Autogenerated struct, don't modify manually!\n")
		fmt.Printf("type %s struct {\n", name)
	} else {
		fmt.Printf("%s {\n", olevel)
	}

	for k, v := range amap {
		fmt.Printf("%s%s ", level, string(k))
		switch v.(type) {
		case map[string]interface{}:
			fmt.Printf("struct\n")
			dumpJSON("", v.(map[string]interface{}), level)
		case float64:
			fmt.Printf("uint32")
		default:
			fmt.Printf("%T", v)

		}
		fmt.Printf("\n")
	}
	fmt.Printf("%s}\n", olevel)
}

func (discovery *Discovery) Start() {
	// Create a listener
	conn, err := net.ListenUDP("udp", discovery.addr)
	if err != nil {
		return
	}

	// Allocate buffer to receive data
	buf := make([]byte, 1500)

	for {
		// Read data
		rlen, raddr, err := conn.ReadFrom(buf)
		if err != nil {
			break
		}

		fmt.Printf("Rx message from [%s], length [%d]\n", raddr.String(), rlen)

		// Decode message header
		hdr := DecodeMessageHeader(buf[:DeviceMessageHeaderLen])

		fmt.Printf("%b 0x%X %d %d %d\n", hdr.version, hdr.sessionId, hdr.seqNum, hdr.msgId, hdr.dataLen)

		if hdr.dataLen > 0 {
			// Unmarshal message
			var x map[string]interface{}
			if err := json.Unmarshal(buf[DeviceMessageHeaderLen:DeviceMessageHeaderLen+hdr.dataLen-1], &x); err != nil {
				fmt.Printf(err.Error())
				continue
			}

			dumpJSON("X", x, "")
		}
	}

}
