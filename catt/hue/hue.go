package hue

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/catt-ha/catt-go/catt/types"
	"github.com/gbbr/hue"
)

type Hue struct {
	mu            sync.Mutex
	internal      *hue.Bridge
	items         map[string]*HueItem
	notifications chan types.Notification
}

func NewHue() (*Hue, error) {
	b, err := hue.Discover()
	if err != nil {
		return nil, err
	}
	if !b.IsPaired() {
		// link button must be pressed before calling
		fmt.Println("Please press the link button on your Hue!")
		if err := b.Pair(); err != nil {
			return nil, err
		}

		fmt.Println("Pairing successful!")
	}
	binding := &Hue{
		internal:      b,
		items:         make(map[string]*HueItem),
		notifications: make(chan types.Notification),
	}

	startWatcher(binding)

	return binding, nil
}

func startWatcher(binding *Hue) {
	go func() {
		for range time.NewTicker(5 * time.Second).C {
			binding.mu.Lock()
			lights, err := binding.internal.Lights().List()
			if err != nil {
				log.Print("ERROR: ", err)
			}
			lightsMap := buildMap(lights)
			for k, v := range lightsMap {
				if i, ok := binding.items[k]; !ok {
					newItem := &HueItem{
						itemType: v.itemType,
						light:    v.light,
						updates:  binding.notifications,
					}
					binding.items[k] = newItem
					binding.notifications <- types.Notification{
						Type: types.AddedNotification,
						Item: newItem,
					}
				} else {
					if lightChanged(i, v) {
						binding.notifications <- types.Notification{
							Type: types.ChangedNotification,
							Item: i,
						}
					}
				}
			}
			toDelete := []string{}
			for k, v := range binding.items {
				if _, ok := lightsMap[k]; !ok {
					toDelete = append(toDelete, k)
					binding.notifications <- types.Notification{
						Type: types.RemovedNotification,
						Item: v,
					}
				}
			}
			for _, v := range toDelete {
				delete(binding.items, v)
			}
		}
	}()
}

func buildMap(lights []*hue.Light) map[string]*HueItem {
	m := make(map[string]*HueItem)
	for _, v := range lights {
		colorName := v.Name + "_Color"
		onoffName := v.Name + "_Switch"
		m[onoffName] = &HueItem{
			light:    v,
			itemType: onoffType,
		}
		m[colorName] = &HueItem{
			light:    v,
			itemType: colorType,
		}
	}
	return m
}

func lightChanged(from, to *HueItem) bool {
	switch from.itemType {
	case onoffType:
		return from.light.State.On != to.light.State.On
	case colorType:
		from := from.light.State
		to := to.light.State
		return from.Hue != to.Hue || from.Saturation != to.Saturation || from.Brightness != to.Brightness
	default:
		return false
	}
}

func (h *Hue) GetValue(name string) types.Item {
	h.mu.Lock()
	it, ok := h.items[name]
	h.mu.Unlock()
	if ok {
		return it
	}
	return nil
}

func (h *Hue) Notifications() <-chan types.Notification {
	return h.notifications
}

var _ types.Binding = &Hue{}
