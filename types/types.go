package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

var (
	ErrNotFunction             = fmt.Errorf("not a function")
	ErrReturnNoParameters      = fmt.Errorf("returns no parameters")
	ErrTooManyReturnParameters = fmt.Errorf("too many return parameters")
	ErrSecondReturnParameters  = fmt.Errorf("the second return parameter must be error")
)

var Default = NewEmptyProvider()

type Provider interface {
	Register(kind string, fun interface{}) error
	Find(kind string) (reflect.Value, bool)
	Kind(config []byte) string
	ForEach(f func(kind string, fun reflect.Value))
}

type provider struct {
	functions map[string]reflect.Value
}

func NewEmptyProvider() Provider {
	return &provider{
		functions: map[string]reflect.Value{},
	}
}

func (h *provider) ForEach(f func(kind string, fun reflect.Value)) {
	keys := make([]string, 0, len(h.functions))
	for key := range h.functions {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		f(key, h.functions[key])
	}
}

func (h *provider) Register(kind string, v interface{}) error {
	if v == nil {
		return nil
	}
	fun := reflect.ValueOf(v)
	return h.register(kind, fun)
}

func (h *provider) register(kind string, fun reflect.Value) error {
	_, err := CheckFunc(fun)
	if err != nil {
		return fmt.Errorf("register %s: %v: %w", kind, fun, err)
	}

	h.functions[kind] = fun
	return nil
}

func (h *provider) Find(kind string) (reflect.Value, bool) {
	pairs, ok := h.functions[kind]
	return pairs, ok
}

func (h *provider) Kind(config []byte) string {
	var k struct {
		Kind string `json:"@kind"`
	}
	json.Unmarshal(config, &k)
	return k.Kind
}

func CheckFunc(funcValue reflect.Value) (reflect.Type, error) {
	if funcValue.Kind() != reflect.Func {
		return nil, ErrNotFunction
	}
	funcType := funcValue.Type()

	numOut := funcType.NumOut()
	switch numOut {
	case 0:
		return nil, ErrReturnNoParameters
	case 1:
	case 2:
		if !funcType.Out(1).Implements(errImplements) {
			return nil, ErrSecondReturnParameters
		}
	default:
		return nil, ErrTooManyReturnParameters
	}
	return funcType.Out(0), nil
}

var errImplements = reflect.TypeOf(new(error)).Elem()
