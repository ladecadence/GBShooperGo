package flashcart

import (
	"errors"
	"fmt"

	"github.com/ladecadence/GBShooperGo/pkg/comms"
)

const (
	GBS_ID    = 0x17 // 23 decimal
	SLEEPTIME = 3
)

type Status struct {
	VersionMayor uint8
	VersionMinor uint8
}

func GBSStatus() (Status, error) {
	status := Status{}
	gbs := comms.GBSDevice{}
	err := gbs.Open()
	if err != nil {
		return Status{}, err
	}
	defer gbs.Close()
	gbs.Dev.PurgeReadBuffer()

	// create packet
	packet := comms.Packet{Type: comms.TYPE_INFO, Data: 0x00}
	// send it
	gbs.SendPacket(packet)

	// read answer (3 packets)
	packet, err = gbs.ReceivePacket(SLEEPTIME)
	if err != nil {
		return Status{}, err
	}
	id := packet.Data
	ty := packet.Type
	packet, err = gbs.ReceivePacket(SLEEPTIME)
	if err != nil {
		return Status{}, err
	}
	status.VersionMayor = packet.Data
	packet, err = gbs.ReceivePacket(SLEEPTIME)
	if err != nil {
		return Status{}, err
	}
	status.VersionMinor = packet.Data

	// checks
	if id != GBS_ID {
		fmt.Printf("Type: %x\n", ty)
		fmt.Printf("ID: %x\n", id)
		return Status{}, errors.New("Bad GBShooper ID")
	}

	// ok
	return status, nil
}
