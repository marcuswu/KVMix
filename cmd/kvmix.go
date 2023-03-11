package main

import ddc "github.com/marcuswu/KVMix/DDC"

func main() {
	myDDC := ddc.NewDDC()
	myDDC.SetInputSource(1, ddc.HDMI1) // From Windows to OSX
	// myDDC.SetInputSource(0, ddc.DISPLAY_PORT1) // From OSX to Windows
}
