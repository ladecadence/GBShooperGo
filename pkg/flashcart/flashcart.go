package flashcart

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/ladecadence/GBShooperGo/pkg/comms"
)

const (
	GBS_ID      = 0x17 // 23 decimal
	SLEEPTIME   = 3
	ERASETIME   = 60
	BUFFER_SIZE = 256

	// status
	STAT_OK      = 0x14 // 10.4 ;-)
	STAT_ERROR   = 0xEE
	STAT_TIMEOUT = 0xAA

	// Sizes
	S_0K    = 0
	S_2K    = 2048
	S_8K    = 8192
	S_32K   = 32768
	S_64K   = 65536
	S_128K  = 131072
	S_256K  = 262144
	S_512K  = 524288
	S_1MB   = 1048576
	S_2MB   = 2097152
	S_4MB   = 4194304
	S_1_1MB = 1179648
	S_1_2MB = 1310720
	S_1_5MB = 1572864
)

type Status struct {
	VersionMayor uint8
	VersionMinor uint8
}

type FlashID struct {
	ManufacturerID uint8
	ChipID         uint8
	Manufacturer   string
	Chip           string
}

type RomHeader struct {
	Title    string
	Cart     string
	CartType uint8
	ROMSize  uint8
	RAMSize  uint8
	ROM      string
	RAM      string
	ROMBytes int
	RAMBytes int
}

type FlashProducer struct {
	ID   uint8
	Name string
}

type FlashNames struct {
	ID   uint8
	Name string
}

type CartType struct {
	ID   uint8
	Type string
}

type ROMSize struct {
	ID   uint8
	Name string
	Size int
}

type RAMSize struct {
	ID   uint8
	Name string
	Size int
}

// flash chip producers
var FlashProducers = []FlashProducer{
	{0x01, "AMD"}, {0x02, "AMI"}, {0xe5, "Analog Devices"},
	{0x1f, "Atmel"}, {0x31, "Catalyst"}, {0x34, "Cypress"},
	{0x04, "Fujitsu"}, {0xE0, "Goldstar"}, {0x07, "Hitachi"},
	{0xad, "Hyundai"}, {0xc1, "Infineon"}, {0x89, "Intel"},
	{0xd5, "Intg. Silicon Systems"}, {0xc2, "Macronix"}, {0x29, "Microchip"},
	{0x2c, "Micron"}, {0x1c, "Mitsubishi"}, {0x10, "Nec"},
	{0x15, "Philips Semiconductors"}, {0xce, "Samsung"}, {0x62, "Sanyo"},
	{0x20, "SGS Thomson"}, {0xb0, "Sharp"}, {0xbf, "SST"},
	{0x97, "Texas Instruments"}, {0x98, "Toshiba"}, {0xda, "Winbond"},
	{0x19, "Xicor"}, {0xc9, "Xilinx"},
}

// flash chip IDs
var ChipIDs = []FlashNames{
	{0xA4, "29F040B"}, {0xAD, "AM29F016"},
}

// Cartridge types
var CartTypes = []CartType{
	{0x00, "ROM ONLY"}, {0x01, "ROM+MBC1"},
	{0x02, "ROM+MBC1+RAM"}, {0x03, "ROM+MBC1+RAM+BATT"},
	{0x05, "ROM+MBC2"}, {0x06, "ROM+MBC2+BATTERY"},
	{0x08, "ROM+RAM"}, {0x09, "ROM+RAM+BATTERY"},
	{0x11, "ROM+MBC3"},
	{0x0b, "ROM+MMMO1"}, {0x0c, "ROM+MMMO1+SRAM"},
	{0x0d, "ROM+MMMO1+SRAM+BATT"}, {0x0f, "ROM+MBC3+TIMER+BATT"},
	{0x10, "ROM+MBC3+TIMER+RAM+BAT"}, {0x12, "ROM+MBC3+RAM"},
	{0x13, "ROM+MBC3+RAM+BATT"}, {0x19, "ROM+MBC5"},
	{0x1a, "ROM+MBC5+RAM"}, {0x1b, "ROM+MBC5+RAM+BATT"},
	{0x1c, "ROM+MBC5+RUMBLE"}, {0x1d, "ROM+MBC5+RUMBLE+SRAM"},
	{0x1e, "ROM+MBC5+RUMBLE+SRAM+BATT"}, {0x1f, "Pocket Camera"},
	{0xfd, "Bandai TAMA5"}, {0xfe, "Hudson HuC-3"},
}

// ROM Sizes
var ROMSizes = []ROMSize{
	{0x00, "32KB", S_32K}, {0x01, "64KB", S_64K}, {0x02, "128KB", S_128K},
	{0x03, "256KB", S_256K}, {0x04, "512KB", S_512K}, {0x05, "1MB", S_1MB},
	{0x06, "2MB", S_2MB}, {0x07, "4MB", S_4MB}, {0x52, "1.1MB", S_1_1MB},
	{0x53, "1.2MB", S_1_2MB}, {0x54, "1.5MB", S_1_5MB},
}

