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

func handleNonces(vm ViewModel, state *pb.SmartKnobState) bool {
	ret := false
	// Initialize our nonces and positions
	if !vm.getNonceSet() {
		vm.setPressNonce(state.PressNonce)
		// Update the position nonce so our position gets set to the knob
		// Force the overwrite because we're on a new viewmodel / screen
		vm.setPositionNonce(state.GetConfig().PositionNonce + 1)
		vm.setNonceSet(true)
		return false
	}
	// The knob has a more recent position
	if state.GetConfig().PositionNonce == vm.getPositionNonce() && state.CurrentPosition != vm.getPosition() {
		vm.setPosition(state.CurrentPosition)
		vm.setPositionNonce(state.GetConfig().PositionNonce)
		ret = true
	}
	// The knob needs an updated position
	if state.GetConfig().PositionNonce < vm.getPositionNonce() {
		vm.setPositionNonce(state.GetConfig().PositionNonce + 1)
		ret = true
	}
	return ret
}
