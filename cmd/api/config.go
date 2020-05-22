package main

import (
	"github.com/spf13/viper"
)

type configRaft struct {
	NodeId    string `mapstructure:"node_id"`
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	VolumeDir string `mapstructure:"volume_dir"`
}

type configServer struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// configLeaderServer is the host port of raft leader address
type configLeaderServer struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type config struct {
	Server       configServer       `mapstructure:"server"`
	LeaderServer configLeaderServer `mapstructure:"leader_server"`
	Raft         configRaft         `mapstructure:"raft"`
}

func readConfig() (conf config, err error) {
	conf = config{}

	var v = viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile("config.yaml")

	err = v.ReadInConfig()
	if err != nil {
		return
	}

	err = v.Unmarshal(&conf)
	if err != nil {
		return
	}

	return
}
