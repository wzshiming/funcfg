package bind

import (
	"io/ioutil"
	"os"
	"reflect"

	"github.com/wzshiming/funcfg/build"
	"github.com/wzshiming/funcfg/types"
)

func Bind(out string) error {
	if out == "" {
		b := build.NewBuild("bind")
		types.Default.ForEach(func(kind string, fun reflect.Value) {
			b.Add(kind, fun.Type().Out(0), fun)
		})
		os.Stdout.Write(b.Bytes())
		return nil
	}
	b := build.NewBuild(out)
	types.Default.ForEach(func(kind string, fun reflect.Value) {
		b.Add(kind, fun.Type().Out(0), fun)
	})
	return ioutil.WriteFile(out+".go", b.Bytes(), 0655)
}
