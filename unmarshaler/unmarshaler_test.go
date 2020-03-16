package unmarshaler

import (
	"context"
	"reflect"
	"testing"

	"github.com/wzshiming/funcfg/kinder"
	"github.com/wzshiming/funcfg/types"
	"github.com/wzshiming/inject"
)

type Config struct {
	Name string
}

func (c Config) M() {}

type Adapter interface {
	M()
}

func TestUnmarshaler(t *testing.T) {

	ctx := context.Background()
	type args struct {
		ctx    context.Context
		config []byte
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			args: args{ctx, []byte(`{"@Kind":"hello1"}`)},
			want: &Config{"hello1"},
		},
		{
			args: args{ctx, []byte(`[{"@Kind":"hello1"},{"@Kind":"hello2"}]`)},
			want: []*Config{{"hello1"}, {"hello2"}},
		},
		{
			args: args{ctx, []byte(`{"A":{"@Kind":"hello1"}}`)},
			want: &struct{ A *Config }{&Config{"hello1"}},
		},
		{
			args: args{ctx, []byte(`{"A":{"@Kind":"hello1"},"B":[{"@Kind":"hello2"},{"@Kind":"hello3"}]}`)},
			want: &struct {
				A *Config
				B []*Config
			}{&Config{"hello1"}, []*Config{{"hello2"}, {"hello3"}}},
		},
		{
			args: args{ctx, []byte(`{"name":{"@Kind":"hello1"},"name2":{"@Kind":"hello2"}}`)},
			want: &map[string]*Config{"name": {"hello1"}, "name2": {"hello2"}},
		},

		{
			args: args{ctx, []byte(`{"@Kind":"hello1"}`)},
			want: Config{"hello1"},
		},
		{
			args: args{ctx, []byte(`[{"@Kind":"hello1"},{"@Kind":"hello2"}]`)},
			want: []Config{{"hello1"}, {"hello2"}},
		},
		{
			args: args{ctx, []byte(`{"A":{"@Kind":"hello1"}}`)},
			want: struct{ A Config }{Config{"hello1"}},
		},
		{
			args: args{ctx, []byte(`{"A":{"@Kind":"hello1"},"B":[{"@Kind":"hello2"},{"@Kind":"hello3"}]}`)},
			want: struct {
				A Config
				B []Config
			}{Config{"hello1"}, []Config{{"hello2"}, {"hello3"}}},
		},
		{
			args: args{ctx, []byte(`{"name":{"@Kind":"hello1"},"name2":{"@Kind":"hello2"}}`)},
			want: map[string]Config{"name": {"hello1"}, "name2": {"hello2"}},
		},
	}

	fun := []interface{}{
		func(name string, config []byte) (Config, error) {
			return Config{Name: name}, nil
		},
		func(name string, config []byte) (*Config, error) {
			return &Config{Name: name}, nil
		},
		func(name string, config []byte) (Adapter, error) {
			return Config{Name: name}, nil
		},
		func(name string, config []byte) (Adapter, error) {
			return &Config{Name: name}, nil
		},
	}

	for _, f := range fun {
		err := types.Register("hello1", f)
		if err != nil {
			t.Fatal(err)
		}

		err = types.Register("hello2", f)
		if err != nil {
			t.Fatal(err)
		}

		err = types.Register("hello3", f)
		if err != nil {
			t.Fatal(err)
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				gotValue := reflect.New(reflect.TypeOf(tt.want))
				u := Unmarshaler{
					Ctx:  tt.args.ctx,
					Get:  types.Get,
					Kind: kinder.Kind,
				}
				err := u.Unmarshal(tt.args.config, gotValue.Interface())
				if (err != nil) != tt.wantErr {
					t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				}
				if err != nil {
					return
				}

				got := gotValue.Elem().Interface()
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Unmarshal() got = %#v, want %#v", got, tt.want)
				}
			})
		}
	}
}

func Test_indirectTo(t *testing.T) {
	type args struct {
		v  reflect.Value
		to reflect.Type
	}
	tests := []struct {
		name    string
		args    args
		want    reflect.Value
		wantErr bool
	}{
		{
			args: args{v: reflect.ValueOf(Config{Name: ""}), to: reflect.TypeOf(Config{})},
			want: reflect.ValueOf(Config{Name: ""}),
		},
		{
			args: args{v: reflect.ValueOf(Config{Name: "hello1"}), to: reflect.TypeOf(Config{})},
			want: reflect.ValueOf(Config{Name: "hello1"}),
		},
		{
			args: args{v: reflect.ValueOf(&Config{Name: "hello1"}), to: reflect.TypeOf(Config{})},
			want: reflect.ValueOf(Config{Name: "hello1"}),
		},
		{
			args: args{v: reflect.ValueOf(&Config{Name: "hello1"}), to: reflect.TypeOf(&Config{})},
			want: reflect.ValueOf(&Config{Name: "hello1"}),
		},
		{
			args: args{v: reflect.ValueOf(Config{Name: "hello1"}), to: reflect.TypeOf(&Config{})},
			want: reflect.ValueOf(&Config{Name: "hello1"}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := indirectTo(tt.args.v, tt.args.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("indirectTo() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got.Interface(), tt.want.Interface()) {
				t.Errorf("indirectTo() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_callWithInject(t *testing.T) {
	type args struct {
		fun reflect.Value
		inj *inject.Injector
	}
	tests := []struct {
		name    string
		args    args
		want    reflect.Value
		wantErr bool
	}{
		{
			args: args{fun: reflect.ValueOf(func() (Config, error) { return Config{Name: "hello1"}, nil })},
			want: reflect.ValueOf(Config{Name: "hello1"}),
		},
		{
			args: args{fun: reflect.ValueOf(func() (*Config, error) { return &Config{Name: "hello1"}, nil })},
			want: reflect.ValueOf(&Config{Name: "hello1"}),
		},
		{
			args: args{fun: reflect.ValueOf(func() (Adapter, error) { return Config{Name: "hello1"}, nil })},
			want: reflect.ValueOf(Config{Name: "hello1"}),
		},
		{
			args: args{fun: reflect.ValueOf(func() (Adapter, error) { return &Config{Name: "hello1"}, nil })},
			want: reflect.ValueOf(&Config{Name: "hello1"}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := callWithInject(tt.args.fun, tt.args.inj)
			if (err != nil) != tt.wantErr {
				t.Errorf("callWithInject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Interface(), tt.want.Interface()) {
				t.Errorf("callWithInject() got = %v, want %v", got, tt.want)
			}
		})
	}
}
