APP_NAME=service_check


common:

	go build -o ${APP_NAME}

dev: common
	echo "dev"
	./${APP_NAME}

com: common
	echo "company test"
	scp ${APP_NAME} root@10.0.2.36:/bin/${APP_NAME}