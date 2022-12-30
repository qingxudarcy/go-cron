# go-cron
Distributed timing task system

## Master config demo

```json
{
	"apiPort": 8070,
	"apiReadTimeout": 5000,
	"apiWriteTimeout": 5000,
	"etcdEndpoints": ["127.0.0.1:2379"],
	"etcdDialTimeout": 5000,
	"mongoUri": "mongodb://127.0.0.1:27017"
}   // defaulr config path: go-cron/config/master.json
```

## Worker config demo

```json
{
	"etcdEndpoints": ["127.0.0.1:2379"],
	"etcdDialTimeout": 5000,
	"mongoUri": "mongodb://127.0.0.1:27017",
	"logBatchSize": 100,
	"logCommitTimeout": 1000
}  // defaulr config path: go-cron/config/worker.json
```

## Run master

```shell
go run main.go master   # default
go run main.go master -c ./conifg/master.json  # specify the configuration file path
```

## Run worker

```shell
go run main.go worker   # default
go run main.go worker -c ./conifg/master.json  # specify the configuration file path
```

## Features
 - api doc
 - readme.md
 - UI
 - monitoring api for prometheus
 ...