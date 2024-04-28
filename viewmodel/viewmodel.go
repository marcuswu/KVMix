package viewmodel

import (
	"github.com/marcuswu/gosmartknob/pb"
)

type NavType int

const (
	NavBack NavType = iota
	NavTo
	NavNone
)

type NavAction struct {
	Navigation  NavType
	ViewModel   ViewModel
	RegenConfig bool
	pressNonce  uint32
}

type ViewModel interface {
	HandleMessage(*pb.SmartKnobState) NavAction
	Restore(*pb.SmartKnobState)
	GenerateConfig() *pb.SmartKnobConfig
	GenerateConfigBeforeBack() bool
	getNonceSet() bool
	setNonceSet(bool)
	getPositionNonce() uint32
	setPositionNonce(uint32)
	getPosition() int32
	setPosition(int32)
	getPressNonce() uint32
	setPressNonce(uint32)
}

func restorePosition(vm ViewModel, state *pb.SmartKnobState) {
	// When returning to a previous ViewModel, the position from the ViewModel is correct
	vm.setPositionNonce(state.GetConfig().PositionNonce + 1)
	// Update the press nonce so subsequent presses register
	vm.setPressNonce(state.PressNonce)
}

func handleNonces(vm ViewModel, state *pb.SmartKnobState) {
	// Initialize our nonces and positions
	if vm.getNonceSet() {
		return
	}
	vm.setPressNonce(state.PressNonce)
	vm.setNonceSet(true)
}
