package types

import (
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
var stdTypes = NewTypes()

func Register(kind string, fun interface{}) error {
	return stdTypes.Register(kind, fun)
}

func Get(kind string) (reflect.Value, bool) {
	return stdTypes.Get(kind)
}

func ForEach(f func(kind string, fun reflect.Value)) {
	stdTypes.ForEach(f)
}

type Types struct {
	funcs map[string]reflect.Value
}

func NewTypes() *Types {
	return &Types{
		funcs: map[string]reflect.Value{},
	}
}

func (h *Types) ForEach(f func(kind string, fun reflect.Value)) {
	keys := make([]string, 0, len(h.funcs))
	for key := range h.funcs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		f(key, h.funcs[key])
	}
}

func (h *Types) Register(kind string, v interface{}) error {
	fun := reflect.ValueOf(v)
	return h.register(kind, fun)
}

func (h *Types) register(kind string, fun reflect.Value) error {

	_, err := CheckFunc(fun)
	if err != nil {
		return fmt.Errorf("register %s: %v: %w", kind, fun, err)
	}

	h.funcs[kind] = fun
	return nil
}

func (h *Types) Get(kind string) (reflect.Value, bool) {
	pairs, ok := h.funcs[kind]
	return pairs, ok
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
