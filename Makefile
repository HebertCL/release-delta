################# Variables ################
VERSION := "hebertcuellar/release-reporter/release-reporter:$(shell git rev-parse --short HEAD)"
VERSION_LATEST := "hebertcuellar/release-reporter/release-reporter:latest"
############################################

tidy:
	go mod tidy

build:
	docker build \
	-t $(VERSION) \
	-t $(VERSION_LATEST) .

start-local: tidy
	go run *.go