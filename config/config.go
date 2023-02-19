package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func New() (*Root, error) {
	_ = pflag.String("config", "", "path to config file")
	viper.BindPFlag("config", pflag.Lookup("config"))
	pflag.Parse()
	if v := viper.GetString("config"); v != "" {
		viper.SetConfigFile(v)
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := &Root{}
	if err := viper.Unmarshal(conf, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	))); err != nil {
		return nil, err
	}

	return conf, validator.New().Struct(conf)
}
