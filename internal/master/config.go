package master

import (
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ApiPort         int      `mapstructure:"apiPort"`
	ApiReadTimeout  int      `mapstructure:"apiReadTimeout"`
	ApiWriteTimeout int      `mapstructure:"apiWriteTimeout"`
	EtcdEndpoints   []string `mapstructure:"etcdEndpoints"`
	EtcdDialTimeout int      `mapstructure:"etcdDialTimeout"`
	MongoUri        string   `mapstructure:"mongoUri"`
}

var (
	G_config *Config
)

func InitConfig(fileName string) (err error) {
	var (
		configAbsPath string
		configPath    string
		configName    string
		configType    string
		viperConfig   *viper.Viper
		conf          Config
	)

	if configAbsPath, err = filepath.Abs(fileName); err != nil {
		return
	}

	configPath = filepath.Dir(configAbsPath)
	configName = filepath.Base(configAbsPath)
	configType = filepath.Ext(configAbsPath)
	configName = strings.TrimSuffix(configName, configType)
	configType = strings.TrimPrefix(configType, ".")

	viperConfig = viper.New()
	viperConfig.SetConfigName(configName)
	viperConfig.SetConfigType(configType)
	viperConfig.AddConfigPath(configPath)

	if err = viperConfig.ReadInConfig(); err != nil {
		return
	}

	if err = viperConfig.Unmarshal(&conf); err != nil {
		return
	}

	G_config = &conf

	return
}
