IMAGE_NAME := gitlack
CONTAINER_NAME := gitlack
VERSION := v0.0.1


.PHONY: build
build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

.PHONY: run
run:
	docker run -d --name ${CONTAINER_NAME} \
		-p 5000:5000 \
		$(IMAGE_NAME):$(VERSION) \
		--slack-token YOUR-SLACK-TOKEN \
		--gitlab-domain YOUR-GITLAB-DOMAIN \
		--gitlab-token YOUR-GITLAB-TOKEN

.PHONY: test
test:
	go test -coverprofile=coverage.out $(go list ./... | grep -v /mocks)
