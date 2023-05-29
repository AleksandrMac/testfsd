package config

import (
	"encoding/json"
	"fmt"

	"github.com/AleksandrMac/testfsd/internal/log"
	"github.com/AleksandrMac/testfsd/internal/metric"
	"github.com/AleksandrMac/testfsd/internal/trace"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var viperInstance = viper.New()
var Default Config

func init() {
	log.Debug("INIT CONFIG")
}

// Config struct
type Config struct {
	Metadata struct {
		ServiceName string
	}
	Server struct {
		Port uint
		Host string
	}
	Otel struct {
		Log    zap.Config
		Trace  trace.Config
		Metric metric.Config
	}
	DB       DB
	RabbitMQ RabbitMQ
}

func (d Config) String() string {
	b, _ := json.Marshal(d)
	return string(b)
}

// Parse get all config support in app
func Parse() Config {
	if err := viperInstance.Unmarshal(&Default, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc())); err != nil {
		log.Fatal(
			fmt.Sprintf("Fail to read configuration: %s", err.Error()))
	}
	return Default
}

// Viper instance
func Viper() *viper.Viper {
	return viperInstance
}
