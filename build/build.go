package build

import (
	"bytes"
	"go/format"
	"log"
	"reflect"
)

type Build struct {
	pkg        string
	typeOnce   map[string]struct{}
	interfaces []interface{}
	types      []interface{}
}

func NewBuild(pkg string) *Build {
	b := &Build{
		pkg:      pkg,
		typeOnce: map[string]struct{}{},
	}
	return b
}

func (b *Build) Bytes() []byte {
	buf := &bytes.Buffer{}
	err := tpl.Execute(buf, map[string]interface{}{
		"Interfaces": b.interfaces,
		"Types":      b.types,
		"Package":    b.pkg,
		"Key":        "@Kind",
	})
	if err != nil {
		log.Printf("[ERROR] kind %s", err)
	}
	src := buf.Bytes()
	f, err := format.Source(src)
	if err != nil {
		return src
	}
	return f
}

func (b *Build) String() string {
	return string(b.Bytes())
}

func (b *Build) Add(kind string, t reflect.Type, fun reflect.Value) {

	typeName := getTypeName(t)
	name := getKindName(kind)
	if _, ok := b.typeOnce[typeName]; !ok {
		b.interfaces = append(b.interfaces, map[string]interface{}{
			"Type": typeName,
			"Out":  t,
		})
		b.typeOnce[typeName] = struct{}{}
	}

	funTyp := fun.Type()

	num := funTyp.NumIn()

	refType := reflect.TypeOf(struct{}{})
	for i := 0; i != num; i++ {
		in := funTyp.In(i)
		k := in.Kind()
		if k == reflect.Ptr {
			in = in.Elem()
			k = in.Kind()
		}
		if in.Kind() == reflect.Struct {
			refType = in
			break
		}
	}

	b.types = append(b.types, map[string]interface{}{
		"Name": name,
		"Type": typeName,
		"Kind": kind,
		"Ref":  refType,
	})
}
