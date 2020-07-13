// DO NOT EDIT! Code generated.

package {{.Package}}

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"

    "github.com/wzshiming/funcfg/types"
    "github.com/wzshiming/funcfg/unmarshaler"
)

// ========= Begin Common =========
//

var kindKey = `{{.Key}}`

var provider = types.NewEmptyProvider()

// Unmarshal parses the encoded data and stores the result
func Unmarshal(config []byte, v interface{}) error {
    u := unmarshaler.Unmarshaler{
        Ctx:      context.Background(),
        Provider: provider,
    }
    return u.Unmarshal(config, v)
}

// Marshal returns the encoding of v
func Marshal(v interface{}) ([]byte, error) {
    return json.Marshal(v)
}

// Component is basic definition of Component
type Component interface {
    isComponent()
}

// RawComponent is store raw bytes of Component
type RawComponent []byte

func (RawComponent) isComponent() {}

// MarshalJSON returns m as the JSON encoding of m.
func (m RawComponent) MarshalJSON() ([]byte, error) {
    if m == nil {
        return []byte("null"), nil
    }
    return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *RawComponent) UnmarshalJSON(data []byte) error {
    if m == nil {
        return errors.New("RawComponent: UnmarshalJSON on nil pointer")
    }
    *m = append((*m)[:0], data...)
    return nil
}

func prepend(k, v string, data []byte) []byte {
    if data[0] == '{' {
        if len(data) == 2 {
            data = []byte(fmt.Sprintf(`{%q:%q}`, k, v))
        } else {
            data = append([]byte(fmt.Sprintf(`{%q:%q,`, k, v)), data[1:]...)
        }
    }
    return data
}

//
// ========= End Common =========

{{range .Types}}
// ========= Begin {{.Kind}} type =========
//

const kind{{.Name}}{{.Ref.Name}} = `{{.Kind}}`

// {{.Name}}{{.Ref.Name}} {{.Kind}}
{{genType .Name .Type .Ref}}

func init() {
	_ = provider.Register(
		kind{{.Name}}{{.Ref.Name}},
		func(r *{{.Name}}{{.Ref.Name}}) {{.Type}} { return r },
	)
}

func ({{.Name}}{{.Ref.Name}}) is{{.Type}}() {}
func ({{.Name}}{{.Ref.Name}}) isComponent() {}

// MarshalJSON returns m as the JSON encoding of m.
func (m {{.Name}}{{.Ref.Name}}) MarshalJSON() ([]byte, error) {
	type t {{.Name}}{{.Ref.Name}}
	data, err := json.Marshal(t(m))
	if err != nil {
		return nil, err
	}
	data = prepend(kindKey, kind{{.Name}}{{.Ref.Name}}, data)
	return data, nil
}

//
// ========= End {{.Kind}} type =========
{{end}}

{{range .Interfaces}}
// ========= Begin {{.Out}} interface =========
//

// {{.Type}} {{.Out}}
type {{.Type}} interface {
	is{{.Type}}()
	Component
}

// Raw{{.Type}} is store raw bytes of {{.Type}}
type Raw{{.Type}} []byte

func (Raw{{.Type}}) is{{.Type}}() {}
func (Raw{{.Type}}) isComponent() {}

// MarshalJSON returns m as the JSON encoding of m.
func (m Raw{{.Type}}) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *Raw{{.Type}}) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("Raw{{.Type}}: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[:0], data...)
	return nil
}

//
// ========= End {{.Out}} interface =========
{{end}}
