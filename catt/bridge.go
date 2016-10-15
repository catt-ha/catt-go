package catt

import (
	"github.com/Sirupsen/logrus"
	"github.com/catt-ha/catt-go/catt/types"
)

var log = logrus.New()

type Bridge struct {
	done chan struct{}
}

func NewBridge(bus types.Bus, binding types.Binding) Bridge {
	messages := bus.Messages()
	notifications := binding.Notifications()

	done := make(chan struct{})
	busToBinding(messages, binding, done)
	bindingToBus(notifications, bus, done)

	return Bridge{done}
}

func (b Bridge) Run() {
	<-b.done
}

func busToBinding(msgs <-chan types.Message, binding types.Binding, done chan struct{}) {
	go func() {
		defer close(done)
		for msg := range msgs {
			var name string
			var value types.Value
			switch msg.Type {
			case types.CommandMessage:
				name = msg.ItemName
				value = *msg.Value
			default:
				log.WithFields(logrus.Fields{
					"message": msg,
				}).Warn("received non-command message")
				continue
			}

			item := binding.GetValue(name)
			if item == nil {
				log.WithFields(logrus.Fields{
					"item_name": name,
				}).Warn("received message for non-existant item")
				continue
			}

			err := item.SetValue(value)
			if err != nil {
				log.WithFields(logrus.Fields{
					"item":  item,
					"value": value,
				}).Warn("error setting item value")
			}
		}
	}()
}

func bindingToBus(notifications <-chan types.Notification, bus types.Bus, done chan struct{}) {
	go func() {
		defer close(done)
		for notification := range notifications {
			var skipState, newSub, removeSub bool
			var meta *types.Meta

			item := notification.Item

			switch notification.Type {
			case types.ChangedNotification:
			case types.AddedNotification:
				meta = item.GetMeta()
				skipState = true
				newSub = true
			case types.RemovedNotification:
				removeSub = true
				skipState = true
			default:
				log.WithFields(logrus.Fields{
					"notification": notification,
				}).Warn("invalid notification type")
				continue
			}

			if meta != nil {
				if err := bus.Publish(types.Message{
					Type:     types.MetaMessage,
					ItemName: item.GetName(),
					Meta:     meta,
				}); err != nil {
					log.WithFields(logrus.Fields{
						"error": err,
					}).Warn("meta publish error")
				}
			}

			if newSub {
				if err := bus.Subscribe(item.GetName(), types.CommandSub); err != nil {
					log.WithFields(logrus.Fields{
						"error": err,
					}).Warn("subscribe error")
				}
			}

			if removeSub {
				if err := bus.Unsubscribe(item.GetName(), types.CommandSub); err != nil {
					log.WithFields(logrus.Fields{
						"error": err,
					}).Warn("unsubscribe error")
				}
			}

			if skipState {
				continue
			}

			value, err := item.GetValue()
			if err != nil {
				log.WithFields(logrus.Fields{
					"error": err,
					"item":  item,
				}).Warn("error getting item value")
				continue
			}

			if err := bus.Publish(types.Message{
				Type:  types.UpdateMessage,
				Value: &value,
			}); err != nil {
				log.WithFields(logrus.Fields{
					"error": err,
				}).Warn("state publish error")
				continue
			}
		}
	}()
}
