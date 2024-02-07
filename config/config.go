package config

import ddc "github.com/marcuswu/KVMix/ddc"

/*
Config
* List apps to handle volume for
  - Display name
  - App identifier

* List computers to toggle between
  - Name
  - Display
  - Pin
*/

type VolumeApp struct {
	DisplayName string `yaml:"display_name"`
	Identifier  string `yaml:"app_identifier"`
}

type Computer struct {
	Name         string
	MonitorIndex int `yaml:"monitor_index"`
	Display      ddc.InputSource
	Pin          string
}

type Config struct {
	VolumeApps []VolumeApp `yaml:"apps"`
	Computers  []Computer
}
