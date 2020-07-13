package funcfg

import (
	"context"

	"github.com/wzshiming/funcfg/types"
	"github.com/wzshiming/funcfg/unmarshaler"
)

func Unmarshal(config []byte, v interface{}) error {
	u := unmarshaler.Unmarshaler{
		Ctx:      context.Background(),
		Provider: types.Default,
	}
	return u.Unmarshal(config, v)
}
