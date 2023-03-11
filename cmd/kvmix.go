package main

import ddc "github.com/marcuswu/KVMix/DDC"

func main() {
	myDDC := ddc.NewDDC()
	myDDC.SetInputSource(1, ddc.HDMI1)
}
