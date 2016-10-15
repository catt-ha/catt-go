package types

import (
	"bytes"

	"github.com/BurntSushi/toml"
)

type Meta struct {
	Backend   string            `toml:"backend,omitempty"`
	ValueType string            `toml:"value_type,omitempty"`
	Ext       map[string]string `toml:"ext"`
}

func (m Meta) AsString() (string, error) {
	buf := &bytes.Buffer{}
	enc := toml.NewEncoder(buf)
	err := enc.Encode(m)
	return buf.String(), err
}

func (m *Meta) FromString(s string) error {
	_, err := toml.DecodeReader(bytes.NewReader([]byte(s)), m)
	return err
}

type Item interface {
	GetName() string
	GetMeta() *Meta
	GetValue() (Value, error)
	SetValue(Value) error
}
