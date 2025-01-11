VERSION := "0.0.1-$(shell git rev-parse --short HEAD)"

tidy:
	go mod tidy

build:
	docker build \
	-t $(VERSION) \
	-t latest .