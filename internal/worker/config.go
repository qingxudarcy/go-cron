package worker

import (
	"encoding/json"
	"os"
)

type Config struct {
	EtcdEndpoints    []string `json:"etcdEndpoints"`
	EtcdDialTimeout  int      `json:"etcdEtcdDialTimeout"`
	MongoUri         string   `json:"mongoUri"`
	LogBatchSize     int      `json:"logBatchSize"`
	LogCommitTimeout int      `json:"logCommitTimeout"`
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
