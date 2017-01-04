package json

import (
	"bytes"
	"database/sql/driver"
	"io"

	"github.com/sermodigital/errors"
	"github.com/sermodigital/pools"

	j8 "github.com/sermodigital/json1.8"
)

// Marshal custom implements encoding/j8.Marshal.
func Marshal(v interface{}) ([]byte, error) {
	b := pools.GetBuffer()
	if err := MarshalStream(b, v); err != nil {
		return nil, err
	}
	return b.UnsafeBytes(), nil
}

// MarshalIndent is like Marshal but indents the output.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return j8.MarshalIndent(v, prefix, indent)
}

// MarshalStream writes v to the provided io.Writer.
func MarshalStream(w io.Writer, v interface{}) error {
	return NewEncoder(w).Encode(v)
}

type Encoder struct {
	e *j8.Encoder
}

func NewEncoder(w io.Writer) Encoder {
	return Encoder{e: j8.NewEncoder(w)}
}

func (e Encoder) Encode(v interface{}) error {
	if vr, ok := v.(Validator); ok {
		if err := vr.Validate(); err != nil {
			return err
		}
	}
	return e.e.Encode(v)
}

// Validator is a type that can self-validate and report any errors.
// UnmarshalStream and Unmarshal will call the Validate method if the provided
// type implements Validator.
type Validator interface {
	Validate() error
}

// Validate simply validates the JSON and returns an error if the JSON is
// invalid. The underlying JSON is discarded.
func Validate(r io.Reader) error {
	return UnmarshalStream(r, &map[string]interface{}{})
}

// MaxReaderSize is the maximum permitted number of bytes that may be read from
// a stream.
const MaxReaderSize = 2e7 // 2 MB

const ErrTooLarge = errors.New("request was too large (max: 2 MB)")

type Decoder struct {
	d *j8.Decoder
	r *io.LimitedReader
}

func NewDecoder(r io.Reader) Decoder {
	lr := &io.LimitedReader{R: r, N: MaxReaderSize}
	return Decoder{d: j8.NewDecoder(lr), r: lr}
}

func (d Decoder) Decode(v interface{}) error {
	err := d.d.Decode(v)
	if err != nil {
		if d.r.N <= 0 {
			return ErrTooLarge
		}
		return &ErrInvalidJSON{err: err}
	}
	if vr, ok := v.(Validator); ok {
		return vr.Validate()
	}
	return nil
}

// UnmarshalStream custom implements encoding/j8.Unmarshal but reads from the
// provided io.Reader.
func UnmarshalStream(r io.Reader, v interface{}) error {
	return NewDecoder(r).Decode(v)
}

// Unmarshal custom implements encoding/j8.Unmarshal.
func Unmarshal(data []byte, v interface{}) error {
	return UnmarshalStream(bytes.NewReader(data), v)
}

type Number j8.Number

func (n Number) Float64() (float64, error) {
	return ((j8.Number)(n)).Float64()
}

func (n Number) Int64() (int64, error) {
	return ((j8.Number)(n)).Int64()
}

func (n Number) String() string {
	return ((j8.Number)(n)).String()
}

type RawMessage j8.RawMessage

func (m RawMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		// See: https://github.com/SermoDigital/sermocrm/issues/279
		return []byte{'n', 'u', 'l', 'l'}, nil
	}
	return ((*j8.RawMessage)(&m)).MarshalJSON()
}

func (m *RawMessage) UnmarshalJSON(data []byte) error {
	return ((*j8.RawMessage)(m)).UnmarshalJSON(data)
}

func (m RawMessage) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return []byte(m), nil
}

func (m *RawMessage) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return errors.Errorf("j8.RawMessage.Scan: wanted []byte, got %T", value)
	}
	*m = RawMessage(v)
	return nil
}

func (m RawMessage) IsNil() bool {
	return m == nil
}
