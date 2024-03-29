package unmarshaler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/wzshiming/funcfg/types"
	"github.com/wzshiming/inject"
)

var (
	ErrParsedParameter  = fmt.Errorf("the parsed parameter must be a pointer")
	ErrMustBeAssignable = fmt.Errorf("must be assignable")
	ErrFormat           = fmt.Errorf("format error")
	ErrIsInvalid        = fmt.Errorf("is invalid")
)

type Unmarshaler struct {
	Ctx      context.Context
	Inject   *inject.Injector
	Provider types.Provider
}

func (d *Unmarshaler) Unmarshal(config []byte, i interface{}) error {
	v := reflect.ValueOf(i)
	return d.decode(config, v)
}

func (d *Unmarshaler) decodeArray(config []byte, v reflect.Value) error {
	tmp := []json.RawMessage{}
	err := json.Unmarshal(config, &tmp)
	if err != nil {
		return err
	}
	l := len(tmp)
	v.Set(reflect.Zero(v.Type()))
	if l > v.Len() {
		l = v.Len()
	}
	for i := 0; i != l; i++ {
		err := d.decode(tmp[i], v.Index(i).Addr())
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Unmarshaler) decodeSlice(config []byte, v reflect.Value) error {
	tmp := []json.RawMessage{}
	err := json.Unmarshal(config, &tmp)
	if err != nil {
		return err
	}
	v.Set(reflect.MakeSlice(v.Type(), len(tmp), len(tmp)))
	for i := 0; i != len(tmp); i++ {
		err := d.decode(tmp[i], v.Index(i).Addr())
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Unmarshaler) decodeMap(config []byte, v reflect.Value) error {
	tmp := map[string]json.RawMessage{}
	err := json.Unmarshal(config, &tmp)
	if err != nil {
		return err
	}
	typ := v.Type()
	v.Set(reflect.MakeMapWithSize(typ, len(tmp)))
	for key, raw := range tmp {
		val := reflect.New(typ.Elem())
		err := d.decode(raw, val)
		if err != nil {
			return err
		}
		v.SetMapIndex(reflect.ValueOf(key), val.Elem())
	}
	return nil
}

func (d *Unmarshaler) decodeStruct(config []byte, v reflect.Value) error {
	tmp := map[string]json.RawMessage{}
	err := json.Unmarshal(config, &tmp)
	if err != nil {
		return err
	}
	for k, v := range tmp {
		key := strings.ToLower(k)
		if key != k {
			tmp[key] = v
		}
	}

	typ := v.Type()
	v.Set(reflect.Zero(typ))
	num := typ.NumField()
	for i := 0; i != num; i++ {
		f := typ.Field(i)
		name := strings.ToLower(f.Name)
		if value, ok := f.Tag.Lookup("json"); ok {
			n := strings.Split(value, ",")
			if n[0] != "" {
				name = n[0]
				name = strings.ToLower(name)
			}
			for _, arg := range n[1:] {
				switch arg {
				case "string":
					d, _ := strconv.Unquote(string(tmp[name]))
					tmp[name] = json.RawMessage(d)
				}
			}
		}

		if c, ok := tmp[name]; ok {
			field := v.Field(i)
			field.Set(reflect.Zero(f.Type))
			err := d.decode(c, field.Addr())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Unmarshaler) decodeOther(config []byte, v reflect.Value) error {
	switch config[0] {
	case '[':
		v := indirectElem(v)
		switch v.Kind() {
		case reflect.Array:
			return d.decodeArray(config, v)
		case reflect.Slice:
			return d.decodeSlice(config, v)
		}
	case '{':
		v := indirectElem(v)
		switch v.Kind() {
		case reflect.Map:
			return d.decodeMap(config, v)
		case reflect.Struct:
			return d.decodeStruct(config, v)
		}
	}

	err := json.Unmarshal(config, v.Interface())
	if err != nil {
		return err
	}

	return nil
}

func (d *Unmarshaler) decode(config []byte, value reflect.Value) error {
	if len(config) == 0 {
		return ErrFormat
	}
	if value.Kind() != reflect.Ptr {
		return ErrParsedParameter
	}

	if !value.Elem().CanSet() {
		return ErrMustBeAssignable
	}

	config = bytes.TrimSpace(config)
	if config[0] != '{' {
		return d.decodeOther(config, value)
	}

	kind := d.Provider.Kind(config)
	if kind == "" {
		return d.decodeOther(config, value)
	}

	fun, ok := d.Provider.Find(kind)
	if !ok {
		return fmt.Errorf("not found %q in provider", kind)
	}

	err := d.unmarshalKind(fun, kind, config, value)
	if err != nil {
		return err
	}

	return nil
}

func (d *Unmarshaler) unmarshalKind(fun reflect.Value, kind string, config []byte, value reflect.Value) error {
	inj := inject.NewInjector(d.Inject)
	args := []interface{}{d, &d.Ctx, inj, kind, config, &value}
	for _, arg := range args {
		err := inj.Map(reflect.ValueOf(arg))
		if err != nil {
			return err
		}
	}
	funType := fun.Type()
	num := funType.NumIn()
	for i := 0; i != num; i++ {
		in := funType.In(i)
		switch in.Kind() {
		case reflect.Slice:
			if in.Elem().Kind() == reflect.Uint8 {
				continue
			}
		case reflect.Array, reflect.Struct, reflect.Map:
		case reflect.Ptr:
			if in.Elem().Kind() != reflect.Struct {
				continue
			}
		default:
			continue
		}

		n := reflect.New(in)
		err := d.decodeOther(config, n)
		if err != nil {
			return err
		}
		err = inj.Map(n)
		if err != nil {
			return err
		}
	}

	r, err := callWithInject(fun, inj)
	if err != nil {
		return err
	}

	value = indirectElem(value)
	r, err = indirectTo(r, value.Type())
	if err != nil {
		return fmt.Errorf("%q %w", config, err)
	}

	err = setValue(value, r)
	if err != nil {
		return err
	}
	return nil
}

func setValue(value reflect.Value, r reflect.Value) error {
	switch value.Kind() {
	case reflect.Interface:
		if !r.Type().Implements(value.Type()) {
			return fmt.Errorf("value of %s is not assignable to %s", r.Type(), value.Type())
		}
	default:
		if !r.Type().AssignableTo(value.Type()) {
			return fmt.Errorf("value of %s is not assignable to %s", r.Type(), value.Type())
		}
	}

	value.Set(r)
	return nil
}

func callWithInject(fun reflect.Value, inj *inject.Injector) (reflect.Value, error) {
	if inj == nil {
		inj = inject.NewInjector(nil)
	}
	ret, err := inj.Call(fun)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("call error: %w", err)
	}
	if len(ret) == 2 {
		errInterface := ret[1].Interface()
		if errInterface != nil {
			err, ok := errInterface.(error)
			if !ok {
				panic("this should not be performed until")
			}
			if err != nil {
				return reflect.Value{}, err
			}
		}
	}
	r := ret[0]
	if r.Kind() == reflect.Invalid {
		return reflect.Value{}, fmt.Errorf("%s error return invalid", fun.String())
	}
	return r, nil
}

func indirectTo(v reflect.Value, to reflect.Type) (r reflect.Value, err error) {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if !v.IsValid() {
		return v, ErrIsInvalid
	}
	if v.Type().AssignableTo(to) {
		return v, nil
	}
	if v.Kind() != reflect.Ptr {
		nv := reflect.New(v.Type())
		nv.Elem().Set(v)
		if nv.Type().AssignableTo(to) {
			return nv, nil
		}
		return reflect.Value{}, fmt.Errorf("can't indirect %v to %v", v.Type(), to)
	}
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	return indirectTo(v.Elem(), to)
}

func indirectElem(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr {
		return v
	}
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	return indirectElem(v.Elem())
}
