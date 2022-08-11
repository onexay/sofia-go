package main

import (
	"fmt"
	"sofia-go/sofia"
	"sync"
)

func main() {
	wg := sync.WaitGroup{}

	go func(pwg *sync.WaitGroup) {
		fmt.Printf("Starting task 1 ...\n")
		pwg.Add(1)

		// Create a new device
		device1 := sofia.NewDevice("192.168.31.177", "34567", 5, 5)
		device1.Connect()
		session1 := device1.NewSession("admin", "")
		session1.Login()
		session1.SysInfo()
		session1.SysAbilities()
		<-*device1.WorkerChan()

		pwg.Done()
	}(&wg)

	go func(pwg *sync.WaitGroup) {
		fmt.Printf("Starting task 2 ...\n")
		pwg.Add(1)

		// Create another device
		device2 := sofia.NewDevice("192.168.31.156", "34567", 5, 5)
		device2.Connect()
		session2 := device2.NewSession("admin", "")
		session2.Login()
		session2.SysInfo()
		session2.SysAbilities()
		<-*device2.WorkerChan()

		pwg.Done()
	}(&wg)

	fmt.Printf("Waiting for other tasks to complete...")
	wg.Wait()
	fmt.Scanln()
}

/*
func main() {
	// Create a logger
	var log = logrus.New()

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

	fmt.Printf("New session is %s (%d)\n", *session.IDStr(), session.ID())

	var sysInfo sofia.SysInfo
	sysInfo, _ = device.SystemInfo(session)

	fmt.Printf("%s %s %s\n", sysInfo.SystemInfo.SerialNo, sysInfo.SystemInfo.BuildTime, sysInfo.SystemInfo.SoftWareVersion)

	//==============================

	var session2 *sofia.Session

	device2 := sofia.NewDevice()

	// Connect to a device, all sessions to device share the same connection
	if err = device2.Connect("192.168.31.156", "34567"); err != nil {
		fmt.Printf("Unable to connect [%s]\n", err.Error())
		return
	}

	// Login to device, a new login generates a new session
	if session2, err = device2.Login("", ""); err != nil {
		fmt.Printf("Unable to login [%s]\n", err.Error())
		return
	}

	if session2 == nil {
		fmt.Printf("Error\n")
	}

	fmt.Printf("New session is %s (%d)\n", *session2.IDStr(), session2.ID())

	var sysInfo2 sofia.SysInfo
	sysInfo2, _ = device2.SystemInfo(session)

	fmt.Printf("%s %s %s\n", sysInfo2.SystemInfo.SerialNo, sysInfo2.SystemInfo.BuildTime, sysInfo2.SystemInfo.SoftWareVersion)

	//var sysAbility sofia.SysAbility
	//device.SystemAbility(session)

	//var sysOEmInfo sofia.SysOEMInfo
	//device.SystemOEMInfo(session)

	//var sysConfig sofia.SysConfig
	//device.SystemConfig(session, "NetWork")

	fmt.Scanln()
}
*/
