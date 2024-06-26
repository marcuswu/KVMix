package channel

import (
	"fmt"
	"strconv"

	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
)

type DarwinMainChannel struct {
	name string
}

func (ch DarwinMainChannel) Name() string {
	return ch.name
}

func (ch DarwinMainChannel) GetVolume() (float64, error) {
	var error *core.NSDictionary = nil
	code := core.NSString_FromString("tell application \"Background Music\" to get output volume")
	script := objc.Get("NSAppleScript").Alloc().Send("initWithSource:", code)
	result := script.Send("executeAndReturnError:", &error)
	strVolume := result.Send("stringValue", nil).String()
	volume, err := strconv.ParseFloat(strVolume, 64)
	if err != nil {
		return 0, err
	}

	return volume, nil
}

func (ch DarwinMainChannel) SetVolume(vol float64) error {
	error := objc.ObjectPtr(0)
	code := core.NSString_FromString(fmt.Sprintf("tell application \"Background Music\" to set output volume to %f", vol))
	script := objc.Get("NSAppleScript").Alloc().Send("initWithSource:", code)
	_ = script.Send("executeAndReturnError:", &error)

	return nil
}

type DarwinChannel struct {
	name string
}

func (ch DarwinChannel) Name() string {
	return ch.name
}

func (ch DarwinChannel) GetVolume() (float64, error) {
	var error *core.NSDictionary = nil
	code := core.NSString_FromString("tell application \"Background Music\" to get vol of (a reference to (the first audio application whose name is equal to \"" + ch.name + "\"))")
	script := objc.Get("NSAppleScript").Alloc().Send("initWithSource:", code)
	result := script.Send("executeAndReturnError:", &error)
	strVolume := result.Send("stringValue", nil).String()
	volume, err := strconv.ParseFloat(strVolume, 64)
	if err != nil {
		return 0, err
	}

	return volume / maxVolume, nil
}

func (ch DarwinChannel) SetVolume(vol float64) error {
	volume := uint32(vol * maxVolume)
	var error *core.NSDictionary = nil
	code := core.NSString_FromString(fmt.Sprintf("tell application \"Background Music\" to set vol of (a reference to (the first audio application whose name is equal to \"%s\")) to %d", ch.name, volume))
	script := objc.Get("NSAppleScript").Alloc().Send("initWithSource:", code)
	_ = script.Send("executeAndReturnError:", &error)

	return nil
}
