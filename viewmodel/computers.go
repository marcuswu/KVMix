package viewmodel

import (
	"fmt"
	"math"

	"github.com/marcuswu/KVMix/config"
	"github.com/marcuswu/KVMix/ddc"
	"github.com/marcuswu/gosmartknob/pb"
	"github.com/rs/zerolog/log"
)

type ComputersViewModel struct {
	position       int32
	positionNonce  uint32
	pressNonce     uint32
	pinNonce       uint32
	computers      []config.Computer
	ddc            ddc.DDC
	setNonce       bool
	updatePinNonce bool
}

func NewComputersViewModel(computers []config.Computer, ddc ddc.DDC) *ComputersViewModel {
	return &ComputersViewModel{
		position:       0,
		pressNonce:     0,
		pinNonce:       0,
		positionNonce:  0,
		computers:      computers,
		ddc:            ddc,
		setNonce:       false,
		updatePinNonce: false,
	}
}

func (cvm *ComputersViewModel) getNonceSet() bool {
	return cvm.setNonce
}

func (cvm *ComputersViewModel) setNonceSet(isSet bool) {
	cvm.setNonce = isSet
}

func (cvm *ComputersViewModel) getPositionNonce() uint32 {
	return cvm.positionNonce
}

func (cvm *ComputersViewModel) setPositionNonce(nonce uint32) {
	cvm.positionNonce = nonce
}

func (cvm *ComputersViewModel) getPosition() int32 {
	return cvm.position
}

func (cvm *ComputersViewModel) setPosition(pos int32) {
	cvm.position = pos
	if len(cvm.computers) < int(pos) {
		cvm.position = 0
	}
}

func (cvm *ComputersViewModel) getPressNonce() uint32 {
	return cvm.pressNonce
}

func (cvm *ComputersViewModel) setPressNonce(nonce uint32) {
	cvm.pressNonce = nonce
}

func (cvm *ComputersViewModel) HandleMessage(state *pb.SmartKnobState) NavAction {
	// Check for select of App Channel or Back
	// Return new ViewModel accordingly
	ret := NavAction{
		Navigation:  NavNone,
		ViewModel:   nil,
		RegenConfig: false,
	}
	if !cvm.getNonceSet() {
		log.Debug().Uint32("nonce", cvm.pinNonce).Msg("Initializing pin nonce")
		cvm.pinNonce = state.GetConfig().PinNonce
	}
	handleNonces(cvm, state)
	// The knob position has updated
	if state.CurrentPosition != cvm.getPosition() {
		cvm.setPosition(state.CurrentPosition)
		ret.RegenConfig = true
	}
	if state.PressNonce != cvm.pressNonce {
		ret.Navigation = NavBack
		ret.pressNonce = state.PressNonce
		ret.RegenConfig = true
		if cvm.position > 0 {
			// Switch to the designated computer as we head back
			computer := cvm.computers[cvm.position-1]
			log.Debug().Int("monitor", computer.MonitorIndex).Int("display", int(computer.Display)).Msg("Setting input source")
			cvm.ddc.SetInputSource(computer.MonitorIndex, computer.Display)
			log.Debug().Msg("Switching monitor and setting need nonce change")
			cvm.updatePinNonce = true
		}
	}
	return ret
}

func (cvm *ComputersViewModel) Restore(state *pb.SmartKnobState) {
	restorePosition(cvm, state)
}

func (cvm *ComputersViewModel) GenerateConfigBeforeBack() bool {
	return true
}

func (cvm *ComputersViewModel) GenerateConfig() *pb.SmartKnobConfig {
	title := "Back"
	if cvm.position > 0 {
		title = fmt.Sprintf("Switch to %s", cvm.computers[cvm.position-1].Name)
	}
	if cvm.updatePinNonce {
		cvm.pinNonce += 1
		log.Debug().Uint32("nonce", cvm.pinNonce).Msg("Updating pin nonce")
		cvm.updatePinNonce = false
	}
	return &pb.SmartKnobConfig{
		MinPosition:          0,
		MaxPosition:          int32(len(cvm.computers)),
		PositionWidthRadians: (30 / 180.0) * math.Pi,
		EndstopStrengthUnit:  1,
		DetentStrengthUnit:   0.4,
		SnapPoint:            0.7,
		Text:                 title,
		PinNonce:             cvm.pinNonce,
	}
}
