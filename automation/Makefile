
VERSION=1.0.0
REGISTRY=hub.test.com
NAME=automation
GRCPNAME=go.micro.srv.ops.automation


proto:
	protoc --proto_path=. --micro_out=. --go_out=. proto/domain/domain.proto

build:

	GOARCH=amd64 GOOS=linux go build -o $(NAME)

build-image:
	GOARCH=amd64 GOOS=linux go build -o $(NAME)
	docker build -f Dockerfile -t $(REGISTRY)/micro/$(NAME):$(VERSION) .
	docker push $(REGISTRY)/micro/$(NAME):$(VERSION)
	rm -rf $(NAME)
