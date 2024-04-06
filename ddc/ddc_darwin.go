package ddc

//#cgo CFLAGS: -x objective-c
//#cgo LDFLAGS: -framework IOKit -framework Foundation -framework CoreGraphics -framework AppKit
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
	isIntel := C.isIntelHardware()
	if isIntel {
		display := C.findDisplayIntel(C.int(index))
		if display == 0 {
			return errors.New("failed to find display")
		}

		err := C.sendDDCIntel(display, SELECT_SOURCE, C.int(value))
		if err != 0 {
			return errors.New("failed to send DDC command")
		}
		return nil
	} else {
		display := C.findDisplayM1(C.int(index))
		if display == nil {
			return errors.New("failed to find display")
		}
		defer C.IOObjectRelease(C.uint(uintptr(unsafe.Pointer(display))))

		err := C.sendDDCM1(display, SELECT_SOURCE, C.int(value))
		if err != 0 {
			return errors.New("failed to send DDC command")
		}
		return nil
	}
}

func (d *DDCControl) SetInputSource(index int, source InputSource) error {
	err := d.SendCommand(index, SELECT_SOURCE, int(source))
	if err != nil {
		log.Error().Msg(err.Error())
	}
	return err
}
