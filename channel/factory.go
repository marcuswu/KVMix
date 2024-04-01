package channel

type ChannelFactory interface {
	ChannelsMatching(map[string]interface{}) ([]Channel, error)
}
