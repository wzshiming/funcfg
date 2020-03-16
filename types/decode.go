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
var stdTypes = newTypes()

func Register(kind string, fun interface{}) error {
	return stdTypes.Register(kind, fun)
}

func RegisterWithPkg(kind string, fun interface{}) error {
	return stdTypes.RegisterWithPkg(kind, fun)
}

func Get(kind string) (reflect.Value, bool) {
	return stdTypes.Get(kind)
}

func ForEach(f func(kind string, fun reflect.Value)) {
	stdTypes.ForEach(f)
}

type types struct {
	funcs map[string]reflect.Value
}

func newTypes() *types {
	return &types{
		funcs: map[string]reflect.Value{},
	}
}

func (h *types) ForEach(f func(kind string, fun reflect.Value)) {
	keys := make([]string, 0, len(h.funcs))
	for key := range h.funcs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		f(key, h.funcs[key])
	}
}

func (h *types) Register(kind string, v interface{}) error {
	fun := reflect.ValueOf(v)
	return h.register(kind, fun)
}

func (h *types) RegisterWithPkg(kind string, v interface{}) error {
	fun := reflect.ValueOf(v)
	kind = fun.Type().PkgPath() + "@" + kind
	return h.register(kind, fun)
}

func (h *types) register(kind string, fun reflect.Value) error {

	err := checkFunc(fun)
	if err != nil {
		err = fmt.Errorf("register %s: %v: %w", kind, fun, err)
		return err
	}

	h.funcs[kind] = fun
	return nil
}

func (h *types) Get(kind string) (reflect.Value, bool) {
	pairs, ok := h.funcs[kind]
	return pairs, ok
}

func checkFunc(funcValue reflect.Value) error {
	if funcValue.Kind() != reflect.Func {
		return ErrNotFunction
	}
	funcType := funcValue.Type()

	numOut := funcType.NumOut()
	switch numOut {
	case 0:
		return ErrReturnNoParameters
	case 1:
	case 2:
		if !funcType.Out(1).Implements(errImplements) {
			return ErrSecondReturnParameters
		}
	default:
		return ErrTooManyReturnParameters
	}
	return nil
}

var errImplements = reflect.TypeOf(new(error)).Elem()
