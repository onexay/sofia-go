package main

import (
	"fmt"
	"sofia-go/sofia"
)

func main() {
	var err error
	var session *sofia.Session

	device := sofia.NewDevice()

	// Connect to a device, all sessions to device share the same connection
	if err = device.Connect("192.168.31.177", "34567"); err != nil {
		fmt.Printf("Unable to connect [%s]\n", err.Error())
		return
	}

	// Login to device, a new login generates a new session
	if session, err = device.Login("", ""); err != nil {
		fmt.Printf("Unable to login [%s]\n", err.Error())
		return
	}

	if session == nil {
		fmt.Printf("Error\n")
	}

	fmt.Printf("New session is %s(%d)\n", *session.IDStr(), session.ID())

	var sysInfo sofia.SysInfo
	sysInfo, _ = device.SystemInfo(session)

	fmt.Printf("%s %s %s\n", sysInfo.SystemInfo.SerialNo, sysInfo.SystemInfo.BuildTime, sysInfo.SystemInfo.SoftWareVersion)

	//var sysAbility sofia.SysAbility
	device.SystemAbility(session)

	//var sysOEmInfo sofia.SysOEMInfo
	device.SystemOEMInfo(session)

	//var sysConfig sofia.SysConfig
	device.SystemConfig(session, "System")

	fmt.Scanln()
}
