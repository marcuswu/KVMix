package channel

/*
We need to get a list of Channels.
Each channel will be the handle to control volume for a process

So to produce a list of Channels, we will iterate processes.

A Channel is simply something you can get and set volume on
*/

type Channel interface {
	Name() string
	GetVolume() float64
	SetVolume(float64) error
}
