export tag=v1.0.0
export image_prefix=andywuwu/httpserver
export image=${image_prefix}:${tag}

.PHONY: build docker docker_push

build:
	echo "building httpserver binary"
	go build -o bin/linux/httpServer .

docker:
	docker build -t "${image}" .

docker_push: docker
	docker push "${image}"
