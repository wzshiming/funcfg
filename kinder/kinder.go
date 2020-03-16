package kinder

import (
	"encoding/json"
)

func Kind(config []byte) string {
	var k struct {
		Kind string `json:"@kind"`
	}
	json.Unmarshal(config, &k)
	return k.Kind
}
