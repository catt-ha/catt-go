package catt

import (
	"github.com/catt-ha/catt-go/catt/types"
)

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
				// TODO log it
				continue
			}

			item := binding.GetValue(name)
			if item == nil {
				// TODO log it
				continue
			}

			err := item.SetValue(value)
			if err != nil {
				// TODO log it
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
				//TODO log it
				continue
			}

			if meta != nil {
				if err := bus.Publish(types.Message{
					Type:     types.MetaMessage,
					ItemName: item.GetName(),
					Meta:     meta,
				}); err != nil {
					// TODO log it
				}
			}

			if newSub {
				if err := bus.Subscribe(item.GetName(), types.CommandSub); err != nil {
					// TODO log it
				}
			}

			if removeSub {
				if err := bus.Unsubscribe(item.GetName(), types.CommandSub); err != nil {
					// TODO log it
				}
			}

			if skipState {
				continue
			}

			value, err := item.GetValue()
			if err != nil {
				// TODO log it
				continue
			}

			if err := bus.Publish(types.Message{
				Type:  types.UpdateMessage,
				Value: &value,
			}); err != nil {
				// TODO log it
				continue
			}
		}
	}()
}
