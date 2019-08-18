IMAGE_NAME := gitlack
CONTAINER_NAME := gitlack
VERSION := v0.0.3


.PHONY: build
build:
	go build -a -v -o main ./cmd

.PHONY: run
run:
	./main \
		--slack-token YOUR-SLACK-TOKEN \
		--gitlab-domain YOUR-GITLAB-DOMAIN \
		--gitlab-token YOUR-GITLAB-TOKEN

.PHONY: test
test:
	go test -coverprofile=coverage.out $(shell go list ./... | grep -v /mocks)
