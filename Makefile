all: help
build_docker:
	./docker_build.sh
build:
	go build -ldflags "-s -w" -v -o taotie main.go
debug:
	go run main.go
debug1:
	go run main.go -role=proxy
debug2:
	go run main.go -role=awsCategoryTimer
debug3:
	go run main.go -role=awsAsinTimer
debug4:
	go run main.go -role=awsCategoryTask
debug5:
	go run main.go -role=awsAsinTask
install_db:
	cd doc/db && ./install.sh
install_es:
	cd doc/es && ./install.sh
install:
	cd doc/install && ./install.sh
restart:
	cd doc/install && docker-compose up -d
install_mac:
	cd doc/install && ./install_mac.sh
clean:
	docker exec -it myredis redis-cli -a hunterhug
help:
	@echo "使用说明：\n\
	make build：编译裸机二进制\n\
	make build_docker：构建容器镜像\n\
	make debug：调试运行程序\n\
	make install_db：Linux下部署数据库环境\n\
	make install_es：Linux下部署Es\n\
	make install：Linux下部署服务(需要手动编辑配置doc/config.yaml)\n\
	make install_mac：Mac下部署服务(需要手动编辑配置doc/config_mac.yaml)\
	"
.PHONY: all build restart debug debug1 debug2 debug3 debug4 debug5 clean install install_db install_es install_mac build_docker help
