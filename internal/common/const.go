package common

var (
	JobKeyPrefix    string = "/cron/job/"
	JobKillerPrefix string = "/cron/killer/"
	JobLockDir      string = "/cron/lock/"
	JobWorkerDir    string = "/cron/workers/"
)

var (
	JobDeleteEvent int = 0
	JobPutEvent    int = 1
	JobKillEvent   int = 2
)
