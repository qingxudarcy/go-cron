package master

import (
	"encoding/json"
	"os"
)

type Config struct {
	ApiPort         int      `json:"apiPort"`
	ApiReadTimeout  int      `json:"apiReadTimeout"`
	ApiWriteTimeout int      `json:"apiWriteTimeout"`
	EtcdEndpoints   []string `json:"etcdEndpoints"`
	EtcdDialTimeout int      `json:"etcdEtcdDialTimeout"`
	MongoUri        string   `json:"mongoUri"`
}

var (
	G_config *Config
)

func InitConfig(fileName string) (err error) {
	var (
		content []byte
		conf    Config
	)
	if content, err = os.ReadFile(fileName); err != nil {
		return
	}

	if err = json.Unmarshal(content, &conf); err != nil {
		return
	}

	G_config = &conf

	return

}
