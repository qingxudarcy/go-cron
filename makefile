CRON_FILE=go-cron

MATER_SERVICE=master
WORKER_SERVICE=worker

all: build

build:
	@go build -o "${CRON_FILE}"

run_master:
	@go build -o "${CRON_FILE}"
	./"${CRON_FILE}" "${MATER_SERVICE}"

run_worker:
	@go build -o "${CRON_FILE}"
	./"${CRON_FILE}" "${WORKER_SERVICE}"

docker_master:
	@docker build -f docker/Dockerfile_master -t go-cron-master .

docker_worker:
	@docker build -f docker/Dockerfile_worker -t go-cron-worker .

help:
	@echo "make 编译生成二进制文件"
	@echo "make build 编译go代码生成二进制文件"
	@echo "make run_master 编译程序并启动master服务"
	@echo "make run_worker 编译程序并启动worker服务"
	@echo "make docker_master 构建master服务docker镜像"
	@echo "make docker_worker 构建worker服务docker镜像"