package mqtt

import (
	"fmt"
	"testing"
	"time"

	"github.com/catt-ha/catt-go/catt/types"
)

func TestMqtt(t *testing.T) {
	cfg := Config{
		Broker:   "10.8.0.1:1883",
		ItemBase: "catt/items",
	}

	m, err := NewMqtt(cfg)
	if err != nil {
		panic(err)
	}

	err = m.Subscribe("Light_Living_Table_Switch", types.UpdateSub)
	if err != nil {
		panic(err)
	}

	fmt.Println("subscribed!")

	msgs := m.Messages()

	go func() {
		time.Sleep(time.Second * 1)
		val := types.NewStringValue("ON")
		m.Publish(types.Message{
			Type:     types.UpdateMessage,
			ItemName: "Light_Living_Table_Switch",
			Value:    &val,
		})
		fmt.Println("pubtime: ", time.Now())
	}()

	for msg := range msgs {
		fmt.Println("from chan: ", time.Now())
		fmt.Println(msg.Value.AsString())
	}
}
