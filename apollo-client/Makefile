VERSION=1.0.0
ifndef NAME
    NAME=apollo-client
endif

dev:
	fresh
run:
	go run main.go
run-bg:
	nohup ./$(NAME)  2>&1 > $(NAME).nohup &
run-grpc:
	go run main.go
run-grpc-bg:
	nohup ./$(NAME)  2>&1 > $(NAME).nohup &
build:
	GO111MODULE=on CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -mod=vendor -o $(NAME)
build-darwin:
	GO111MODULE=on CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -mod=vendor -o $(NAME)
clean:
	rm -rf ./$(NAME) ./$(NAME).nohup
kill:
	killall $(NAME)
