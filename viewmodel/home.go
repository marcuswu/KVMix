package viewmodel

import (
	"github.com/marcuswu/KVMix/channel"
	"github.com/marcuswu/KVMix/config"
	"github.com/marcuswu/KVMix/ddc"
	"github.com/marcuswu/gosmartknob/pb"
)

type HomePositions int

const (
	ChangeVolumePosition HomePositions = iota
	ChangeComputerPosition
)

type HomeViewModel struct {
	position      int32
	positionNames []string
	pressNonce    uint32
	config        config.Config
	ddc           ddc.DDC
}

func NewHomeViewModel(pressNonce uint32, config config.Config, ddc ddc.DDC) *HomeViewModel {
	return &HomeViewModel{
		position:      0,
		positionNames: []string{"Change Volume", "Change Computer"},
		pressNonce:    pressNonce,
		config:        config,
		ddc:           ddc,
	}
}

func (hvm *HomeViewModel) HandleMessage(state *pb.SmartKnobState) NavAction {
	// Check for select of ChangeVolume or ChangeComputer
	// Return new ViewModel accordingly
	ret := NavAction{
		Navigation:  NavNone,
		ViewModel:   nil,
		RegenConfig: false,
	}
	if state.CurrentPosition != hvm.position {
		hvm.position = state.CurrentPosition
		ret.RegenConfig = true
	}
	if state.PressNonce != hvm.pressNonce {
		switch hvm.position {
		case int32(ChangeVolumePosition):
			ret.ViewModel = NewVolumeAppViewModel(hvm.pressNonce, channel.New())
		case int32(ChangeComputerPosition):
			ret.ViewModel = NewComputersViewModel(hvm.pressNonce, hvm.config.Computers, hvm.ddc)
		default:
			// Shouldn't land here, but if we do, do nothing
			return ret
		}
		ret.Navigation = NavTo
		ret.RegenConfig = true
	}
	return ret
}

func (hvm *HomeViewModel) GenerateConfig() *pb.SmartKnobConfig {
	return &pb.SmartKnobConfig{
		Position:           hvm.position,
		MinPosition:        0,
		MaxPosition:        int32(len(hvm.positionNames)) - 1,
		DetentStrengthUnit: 0.4,
		SnapPoint:          0.7,
		Text:               hvm.positionNames[hvm.position],
	}
}
