package config

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
)

func MustNew(defaultValues map[string]any) *T {
	c, err := New(defaultValues)
	if err != nil {
		panic(err)
	}
	return c
}

func New(defaultValues map[string]any) (*T, error) {
	k := koanf.New(".")
	const (
		envPrefix = "INIT__"
	)

	_ = k.Load(confmap.Provider(defaultValues, "."), nil)
	_ = k.Load(env.Provider(".", env.Opt{
		Prefix: envPrefix,
		TransformFunc: func(k, v string) (string, any) {
			k = strings.TrimPrefix(strings.ToUpper(k), envPrefix)
			k = strings.ReplaceAll(strings.ToLower(k), "__", ".")
			return k, v
		},
	}), nil)

	var conf T
	var err error

	err = k.Unmarshal("", &conf)
	if err != nil {
		return nil, fmt.Errorf("unmarshal:\n%w", err)
	}

	err = validate(&conf)
	if err != nil {
		return nil, fmt.Errorf("validate:\n%w", err)
	}

	return &conf, nil
}
