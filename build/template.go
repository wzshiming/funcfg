package build

import (
	"reflect"
	"strings"
	"text/template"

	"github.com/wzshiming/namecase"
)

func getTypeName(t reflect.Type) string {
	name := t.String()
	i := strings.LastIndex(name, ".")
	if i != -1 && strings.ToLower(name[:i]) == strings.ToLower(name[i+1:]) {
		return getName(name[i+1:])
	}
	return getName(name)
}

func getKindName(name string) string {
	i := strings.LastIndex(name, ".")
	if i != -1 && strings.HasSuffix(strings.ToLower(name[:i]), strings.ToLower(name[i+1:])) {
		return getName(name[:i])
	}
	return getName(name)
}

func getName(name string) string {
	return namecase.ToUpperHumpInitialisms(name)
}

var tempType = template.Must(template.New("_").
	Parse(`

// ========= Begin {{.Out}} =========
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
// ========= End {{.Out}} =========
`))

func tempKindGenType(prefix, t string, typ reflect.Type) string {
	return GenType(prefix, typ, t, getTypeName)
}

var tempKind = template.Must(template.New("_").
	Funcs(template.FuncMap{"genType": tempKindGenType}).
	Parse(`

// ========= Begin {{.Kind}} =========
//

const kind{{.Name}}{{.Ref.Name}} = "{{.Kind}}"

// {{.Name}}{{.Ref.Name}} {{.Kind}}
{{genType .Name .Type .Ref}}

func init() {
	_ = defTypes.Register(
		kind{{.Name}}{{.Ref.Name}}, 
		func(r *{{.Name}}{{.Ref.Name}}) {{.Type}} {
			return r
		},
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
	data = appendKV(kindKey, kind{{.Name}}{{.Ref.Name}}, data)
	return data, nil
}

//
// ========= End {{.Kind}} =========

`))

var tempConfig = `

// ========= Begin Common =========
//

var kindKey = "@Kind"

var defTypes = types.NewTypes()

// Unmarshal parses the encoded data and stores the result
func Unmarshal(config []byte, v interface{}) error {
	u := unmarshaler.Unmarshaler{
		Ctx:  context.Background(),
		Get:  defTypes.Get,
		Kind: kinder.Kind,
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

func appendKV(k, v string, data []byte) []byte {
	if data[0] == '{' {
		if len(data) == 2 {
			data = []byte(fmt.Sprintf("{\"%s\":%q}", k, v))
		} else {
			data = append([]byte(fmt.Sprintf("{\"%s\":%q,", k, v)), data[1:]...)
		}
	}
	return data
}

//
// ========= End Common =========

`

var tempHeader = `
// DO NOT EDIT! Code generated.

package %s

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/wzshiming/funcfg/kinder"
	"github.com/wzshiming/funcfg/types"
	"github.com/wzshiming/funcfg/unmarshaler"
)

`
