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

		// start
		bar := progressbar.NewOptions(100, progressbar.OptionClearOnFinish(), progressbar.OptionSetPredictTime(false), progressbar.OptionSetWidth(20), progressbar.OptionSetTheme(progressbar.ThemeUnicode))
		GBSVersion()
		fmt.Println(color.Yellow + "📝 Writing FLASH... " + color.Reset)
		go func() {
			err = flashcart.GBSWriteFlash(romFile, finished, progress)
		}()
	outer:
		for {
			select {
			case <-finished:
				break outer
			case percent := <-progress:
				bar.Set(int(percent))
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
}
