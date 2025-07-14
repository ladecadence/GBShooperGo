package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/schollz/progressbar/v3"

	"github.com/ladecadence/GBShooperGo/pkg/color"
	"github.com/ladecadence/GBShooperGo/pkg/flashcart"
)

const (
	VER_MAYOR = 0
	VER_MINOR = 2
)

func GBSHelp() {
	fmt.Println(color.Green + "⚡ GBShooperGo version: " + color.Purple + strconv.Itoa(VER_MAYOR) + "." + strconv.Itoa(VER_MINOR) + color.Reset)
	fmt.Println("David Pello 2025")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("gbshoopergo <action> <options> [file]")
	fmt.Println()
	fmt.Println("Actions:")
	fmt.Println("\t --version: prints the software version.")
	fmt.Println("\t --status: checks the hardware.")
	fmt.Println("\t --id: gets the ID of the flash chip.")
	fmt.Println("\t --read-header: gets header information, mapper and RAM/ROM sizes.")
	fmt.Println("\t --erase-flash: clears the contents of the flash chip.")
	fmt.Print("\t --read-flash: reads the contents of the flash chip ")
	fmt.Println("and writes it on [file].")
	fmt.Println("\t\toptions:")
	fmt.Println("\t\t  --size N: Specify ROM size:")
	fmt.Print("\t\t\t 1=32KB, 2=64KB, 3=128KB, 4=256KB, 5=512KB, 6=1MB, ")
	fmt.Println("7=2MB, 8=4MB")
	fmt.Println("\t\t If no size is specified, 32KB are read")
	fmt.Println("\t --write-flash: writes the flash with contents from [file].")
	fmt.Print("\t --read-ram: reads the contents of the save RAM ")
	fmt.Println("and writes it on [file].")
	fmt.Println("\t\toptions:")
	fmt.Println("\t\t  --size N: Specify RAM size:")
	fmt.Println("\t\t\t 1=8KB, 2=32KB, 3=1MB")
	fmt.Println("\t\t If no size is specified, 8KB are read")
	fmt.Println("\t --write-ram: writes the save RAM with contents from [file].")
	fmt.Println("\t --erase-ram: clears the contents of the save RAM with 0's.")
	fmt.Println("\t\toptions:")
	fmt.Println("\t\t  --size N: Specify RAM size:")
	fmt.Println("\t\t\t 1=8KB, 2=32KB, 3=1MB")
	fmt.Println("\t\t If no size is specified, 8KB are erased")
	fmt.Println("\t --help: show this help.\n")
}

func GBSVersion() {
	fmt.Println(color.Green + "⚡ GBShooper version: " + color.Purple + strconv.Itoa(VER_MAYOR) + "." + strconv.Itoa(VER_MINOR) + color.Reset)
}

