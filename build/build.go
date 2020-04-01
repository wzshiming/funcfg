package build

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"reflect"
)

type Build struct {
	pkg      string
	buf      *bytes.Buffer
	typeOnce map[string]struct{}
}

func NewBuild(pkg string) *Build {
	b := &Build{
		pkg:      pkg,
		buf:      &bytes.Buffer{},
		typeOnce: map[string]struct{}{},
	}
	b.init()
	return b
}

func (b *Build) init() {
	fmt.Fprintf(b.buf, tempHeader, b.pkg)

	fmt.Fprint(b.buf, tempConfig)

}

func (b *Build) Bytes() []byte {

	src := b.buf.Bytes()

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
		err := tempType.Execute(b.buf, map[string]string{
			"Type": typeName,
		})
		if err != nil {
			log.Printf("[ERROR] type %s", err)
		}
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

	err := tempKind.Execute(b.buf, map[string]interface{}{
		"Name": name,
		"Type": typeName,
		"Kind": kind,
		"Out":  t,
		"Ref":  refType,
	})
	if err != nil {
		log.Printf("[ERROR] kind %s", err)
	}
}
