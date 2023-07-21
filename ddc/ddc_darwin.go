package ddc

//#cgo CFLAGS: -x objective-c
//#cgo LDFLAGS: -framework IOKit -framework Foundation
//#import <Foundation/Foundation.h>
//#import <IOKit/IOKitLib.h>
//#include "ddc_darwin.h"
import "C"
import (
	"errors"
	"unsafe"

	"github.com/rs/zerolog/log"
)

type DDCControl struct {
}

func NewDDC() DDC {
	return &DDCControl{}
}

func (d *DDCControl) SendCommand(index int, command int, value int) error {
	display := C.findDisplay(C.int(index))
	defer C.IOObjectRelease(C.uint(uintptr(unsafe.Pointer(display))))

	err := C.sendDDC(display, SELECT_SOURCE, C.int(value))
	if err != 0 {
		return errors.New("Failed to send DDC command")
	}
	return nil
}

func (d *DDCControl) SetInputSource(index int, source InputSource) error {
	err := d.SendCommand(index, SELECT_SOURCE, int(source))
	if err != nil {
		log.Error().Msg(err.Error())
	}
	return err
}
