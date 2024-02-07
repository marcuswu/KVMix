package config

import "github.com/marcuswu/gosmartknob/pb"

type Option interface {
	GenerateConfig() *pb.SmartKnobConfig
}
