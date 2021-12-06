package build

import (
	_ "embed"
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

func tempKindGenType(prefix, t string, typ reflect.Type) string {
	return GenType(prefix, typ, t, getTypeName)
}

var (
	//go:embed bing.go.tpl
	bingGoTpl string
	tpl       = template.Must(template.New("_").Funcs(template.FuncMap{"genType": tempKindGenType}).Parse(bingGoTpl))
)
