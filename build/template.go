package build

import (
	"reflect"
	"text/template"

	"github.com/wzshiming/namecase"
)

func getTypeName(t reflect.Type) string {
	return getName(t.Name())
}

func getKindName(typName string) string {
	return getName(typName)
}

func getName(name string) string {
	return namecase.ToUpperHumpInitialisms(name)
}

var tempType = template.Must(template.New("_").
	Parse(`

type {{.Type}} interface {
	is{{.Type}}()
	Component
}

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

`))

func tempKindGenType(prefix string, typ reflect.Type) string {
	return GenType(prefix, typ, getTypeName)
}

var tempKind = template.Must(template.New("_").
	Funcs(template.FuncMap{"genType": tempKindGenType}).
	Parse(`
// {{.Name}}{{.Ref.Name}} {{.Kind}}
{{genType .Name .Ref}}

func ({{.Name}}{{.Ref.Name}}) is{{.Type}}() {}
func ({{.Name}}{{.Ref.Name}}) isComponent() {}

// MarshalJSON returns m as the JSON encoding of m.
func (m {{.Name}}{{.Ref.Name}}) MarshalJSON() ([]byte, error) {
	const kind{{.Name}}{{.Ref.Name}} = "{{.Kind}}"
	type t {{.Name}}{{.Ref.Name}}
	data, err := json.Marshal(t(m))
	if err != nil {
		return nil, err
	}
	data = appendKV("@Kind", kind{{.Name}}{{.Ref.Name}}, data)
	return data, nil
}
`))

var tempConfig = `

type Component interface {
	isComponent()
	json.Marshaler
}

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

func appendKV(k, v string, data []byte) []byte{
	if data[0] == '{' {
		if len(data) == 2 {
			data = []byte(fmt.Sprintf("{\"%s\":%q}", k, v))
		} else {
			data = append([]byte(fmt.Sprintf("{\"%s\":%q,", k, v)), data[1:]...)
		}
	}
	return data
}
`

var tempHeader = `

// DO NOT EDIT! Code generated.

package %s

import (
	"encoding/json"
	"errors"
	"fmt"
)

`
