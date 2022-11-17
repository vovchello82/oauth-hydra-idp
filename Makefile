.DEFAULT_GOAL := compile
DOCKER_REPO=vzlobins/hydra-id-provider

publish: docker docker-push manifestation

exe:
	go run main.go

compile:
	go clean
	go build -o ./hydra-id-provider main.go 

clean:
	rm -f ./hydra-id-provider
tests:
	go test -v 

docker:
	docker build --platform=linux/amd64 -t ${DOCKER_REPO}:amd64-latest . 
	docker build --platform=linux/arm64 -t ${DOCKER_REPO}:arm64-latest . 
	
docker-arm:
	docker build -t ${DOCKER_REPO}:arm64-latest .

docker-push:
	docker push ${DOCKER_REPO}:amd64-latest 
	docker push ${DOCKER_REPO}:arm64-latest 
	 
manifestation:	 
	docker manifest create ${DOCKER_REPO}:latest --amend ${DOCKER_REPO}:amd64-latest --amend ${DOCKER_REPO}:arm64-latest
	docker manifest push ${DOCKER_REPO}:latest