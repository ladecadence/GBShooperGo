# ☄️ GBShooperGo

Command line client software for the GBShooper hardware.

## Building

You'll need Go (> 1.24, golang.org) and libftdi installed. In debian install libftdi-dev and libftdi1-dev packages.
Then you can get and compile the program with:

```
$ git clone https://github.com/ladecadence/GBShooperGo.git
$ cd GBShooperGo
$ go mod tidy
$ go build cmd/gbshooper.go
```

## Running

You can see the available commands running the program without options:

```
$ ./gbshooper
☄️  GBShooperGo version: 0.2
David Pello 2025

Usage:
gbshoopergo <action> <options> [file]

Actions:
	 --version: prints the software version.
	 --status: checks the hardware.
	 --id: gets the ID of the flash chip.
	 --read-header: gets header information, mapper and RAM/ROM sizes.
	 --erase-flash: clears the contents of the flash chip.
	 --read-flash: reads the contents of the flash chip and writes it on [file].
		options:
		  --size N: Specify ROM size:
			 1=32KB, 2=64KB, 3=128KB, 4=256KB, 5=512KB, 6=1MB, 7=2MB, 8=4MB
		 If no size is specified, 32KB are read
	 --write-flash: writes the flash with contents from [file].
	 --read-ram: reads the contents of the save RAM and writes it on [file].
		options:
		  --size N: Specify RAM size:
			 1=8KB, 2=32KB, 3=1MB
		 If no size is specified, 8KB are read
	 --write-ram: writes the save RAM with contents from [file].
	 --erase-ram: clears the contents of the save RAM with 0's.
		options:
		  --size N: Specify RAM size:
			 1=8KB, 2=32KB, 3=1MB
		 If no size is specified, 8KB are erased
	 --help: show this help.
```