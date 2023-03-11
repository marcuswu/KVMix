package ddc

//#include <IOKitLib.h>
//#include "ddc_darwin.h"
import "C"

type DDCControl struct {
}

func NewDDC() DDC {
	return &DDCControl{}
}

func (d *DDCControl) SendCommand(index int, command int, value int) error {
	display := C.findDisplay(index)
	defer C.IOObjectRelease(display)

	err := C.sendDDC(display, C.INPUT_SWITCH, source)
	if err != 0 {
		return errors.error("Failed to send DDC command")
	}
}

func (d *DDCControl) SetInputSource(index int, source InputSource) error {
	err := d.SendCommand(index, SELECT_SOURCE, int(source))
	if err != nil {
		log.Error().Msg(err.Error())
	}
	return err
}
