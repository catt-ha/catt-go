package types

type MessageType uint8
type SubType uint8

const (
	UpdateMessage MessageType = iota
	CommandMessage
	MetaMessage
)

const (
	UpdateSub SubType = iota
	CommandSub
	MetaSub
	AllSub
)

type Message struct {
	Type     MessageType
	ItemName string
	Value    *Value
	Meta     *Meta
}

type Bus interface {
	Publish(Message) error
	Subscribe(string, SubType) error
	Unsubscribe(string, SubType) error
	Messages() <-chan Message
}
