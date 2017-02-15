// +build protobuf

package json

import (
	"bytes"
	"reflect"
	"strconv"

	"github.com/sermodigital/errors"
)

func (m RawMessage) Equal(m2 RawMessage) bool {
	return bytes.Equal(m, m2)
}

func (m RawMessage) Size() int { return len(m) }

func (m RawMessage) MarshalTo(data []byte) (int, error) {
	return copy(data, m), nil
}

func (m RawMessage) Marshal() ([]byte, error) {
	return m.MarshalJSON()
}

func (m *RawMessage) Unmarshal(data []byte) error {
	return m.UnmarshalJSON(data)
}

func (m RawMessage) Compare(m2 RawMessage) int {
	return bytes.Compare(m, m2)
}

type randy interface {
	Intn(n int) int
}

func NewPopulatedRawMessage(r randy, _ ...bool) *RawMessage {
	m := make(RawMessage, r.Intn(255))
	for i := range m {
		m[i] = byte(r.Intn(255))
	}
	return &m
}

// pbKey is a bit of a hack to skirt some gogoprotobuf weirdness. How
// gogoprotobuf currently works is it uses the return value of Size to create
// a buffer it then marshals the data structure into. It would be unfortunate
// to have to marshal the same structure twice--once to determine its size and
// another to copy it somewhere--so we store the result of the marshaling as
// a field in the JSON object. This name should be random enough that we won't
// run into any collisions.
//
// If Marshal returns an error this field will contain the error. If Marshal
// returns nil this field will contain the []byte result.
//
// This should only be used for protobuf's Size issue and nothing else. I
// regret having to even write this code, but we use JSON a lot in entities
// and models.
const pbKey = "--FZjwSgQwZDqqDxNj9x0a_!_protobuf_!_ZSRvCydU2eBZzicdw3oW--"

func (j JSON) Equal(j2 JSON) bool {
	if j == nil {
		return j2 == nil
	}

	if reflect.DeepEqual(j, j2) {
		return true
	}

	// Sometimes reflect.DeepEqual isn't a good test because, e.g., Unmarshal
	// will convert integers to floats, so even though the numerical values
	// maybe equal, reflect.DeepEqual returns false as their types differ.
	//
	// Recursively check to see either has previously been umarshaled. If not,
	// marshal it so our types match.
	bm1 := beenmarshaled(j)
	bm2 := beenmarshaled(j2)

	// Short circuit if both have been marshaled.
	if bm1 && bm2 {
		return false
	}

	var b1, b2 []byte
	var err error

	if !bm1 {
		b1, err = j.Marshal()
		if err != nil {
			return false
		}
	}
	if !bm2 {
		b2, err = j2.Marshal()
		if err != nil {
			return false
		}
	}

	mj1, mj2 := j, j2
	if b1 != nil {
		mj1 = make(JSON)
		err = mj1.Unmarshal(b1)
		if err != nil {
			return false
		}
	}
	if b2 != nil {
		mj2 = make(JSON)
		err = mj2.Unmarshal(b2)
		if err != nil {
			return false
		}
	}
	return reflect.DeepEqual(mj1, mj2)
}

// beenmarshaled returns true if v only consists of the types json.Unmarshal
// uses, i.e. float64, bool, string, nil. It recursively descends v checking all
// container types.
func beenmarshaled(v interface{}) bool {
	switch t := v.(type) {
	case float64, bool, string, nil:
		return true
	case []interface{}:
		for v := range t {
			if !beenmarshaled(v) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		for _, v := range t {
			if !beenmarshaled(v) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (j JSON) Size() int {
	if len(j) == 0 {
		return len("{}")
	}
	// Cached the marshaled version.
	v, ok := j[pbKey]
	if ok {
		b, ok := v.([]byte)
		if ok {
			return len(b)
		}

		// TODO: panic? Since Marshal(j) (called below) should never return an
		// error unless j contains something Go cannot marshal like a channel.
		err, ok := v.(error)
		if !ok {
			panic("should be unreachable")
		}
		return 0
	}

	// No cache, so do it ourselves.
	b, err := Marshal(j)
	if err != nil {
		j[pbKey] = err
		return 0
	}
	j[pbKey] = b
	return len(b)
}

func (j JSON) Marshal() ([]byte, error) {
	if len(j) == 0 {
		return []byte{'{', '}'}, nil
	}
	v, ok := j[pbKey]
	if !ok {
		return Marshal(j)
	}
	delete(j, pbKey)
	switch t := v.(type) {
	case error:
		return nil, t
	case []byte:
		return t, nil
	default:
		return nil, errors.Errorf("invalid type: %T", v)
	}
}

func (j *JSON) MarshalTo(data []byte) (int, error) {
	b, err := j.Marshal()
	if err != nil {
		return 0, err
	}
	return copy(data, b), nil
}

func (j *JSON) Unmarshal(data []byte) error {
	return Unmarshal(data, j)
}

func NewPopulatedJSON(r randy, _ ...bool) *JSON {
	j := make(JSON)
	for i := 0; i < r.Intn(255); i++ {
		// idk
		j[strconv.Itoa(i)] = r.Intn(1<<32 - 1)
	}
	return &j
}