func main() {
	// no args, print help
	if len(os.Args) == 1 {
		GBSHelp()
		os.Exit(1)
	}

	// select command
	if os.Args[1] == "--version" {
		GBSVersion()
		os.Exit(0)
	}

	if os.Args[1] == "--status" {
		status, err := flashcart.GBSStatus()
		if err != nil {
			fmt.Println("❌ " + color.Red + "Hardware error: ")
			fmt.Println(err.Error() + color.Reset)
			os.Exit(1)
		}
		GBSVersion()
		fmt.Println(color.Green + "🔩 Hardware version: " + color.Purple + string(status.VersionMayor) + "." + string(status.VersionMinor) + color.Reset)
		os.Exit(0)
	}

	if os.Args[1] == "--id" {
		id, err := flashcart.GBSChipID()
		if err != nil {
			fmt.Println("❌ " + color.Red + "Hardware error: ")
			fmt.Println(err.Error() + color.Reset)
			os.Exit(1)
		}
		GBSVersion()
		fmt.Println(color.Green + "🪪  Flash chip ID: " + id.Manufacturer + ", " + id.Chip + color.Reset)
		os.Exit(0)
	}

	if os.Args[1] == "--read-header" {
		header, err := flashcart.GBSReadHeader()
		if err != nil {
			fmt.Println("❌ " + color.Red + "Hardware error: ")
			fmt.Println(err.Error() + color.Reset)
			os.Exit(1)
		}
		GBSVersion()
		fmt.Println(color.Green + "👤 Cart name: " + color.Purple + header.Title + color.Reset)
		fmt.Println(color.Green + "🫆  Cart type: " + color.Purple + header.Cart + color.Reset)
		fmt.Println(color.Green + "📏 ROM size: " + color.Purple + header.ROM + color.Reset)
		fmt.Println(color.Green + "📐 RAM size: " + color.Purple + header.RAM + color.Reset)
		os.Exit(0)
	}

	if os.Args[1] == "--erase-flash" {
		GBSVersion()
		fmt.Println(color.Yellow + "🧼 Erasing FLASH... " + color.Reset)
		bar := progressbar.NewOptions(-1, progressbar.OptionClearOnFinish(), progressbar.OptionSetPredictTime(false), progressbar.OptionSetTheme(progressbar.ThemeUnicode))
		go func() {
			time.Sleep(100 * time.Millisecond)
			bar.Add(1)
		}()
		err := flashcart.GBSEraseFlash()
		if err != nil {
			bar.Clear()
			fmt.Println("❌ " + color.Red + "Error: ")
			fmt.Println(err.Error() + color.Reset)
			os.Exit(1)
		}
		bar.Clear()
		fmt.Println(color.Green + "✅ FLASH erased." + color.Reset)
	}

	if os.Args[1] == "--write-flash" {
		if len(os.Args) < 3 {
			GBSHelp()
			os.Exit(1)
		}

		romFile := os.Args[2]

		// check we can open the file
		rom, err := os.Open(romFile)
		if err != nil {
			fmt.Println("❌ "+color.Red+"Can't open file: ", os.Args[2])
			os.Exit(1)
		}
		rom.Close()

		// sync
		progress := make(chan int64)
		finished := make(chan bool)
		errchan := make(chan error)

		// start
		bar := progressbar.NewOptions(100, progressbar.OptionClearOnFinish(), progressbar.OptionSetPredictTime(false), progressbar.OptionSetWidth(20), progressbar.OptionSetTheme(progressbar.ThemeUnicode))
		GBSVersion()
		fmt.Println(color.Yellow + "📝 Writing FLASH... " + color.Reset)
		go func() {
			err = flashcart.GBSWriteFlash(romFile, finished, progress, errchan)
		}()
	writeflash_outer:
		for {
			select {
			case <-finished:
				break writeflash_outer
			case percent := <-progress:
				bar.Set(int(percent))
			case e := <-errchan:
				err = e
				break writeflash_outer
			}
		}

		if err != nil {
			bar.Clear()
			fmt.Println("❌ "+color.Red+"Error writing flash: ", err.Error())
			os.Exit(1)
		}
		bar.Clear()
		fmt.Println(color.Green + "✅ FLASH written." + color.Reset)
	}

	if os.Args[1] == "--read-flash" {
		if len(os.Args) < 3 {
			GBSHelp()
			os.Exit(1)
		}
		var size int64 = 0
		romFile := ""
		if os.Args[2] == "--size" {
			if len(os.Args) < 5 {
				GBSHelp()
				os.Exit(1)
			}
			s, _ := strconv.Atoi(os.Args[3])
			switch s {
			case 1:
				size = flashcart.S_32K
			case 2:
				size = flashcart.S_64K
			case 3:
				size = flashcart.S_128K
			case 4:
				size = flashcart.S_256K
			case 5:
				size = flashcart.S_512K
			case 6:
				size = flashcart.S_1MB
			case 7:
				size = flashcart.S_2MB
			case 8:
				size = flashcart.S_4MB
			default:
				size = flashcart.S_32K
			}
			romFile = os.Args[4]
		} else {
			romFile = os.Args[2]
			size = flashcart.S_32K
		}

		// sync
		progress := make(chan int64)
		finished := make(chan bool)
		errchan := make(chan error)
		var err error

		// start
		bar := progressbar.NewOptions(100, progressbar.OptionClearOnFinish(), progressbar.OptionSetPredictTime(false), progressbar.OptionSetWidth(20), progressbar.OptionSetTheme(progressbar.ThemeUnicode))
		GBSVersion()
		fmt.Println(color.Yellow + "📖 Reading FLASH... " + color.Reset)
		go func() {
			flashcart.GBSReadFlash(romFile, size, finished, progress, errchan)
		}()
	outerreadflash:
		for {
			select {
			case <-finished:
				break outerreadflash
			case percent := <-progress:
				bar.Set(int(percent))
			case e := <-errchan:
				err = e
				break outerreadflash
			}
		}

		if err != nil {
			bar.Clear()
			fmt.Println("❌ "+color.Red+"Error reading flash: ", err.Error())
			os.Exit(1)
		}
		bar.Clear()
		fmt.Println(color.Green + "✅ FLASH read." + color.Reset)
	}

	if os.Args[1] == "--write-ram" {
		if len(os.Args) < 3 {
			GBSHelp()
			os.Exit(1)
		}

		ramFile := os.Args[2]

		// check we can open the file
		ram, err := os.Open(ramFile)
		if err != nil {
			fmt.Println("❌ "+color.Red+"Can't open file: ", os.Args[2])
			os.Exit(1)
		}
		ram.Close()

		// sync
		progress := make(chan int64)
		finished := make(chan bool)

		// start
		bar := progressbar.NewOptions(100, progressbar.OptionClearOnFinish(), progressbar.OptionSetPredictTime(false), progressbar.OptionSetWidth(20), progressbar.OptionSetTheme(progressbar.ThemeUnicode))
		GBSVersion()
		fmt.Println(color.Yellow + "📝 Writing RAM... " + color.Reset)
		go func() {
			err = flashcart.GBSWriteRAM(ramFile, finished, progress)
		}()
	writeram_outer:
		for {
			select {
			case <-finished:
				break writeram_outer
			case percent := <-progress:
				bar.Set(int(percent))
			}
		}

		if err != nil {
			bar.Clear()
			fmt.Println("❌ "+color.Red+"Error writing RAM: ", err.Error())
			os.Exit(1)
		}
		bar.Clear()
		fmt.Println(color.Green + "✅ RAM written." + color.Reset)
	}

	if os.Args[1] == "--read-ram" {
		if len(os.Args) < 3 {
			GBSHelp()
			os.Exit(1)
		}
		var size int64 = 0
		ramFile := ""
		if os.Args[2] == "--size" {
			if len(os.Args) < 5 {
				GBSHelp()
				os.Exit(1)
			}
			s, _ := strconv.Atoi(os.Args[3])
			switch s {
			case 1:
				size = flashcart.S_8K
			case 2:
				size = flashcart.S_32K
			case 3:
				size = flashcart.S_1MB
			default:
				size = flashcart.S_8K
			}
			ramFile = os.Args[4]
		} else {
			ramFile = os.Args[2]
			size = flashcart.S_32K
		}

		// sync
		progress := make(chan int64)
		finished := make(chan bool)
		errchan := make(chan error)
		var err error

		// start
		bar := progressbar.NewOptions(100, progressbar.OptionClearOnFinish(), progressbar.OptionSetPredictTime(false), progressbar.OptionSetWidth(20), progressbar.OptionSetTheme(progressbar.ThemeUnicode))
		GBSVersion()
		fmt.Println(color.Yellow + "📖 Reading RAM... " + color.Reset)
		go func() {
			flashcart.GBSReadRAM(ramFile, size, finished, progress, errchan)
		}()
	outerreadram:
		for {
			select {
			case <-finished:
				break outerreadram
			case percent := <-progress:
				bar.Set(int(percent))
			case e := <-errchan:
				err = e
				break outerreadram
			}
		}

		if err != nil {
			bar.Clear()
			fmt.Println("❌ "+color.Red+"Error reading RAM: ", err.Error())
			os.Exit(1)
		}
		bar.Clear()
		fmt.Println(color.Green + "✅ RAM read." + color.Reset)
	}

	if os.Args[1] == "--erase-ram" {
		if len(os.Args) < 3 {
			GBSHelp()
			os.Exit(1)
		}
		var size int64 = 0
		if os.Args[2] == "--size" {
			if len(os.Args) < 4 {
				GBSHelp()
				os.Exit(1)
			}
			s, _ := strconv.Atoi(os.Args[3])
			switch s {
			case 1:
				size = flashcart.S_8K
			case 2:
				size = flashcart.S_32K
			case 3:
				size = flashcart.S_1MB
			default:
				size = flashcart.S_8K
			}
		} else {
			size = flashcart.S_32K
		}

		// sync
		progress := make(chan int64)
		finished := make(chan bool)
		errchan := make(chan error)
		var err error

		// start
		bar := progressbar.NewOptions(100, progressbar.OptionClearOnFinish(), progressbar.OptionSetPredictTime(false), progressbar.OptionSetWidth(20), progressbar.OptionSetTheme(progressbar.ThemeUnicode))
		GBSVersion()
		fmt.Println(color.Yellow + "🧼  Erasing RAM... " + color.Reset)
		go func() {
			flashcart.GBSEraseRAM(size, finished, progress, errchan)
		}()
	outereraseram:
		for {
			select {
			case <-finished:
				break outereraseram
			case percent := <-progress:
				bar.Set(int(percent))
			case e := <-errchan:
				err = e
				break outereraseram
			}
		}

		if err != nil {
			bar.Clear()
			fmt.Println("❌ "+color.Red+"Error erasing RAM: ", err.Error())
			os.Exit(1)
		}
		bar.Clear()
		fmt.Println(color.Green + "✅ RAM erased." + color.Reset)
	}
}
