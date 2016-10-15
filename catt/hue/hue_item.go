package hue

import (
	"errors"
	"fmt"
	"sync"

	"github.com/catt-ha/catt-go/catt/types"
	"github.com/gbbr/hue"
)

type itemType uint8

const (
	onoffType itemType = iota
	colorType
)

func (it itemType) String() string {
	switch it {
	case onoffType:
		return "bool"
	case colorType:
		return "color"
	}
	return "???"
}

type HueItem struct {
	itemType itemType
	light    *hue.Light
	mu       sync.Mutex
	updates  chan types.Notification
}

var _ types.Item = &HueItem{}

func (hi *HueItem) GetName() string {
	hi.mu.Lock()
	defer hi.mu.Unlock()
	typeExt := "Switch"
	switch hi.itemType {
	case colorType:
		typeExt = "Color"
	default:
	}
	return fmt.Sprintf("%s_%s", hi.light.Name, typeExt)
}

func (hi *HueItem) GetMeta() *types.Meta {
	return &types.Meta{
		Backend:   "hue",
		ValueType: hi.itemType.String(),
		Ext:       nil,
	}
}

func (hi *HueItem) GetValue() (types.Value, error) {
	hi.mu.Lock()
	defer hi.mu.Unlock()
	switch hi.itemType {
	case onoffType:
		return types.NewBoolValue(hi.light.State.On), nil
	case colorType:
		return toCattColor(&hi.light.State), nil
	default:
		return types.Value{}, errors.New("invalid item type")
	}
}

func (hi *HueItem) SetValue(val types.Value) error {
	hi.mu.Lock()
	defer hi.mu.Unlock()
	switch hi.itemType {
	case onoffType:
		newState, err := val.AsBool()
		if err != nil {
			return err
		}
		if newState {
			hi.light.On()
		} else {
			hi.light.Off()
		}
	case colorType:
		newColor, err := val.AsColor()
		if err != nil {
			return err
		}
		h, s, v := fromCattColor(newColor)
		state := &hue.State{
			Hue:        h,
			Saturation: s,
			Brightness: v,
		}
		if err = hi.light.Set(state); err != nil {
			return err
		}
	}
	hi.updates <- types.Notification{
		Type: types.ChangedNotification,
		Item: hi,
	}
	return nil
}

func toCattColor(state *hue.LightState) types.Value {
	h := float64(float64(state.Hue) * (360.0 / 65535.0))
	s := float64(float64(state.Saturation) * (1.0 / 255.0))
	v := float64(float64(state.Brightness) * (1.0 / 255.0))
	return types.NewColorValue(types.Color{h, s, v})
}

func fromCattColor(val types.Color) (h uint16, s uint8, v uint8) {
	h = uint16(val.H * (65535.0 / 360.0))
	s = uint8(val.S * 255.0)
	v = uint8(val.V * 255.0)
	return
}
