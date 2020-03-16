package funcfg

import (
	"context"

	"github.com/wzshiming/funcfg/kinder"
	"github.com/wzshiming/funcfg/types"
	"github.com/wzshiming/funcfg/unmarshaler"
)

func Unmarshal(config []byte, v interface{}) error {
	u := unmarshaler.Unmarshaler{
		Ctx:  context.Background(),
		Get:  types.Get,
		Kind: kinder.Kind,
	}
	return u.Unmarshal(config, v)
}
