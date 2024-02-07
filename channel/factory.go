package channel

type ChannelFactory interface {
	Channels() ([]Channel, error)
}
