VERSION=1.0.0
ifndef NAME
    NAME=myip_to_dns
endif
PORT=0
GRCPNAME=go.micro.srv.myip_to_dns
dev:
	fresh
run:
	go run main.go --server_address=${BIND_IP}:$(PORT)
run-bg:
	nohup ./$(NAME) --server_address=${BIND_IP}:$(PORT) 2>&1 > $(NAME).nohup &
run-grpc:
	go run main.go --server_name=$(GRCPNAME) --server_version=$(VERSION) --client=grpc --server=grpc --transport=grpc --server_address=${BIND_IP}:$(PORT)
run-grpc-bg:
	nohup ./$(NAME) --server_name=$(GRCPNAME) --server_version=$(VERSION) --client=grpc --server=grpc --transport=grpc --server_address=${BIND_IP}:$(PORT) 2>&1 > $(NAME).nohup &
build:
	GO111MODULE=on CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -mod=vendor -o $(NAME)
build-darwin:
	GO111MODULE=on CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -mod=vendor -o $(NAME)
clean:
	rm -rf ./$(NAME) ./$(NAME).nohup
kill:
	killall $(NAME)
