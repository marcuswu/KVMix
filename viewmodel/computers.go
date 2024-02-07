package viewmodel

import (
	"fmt"

	"github.com/marcuswu/KVMix/config"
	"github.com/marcuswu/KVMix/ddc"
	"github.com/marcuswu/gosmartknob/pb"
)

type ComputersViewModel struct {
	position   int32
	pressNonce uint32
	computers  []config.Computer
	ddc        ddc.DDC
}

func NewComputersViewModel(pressNonce uint32, computers []config.Computer, ddc ddc.DDC) *ComputersViewModel {
	return &ComputersViewModel{
		position:   0,
		pressNonce: pressNonce,
		computers:  computers,
		ddc:        ddc,
	}
}

func (cvm *ComputersViewModel) HandleMessage(state *pb.SmartKnobState) NavAction {
	// Check for select of App Channel or Back
	// Return new ViewModel accordingly
	ret := NavAction{
		Navigation:  NavNone,
		ViewModel:   nil,
		RegenConfig: false,
	}
	if state.CurrentPosition != cvm.position {
		cvm.position = state.CurrentPosition
		ret.RegenConfig = true
	}
	if state.PressNonce != cvm.pressNonce {
		ret.Navigation = NavBack
		ret.pressNonce = state.PressNonce
		ret.RegenConfig = true
		if cvm.position > 0 {
			// Switch to the designated computer as we head back
			computer := cvm.computers[cvm.position-1]
			cvm.ddc.SetInputSource(computer.MonitorIndex, computer.Display)
		}
	}
	return ret
}

func (cvm *ComputersViewModel) GenerateConfig() *pb.SmartKnobConfig {
	title := "Back"
	if cvm.position > 0 {
		title = fmt.Sprintf("Switch to %s", cvm.computers[cvm.position-1].Name)
	}
	return &pb.SmartKnobConfig{
		Position:           cvm.position,
		MinPosition:        0,
		MaxPosition:        int32(len(cvm.computers)),
		DetentStrengthUnit: 0.4,
		SnapPoint:          0.7,
		Text:               title,
	}
}
