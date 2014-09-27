
build:
	rm -f package/*
	docker build -t deis/build-publisher .
	mkdir -p package
	docker cp `docker run -d deis/build-publisher`:/tmp/publisher.tar.gz package/
	
image:	
	# https://medium.com/@kelseyhightower/optimizing-docker-images-for-static-binaries-b5696e26eb07
	cd publish && tar zxpvf ../package/publisher.tar.gz && docker build -t deis/publisher .	
	rm -f publish/publisher*

all: build image

.PHONY: all