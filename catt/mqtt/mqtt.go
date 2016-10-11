package mqtt

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/catt-ha/catt-go/catt/types"
	emqtt "github.com/eclipse/paho.mqtt.golang"
)

type mqtt struct {
	cfg    Config
	client emqtt.Client
	cb     emqtt.MessageHandler
}

func newMqtt(cfg Config) (*mqtt, error) {
	opts := emqtt.NewClientOptions().SetKeepAlive(5 * time.Second).SetAutoReconnect(true).SetMaxReconnectInterval(3 * time.Second)

	if cfg.ClientId != "" {
		opts.SetClientID(cfg.ClientId)
	}

	proto := "tcp"
	if cfg.Tls {
		proto = "ssl"
	}

	broker := "127.0.0.1:1883"
	if cfg.Broker != "" {
		broker = cfg.Broker
	}

	opts.AddBroker(fmt.Sprintf("%s://%s", proto, broker))

	client := emqtt.NewClient(opts)
	tok := client.Connect()
	tok.Wait()
	err := tok.Error()

	if err != nil {
		return nil, err
	}

	return &mqtt{
		cfg:    cfg,
		client: client,
	}, nil
}

func (m *mqtt) Subscribe(subPath string) error {
	if m.cb == nil {
		return errors.New("message handler not set")
	}

	var fullPath string

	if m.cfg.ItemBase == "" {
		fullPath = path.Join("catt/items", subPath)
	} else {
		fullPath = path.Join(m.cfg.ItemBase, subPath)
	}

	tok := m.client.Subscribe(fullPath, 0, m.cb)

	tok.Wait()

	return tok.Error()
}

func (m *mqtt) Publish(pubPath string, state []byte) error {
	var fullPath string
	if m.cfg.ItemBase == "" {
		fullPath = path.Join("catt/items", pubPath)
	} else {
		fullPath = path.Join(m.cfg.ItemBase, pubPath)
	}

	tok := m.client.Publish(fullPath, 0, false, state)
	tok.Wait()

	return tok.Error()
}

func (m *mqtt) Unsubscribe(subPath string) error {
	if m.cb == nil {
		return errors.New("message handler not set")
	}

	var fullPath string

	if m.cfg.ItemBase == "" {
		fullPath = path.Join("catt/items", subPath)
	} else {
		fullPath = path.Join(m.cfg.ItemBase, subPath)
	}

	tok := m.client.Unsubscribe(fullPath)

	tok.Wait()

	return tok.Error()
}

type Mqtt struct {
	client  *mqtt
	msgChan chan types.Message
}

func NewMqtt(cfg Config) (*Mqtt, error) {
	m, err := newMqtt(cfg)
	if err != nil {
		return nil, err
	}

	msgChan := make(chan types.Message, 16)
	cb := func(cl emqtt.Client, msg emqtt.Message) {
		splitPath := strings.Split(msg.Topic(), "/")
		l := len(splitPath)
		if l < 2 {
			// TODO log this
			return
		}

		itemName := splitPath[l-2]
		outMsg := types.Message{
			ItemName: itemName,
		}

		val := new(types.Value)
		meta := new(types.Meta)
		msgTypeStr := splitPath[l-1]
		switch msgTypeStr {
		case "state":
			val.FromRaw(msg.Payload())
			outMsg.Type = types.UpdateMessage
			outMsg.Value = val
		case "command":
			val.FromRaw(msg.Payload())
			outMsg.Type = types.CommandMessage
			outMsg.Value = val
		case "meta":
			err := meta.FromString(string(msg.Payload()))
			if err != nil {
				// TODO log this
				return
			}
			outMsg.Type = types.MetaMessage
			outMsg.Meta = meta
		default:
			// TODO log this
			return
		}

		go func() {
			msgChan <- outMsg
		}()
	}
	m.cb = cb

	return &Mqtt{client: m, msgChan: msgChan}, nil
}

func (m *Mqtt) Subscribe(itemName string, subType types.SubType) error {
	var last string
	switch subType {
	case types.UpdateSub:
		last = "state"
	case types.CommandSub:
		last = "command"
	case types.MetaSub:
		last = "meta"
	case types.AllSub:
		last = "#"
	default:
		return fmt.Errorf("invalid sub type: %d", subType)
	}

	subPath := path.Join(itemName, last)

	return m.client.Subscribe(subPath)
}

func (m *Mqtt) Unsubscribe(itemName string, subType types.SubType) error {
	var last string
	switch subType {
	case types.UpdateSub:
		last = "state"
	case types.CommandSub:
		last = "command"
	case types.MetaSub:
		last = "meta"
	case types.AllSub:
		last = "#"
	default:
		return fmt.Errorf("invalid sub type: %d", subType)
	}

	subPath := path.Join(itemName, last)

	return m.client.Unsubscribe(subPath)
}

func (m *Mqtt) Publish(message types.Message) error {
	var last string
	var val string
	var err error
	switch message.Type {
	case types.UpdateMessage:
		last = "state"
		val, err = message.Value.AsString()
	case types.CommandMessage:
		last = "command"
		val, err = message.Value.AsString()
	case types.MetaMessage:
		last = "meta"
		val, err = message.Meta.AsString()
	default:
		return fmt.Errorf("invalid message type: %d", message.Type)
	}

	if err != nil {
		return err
	}

	pubPath := path.Join(message.ItemName, last)

	return m.client.Publish(pubPath, []byte(val))
}

func (m *Mqtt) Messages() <-chan types.Message {
	return m.msgChan
}
