package types

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/BurntSushi/toml"
)

type ValueType uint8

const (
	RawValue ValueType = iota
	StringValue
	NumberValue
	BoolValue
	ColorValue
)

type Color struct {
	H float64
	S float64
	V float64
}

type Value struct {
	Type   ValueType
	number float64
	string string
	bool   bool
	color  Color
	raw    []byte
}

func NewStringValue(s string) Value {
	return Value{
		Type:   StringValue,
		string: s,
	}
}
func NewNumberValue(n float64) Value {
	return Value{
		Type:   NumberValue,
		number: n,
	}
}
func NewBoolValue(b bool) Value {
	return Value{
		Type: BoolValue,
		bool: b,
	}
}
func NewRawValue(r []byte) Value {
	return Value{
		Type: RawValue,
		raw:  r,
	}
}

func NewColorValue(c Color) Value {
	return Value{
		Type:  ColorValue,
		color: c,
	}
}

func (v Value) AsString() (string, error) {
	switch v.Type {
	case RawValue:
		return base64.StdEncoding.EncodeToString(v.raw), nil
	case StringValue:
		return v.string, nil
	case NumberValue:
		return fmt.Sprintf("%v", v.number), nil
	case BoolValue:
		if v.bool {
			return "ON", nil
		}
		return "OFF", nil
	case ColorValue:
		buf := bytes.Buffer{}
		enc := toml.NewEncoder(&buf)
		enc.Encode(v.color)
		return buf.String(), nil
	default:
		return "", fmt.Errorf("invalid value type: %d", v.Type)
	}
}

func (v Value) AsStringValue() (Value, error) {
	s, err := v.AsString()
	return NewStringValue(s), err
}

func (v Value) AsBool() (bool, error) {
	switch v.Type {
	case RawValue:
		return len(v.raw) > 0 && v.raw[0] != 0, nil
	case StringValue:
		switch strings.ToLower(v.string) {
		case "on", "open", "true":
			return true, nil
		case "off", "closed", "false":
			return false, nil
		default:
			return false, fmt.Errorf("invalid bool string: %s", v.string)
		}
	case NumberValue:
		return int64(v.number) != 0, nil
	case BoolValue:
		return v.bool, nil
	case ColorValue:
		return int64(v.color.V) != 0, nil
	default:
		return false, fmt.Errorf("invalid value type: %d", v.Type)
	}
}

func (v Value) AsBoolValue() (Value, error) {
	b, err := v.AsBool()
	return NewBoolValue(b), err
}

func (v Value) AsNumber() (float64, error) {
	switch v.Type {
	case RawValue:
		var fval float64
		err := binary.Read(bytes.NewReader(v.raw), binary.BigEndian, &fval)
		return fval, err
	case StringValue:
		return strconv.ParseFloat(v.string, 64)
	case NumberValue:
		return v.number, nil
	case BoolValue:
		if v.bool {
			return 1, nil
		}
		return 0, nil
	case ColorValue:
		return float64(v.color.V), nil
	default:
		return 0, fmt.Errorf("invalid value type: %d", v.Type)
	}
}
func (v Value) AsNumberValue() (Value, error) {
	n, err := v.AsNumber()
	return NewNumberValue(n), err
}

func (v Value) AsRaw() ([]byte, error) {
	switch v.Type {
	case RawValue:
		return v.raw, nil
	case StringValue:
		return base64.StdEncoding.DecodeString(v.string)
	case NumberValue:
		buf := &bytes.Buffer{}
		err := binary.Write(buf, binary.BigEndian, v.number)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case BoolValue:
		if v.bool {
			return []byte{0}, nil
		}
		return []byte{1}, nil
	default:
		return nil, fmt.Errorf("invalid value type: %d", v.Type)
	}
}
func (v Value) AsRawValue() (Value, error) {
	r, err := v.AsRaw()
	return NewRawValue(r), err
}

func (v Value) AsColor() (Color, error) {
	switch v.Type {
	case StringValue:
		c := &Color{}
		_, err := toml.Decode(v.string, c)
		return *c, err
	case ColorValue:
		return v.color, nil
	default:
		return Color{}, fmt.Errorf("invalid value type: %d", v.Type)

	}
}

func (v Value) AsColorValue() (Value, error) {
	c, err := v.AsColor()
	return NewColorValue(c), err
}

func (v *Value) FromRaw(bs []byte) {
	if utf8.Valid(bs) {
		*v = Value{
			Type:   StringValue,
			string: string(bs),
		}
	}

	v.unStringify()
}

func (v *Value) unStringify() {
	if v.Type != StringValue {
		return
	}

	if newV, err := v.AsColorValue(); err == nil {
		*v = newV
		return
	}

	if newV, err := v.AsBoolValue(); err == nil {
		*v = newV
		return
	}

	if newV, err := v.AsNumberValue(); err == nil {
		*v = newV
		return
	}
}
