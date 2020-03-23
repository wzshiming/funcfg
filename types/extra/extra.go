package extra

import (
	"fmt"
	"reflect"

	"github.com/wzshiming/funcfg/types"
)

// RegisterWithBuildFunc
// universal kind method registration. result is the function return type. last args must is interface{}.
// register func(arg1, arg2, argN..., result) (error) => func(arg1, arg2, argN...) (result, error)
func RegisterWithBuildFunc(kind string, function, result interface{}) error {
	fun, err := BuildFunc(function, result)
	if err != nil {
		return err
	}
	return types.Register(kind, fun.Interface())
}

func BuildFunc(function, result interface{}) (reflect.Value, error) {
	buildFun := reflect.ValueOf(function)
	buildFunTyp := buildFun.Type()
	if buildFunTyp.Kind() != reflect.Func {
		return reflect.Value{}, fmt.Errorf("must is a function")
	}
	if buildFunTyp.IsVariadic() {
		return reflect.Value{}, fmt.Errorf("last args must is interface{}")
	}
	numIn := buildFunTyp.NumIn()
	if numIn == 0 {
		return reflect.Value{}, fmt.Errorf("args is not nil")
	}
	lastArg := buildFunTyp.In(numIn - 1)
	if lastArg.Kind() != reflect.Interface || lastArg.NumMethod() != 0 {
		return reflect.Value{}, fmt.Errorf("last args must is interface{}")
	}

	in := make([]reflect.Type, 0, numIn-1)
	for i := 0; i != numIn-1; i++ {
		in = append(in, buildFunTyp.In(i))
	}

	numOut := buildFunTyp.NumOut()
	out := make([]reflect.Type, 0, numOut+1)
	resultTyp := reflect.TypeOf(result).Elem()
	out = append(out, resultTyp)
	for i := 0; i != numOut; i++ {
		out = append(out, buildFunTyp.Out(i))
	}
	typ := reflect.FuncOf(in, out, false)
	fun := reflect.MakeFunc(typ, func(args []reflect.Value) (results []reflect.Value) {
		result := reflect.New(resultTyp)
		results = buildFun.Call(append(args, result))
		return append([]reflect.Value{result.Elem()}, results...)
	})
	return fun, nil
}
