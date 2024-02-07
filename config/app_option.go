package config

import "github.com/marcuswu/KVMix/channel"

type AppOption struct {
	channel channel.Channel
}

func NewAppOption(channel channel.Channel) *AppOption {
	return &AppOption{channel: channel}
}
