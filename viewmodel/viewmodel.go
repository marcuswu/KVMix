package viewmodel

import "github.com/marcuswu/gosmartknob/pb"

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
	GenerateConfig() *pb.SmartKnobConfig
}