// RAM sizes
var RAMSizes = []RAMSize{
	{0x00, "0KB", S_0K}, {0x01, "2KB", S_2K}, {0x02, "8KB", S_2K},
	{0x03, "32KB", S_32K}, {0x04, "128KB", S_128K},
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

func GBSChipID() (FlashID, error) {
	id := FlashID{}
	gbs := comms.GBSDevice{}
	err := gbs.Open()
	if err != nil {
		return FlashID{}, err
	}
	defer gbs.Close()
	gbs.Dev.PurgeReadBuffer()

	// create packet
	packet := comms.Packet{Type: comms.TYPE_COMMAND, Data: comms.CMD_ID}
	// send it
	gbs.SendPacket(packet)

	// read answer (2 packets)
	packet, err = gbs.ReceivePacket(SLEEPTIME)
	if err != nil {
		return FlashID{}, err
	}

	id.ManufacturerID = packet.Data

	packet, err = gbs.ReceivePacket(SLEEPTIME)
	if err != nil {
		return FlashID{}, err
	}
	id.ChipID = packet.Data

	if idx := slices.IndexFunc(FlashProducers, func(c FlashProducer) bool { return c.ID == id.ManufacturerID }); idx != -1 {
		id.Manufacturer = FlashProducers[idx].Name
	} else {
		id.Manufacturer = fmt.Sprintf("Unknown manufacurer: 0x%0x", id.ManufacturerID)
	}

	if idx := slices.IndexFunc(ChipIDs, func(c FlashNames) bool { return c.ID == id.ChipID }); idx != -1 {
		id.Chip = ChipIDs[idx].Name
	} else {
		id.Chip = fmt.Sprintf("Unknown Flash chip: 0x%0x", id.ChipID)
	}

	// ok
	return id, nil
}

func GBSReadHeader() (RomHeader, error) {
	header := RomHeader{}
	gbs := comms.GBSDevice{}
	err := gbs.Open()
	if err != nil {
		return RomHeader{}, err
	}
	defer gbs.Close()
	gbs.Dev.PurgeReadBuffer()

	// create packet
	packet := comms.Packet{Type: comms.TYPE_COMMAND, Data: comms.CMD_READ_HEADER}
	// send it
	gbs.SendPacket(packet)

	// read answer ( first 3 packets)
	// pkt1 = mapper, pkt2 = rom size, pkt3 = ram_size
	packet, err = gbs.ReceivePacket(SLEEPTIME)
	if err != nil {
		return RomHeader{}, err
	}
	header.CartType = packet.Data

	packet, err = gbs.ReceivePacket(SLEEPTIME)
	if err != nil {
		return RomHeader{}, err
	}
	header.ROMSize = packet.Data

	packet, err = gbs.ReceivePacket(SLEEPTIME)
	if err != nil {
		return RomHeader{}, err
	}
	header.RAMSize = packet.Data

	// now read cart name (16 bytes)
	for range 16 {
		packet, err = gbs.ReceivePacket(SLEEPTIME)
		if err != nil {
			return RomHeader{}, err
		}
		header.Title += string(packet.Data)
	}

	// fill types
	// fill types
	if idx := slices.IndexFunc(CartTypes, func(c CartType) bool { return c.ID == header.CartType }); idx != -1 {
		header.Cart = CartTypes[idx].Type
	} else {
		header.Cart = "Unknown cart type"
	}

	if idx := slices.IndexFunc(ROMSizes, func(c ROMSize) bool { return c.ID == header.ROMSize }); idx != -1 {
		header.ROMBytes = ROMSizes[idx].Size
		header.ROM = ROMSizes[idx].Name
	} else {
		header.ROMBytes = 0
		header.ROM = "Unknown ROM size"
	}

	// fill types
	if idx := slices.IndexFunc(RAMSizes, func(c RAMSize) bool { return c.ID == header.RAMSize }); idx != -1 {
		header.RAMBytes = RAMSizes[idx].Size
		header.RAM = RAMSizes[idx].Name
	} else {
		header.RAMBytes = 0
		header.RAM = "Unknown RAM size"
	}

	// ok
	return header, nil
}

func GBSEraseFlash() error {
	gbs := comms.GBSDevice{}
	err := gbs.Open()
	if err != nil {
		return err
	}
	defer gbs.Close()
	gbs.Dev.PurgeReadBuffer()

	// create packet
	packet := comms.Packet{Type: comms.TYPE_COMMAND, Data: comms.CMD_ERASE_FLASH}
	// send it
	gbs.SendPacket(packet)

	// read answer
	packet, err = gbs.ReceivePacket(ERASETIME)
	if err != nil {
		return err
	}
	if packet.Data == STAT_OK {
		return nil
	} else {
		return errors.New("Error erasing flash")
	}
}

func GBSWriteFlash(filename string, finished chan bool, progress chan int64) error {
	// finishing
	defer func() { finished <- true }()

	// open rom file
	rom, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer rom.Close()

	// get file size
	stats, err := rom.Stat()
	if err != nil {
		return err
	}
	romSize := stats.Size()

	// open GBShooper
	gbs := comms.GBSDevice{}
	err = gbs.Open()
	if err != nil {
		return err
	}
	defer gbs.Close()
	gbs.Dev.PurgeReadBuffer()

	// and start writing
	var chunkCounter int64 = 0
	buffer := make([]byte, BUFFER_SIZE)

	// create packet
	packet := comms.Packet{Type: comms.TYPE_COMMAND, Data: comms.CMD_PRG_FLASH}
	// send it
	gbs.SendPacket(packet)

	stat, err := gbs.ReceivePacket(SLEEPTIME)
	if err != nil {
		packet := comms.Packet{Type: comms.TYPE_COMMAND, Data: comms.CMD_END}
		// send it
		gbs.SendPacket(packet)
		return err
	}
	if stat.Data == STAT_OK {
		for {
			// calculate percentage
			percent := (100 * chunkCounter * BUFFER_SIZE) / romSize
			progress <- percent

			// read a chunk
			_, err := rom.Read(buffer)
			if err != nil {
				// end?
				if err == io.EOF {
					break
				}

			}
			// checksum
			var check uint8 = 0
			for i := range BUFFER_SIZE {
				check += buffer[i]
			}

			// send the data
			err = gbs.SendBuffer(buffer)
			// get answer
			stat, err := gbs.ReceivePacket(SLEEPTIME)
			// checksum correct?
			if stat.Data != check {
				packet := comms.Packet{Type: comms.TYPE_COMMAND, Data: comms.CMD_END}
				// send it
				gbs.SendPacket(packet)
				return errors.New("Bad checksum")
			}

			// EOF?
			c := make([]uint8, 1)
			_, err = rom.Read(c)
			if err != io.EOF {
				// ok, keep writing
				rom.Seek(-1, 1)
				packet := comms.Packet{Type: comms.TYPE_COMMAND, Data: comms.CMD_PRG_FLASH}
				// send it
				gbs.SendPacket(packet)
				chunkCounter++
			}

		}
	} else {
		packet := comms.Packet{Type: comms.TYPE_COMMAND, Data: comms.CMD_END}
		// send it
		gbs.SendPacket(packet)
		return errors.New("Problem with hardware, can't write to Flash")
	}

	// end
	packet = comms.Packet{Type: comms.TYPE_COMMAND, Data: comms.CMD_END}
	// send it
	gbs.SendPacket(packet)
	return nil
}

func GBSReadFlash(filename string, size int64, finished chan bool, progress chan int64, errchan chan error) error {
	// finishing
	defer func() { finished <- true }()

	// open rom file
	rom, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer rom.Close()

	// open GBShooper
	gbs := comms.GBSDevice{}
	err = gbs.Open()
	if err != nil {
		return err
	}
	defer gbs.Close()
	gbs.Dev.PurgeReadBuffer()

	// start reading
	chunks := size / BUFFER_SIZE
	buffer := make([]byte, BUFFER_SIZE)
	packet := comms.Packet{Type: comms.TYPE_COMMAND, Data: comms.CMD_READ_FLASH}
	// send it
	gbs.SendPacket(packet)

	for n := range chunks {
		// calculate progress
		percent := n * BUFFER_SIZE * 100 / size
		progress <- percent

		// read buffer and calculate checksum
		var check uint8 = 0
		for i := range BUFFER_SIZE {
			buffer[i], err = gbs.ReceiveByte(SLEEPTIME)
			if err != nil {
				errchan <- err
				return err
			}
			check += buffer[i]
		}
		// write buffer in file
		rom.Write(buffer)

		// send checksum
		packet = comms.Packet{Type: comms.TYPE_DATA, Data: check}
		gbs.SendPacket(packet)

		// read answer
		stat, err := gbs.ReceivePacket(SLEEPTIME)
		if err != nil {
			errchan <- err
			return err
		}
		// cheksum bad?
		if stat.Data == comms.CMD_END {
			errchan <- errors.New("Bad checksum")
			return errors.New("Bad checksum")
		}

		// ok, continue
		if n < chunks-1 {
			packet = comms.Packet{Type: comms.TYPE_COMMAND, Data: comms.CMD_READ_FLASH}
			gbs.SendPacket(packet)
		}
	}

	// finished
	packet = comms.Packet{Type: comms.TYPE_COMMAND, Data: comms.CMD_END}
	gbs.SendPacket(packet)

	finished <- true
	return nil
}
