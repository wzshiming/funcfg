package types

import (
	"fmt"
	"reflect"
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

func Get(kind string) (reflect.Value, bool) {
	return stdTypes.Get(kind)
}

type types struct {
	funcs map[string]reflect.Value
}

func newTypes() *types {
	return &types{
		funcs: map[string]reflect.Value{},
	}
}

func (h *types) Register(kind string, v interface{}) error {
	fun := reflect.ValueOf(v)
	err := checkFunc(fun)
	if err != nil {
		err = fmt.Errorf("register %s: %v: %w", kind, v, err)
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
