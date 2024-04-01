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
	position      int32
	positionNonce uint32
	pressNonce    uint32
	computers     []config.Computer
	ddc           ddc.DDC
	setNonce      bool
}

func NewComputersViewModel(computers []config.Computer, ddc ddc.DDC) *ComputersViewModel {
	return &ComputersViewModel{
		position:      0,
		pressNonce:    0,
		positionNonce: 0,
		computers:     computers,
		ddc:           ddc,
		setNonce:      false,
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
	ret.RegenConfig = handleNonces(cvm, state)
	if state.PressNonce != cvm.pressNonce {
		ret.Navigation = NavBack
		ret.pressNonce = state.PressNonce
		ret.RegenConfig = true
		if cvm.position > 0 {
			// Switch to the designated computer as we head back
			computer := cvm.computers[cvm.position-1]
			log.Debug().Int("monitor", computer.MonitorIndex).Int("display", int(computer.Display)).Msg("Setting input source")
			cvm.ddc.SetInputSource(computer.MonitorIndex, computer.Display)
		}
	}
	return ret
}

func (cvm *ComputersViewModel) Restore(state *pb.SmartKnobState) {
	restorePosition(cvm, state)
}

func (cvm *ComputersViewModel) GenerateConfig() *pb.SmartKnobConfig {
	title := "Back"
	if cvm.position > 0 {
		title = fmt.Sprintf("Switch to %s", cvm.computers[cvm.position-1].Name)
	}
	return &pb.SmartKnobConfig{
		Position:             cvm.position,
		PositionNonce:        cvm.positionNonce,
		MinPosition:          0,
		MaxPosition:          int32(len(cvm.computers)),
		PositionWidthRadians: (30 / 180.0) * math.Pi,
		EndstopStrengthUnit:  1,
		DetentStrengthUnit:   0.4,
		SnapPoint:            0.7,
		Text:                 title,
	}
}
