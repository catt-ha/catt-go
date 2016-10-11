package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type ValueType uint8

const (
	RawValue ValueType = iota
	StringValue
	NumberValue
	BoolValue
)

type Value struct {
	Type   ValueType
	number float64
	string string
	bool   bool
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

func (v Value) AsString() (string, error) {
	switch v.Type {
	case RawValue:
		if utf8.Valid(v.raw) {
			return string(v.raw), nil
		}
		return "", errors.New("invalid utf8 byte slice")
	case StringValue:
		return v.string, nil
	case NumberValue:
		return fmt.Sprintf("%v", v.number), nil
	case BoolValue:
		if v.bool {
			return "ON", nil
		}
		return "OFF", nil
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
		return []byte(v.string), nil
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

	if newV, err := v.AsBoolValue(); err == nil {
		*v = newV
		return
	}

	if newV, err := v.AsNumberValue(); err == nil {
		*v = newV
		return
	}
}
