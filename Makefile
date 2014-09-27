include ../includes.mk

BUILD_DIR := image
DOCKER_IMAGE := deis/publisher
RELEASE_IMAGE := $(DOCKER_IMAGE):$(BUILD_TAG)
REMOTE_IMAGE := $(REGISTRY)/$(RELEASE_IMAGE)

build: check-docker
  docker build -t $(RELEASE_IMAGE) .

clean: check-docker check-registry
  docker rmi $(RELEASE_IMAGE) $(REMOTE_IMAGE)

full-clean: check-docker check-registry
  docker images -q $(DOCKER_IMAGE) | xargs docker rmi -f
  docker images -q $(REGISTRY)/$(DOCKER_IMAGE) | xargs docker rmi -f

install: check-deisctl
  deisctl scale publisher=1

push: check-docker check-registry check-deisctl
  docker tag $(RELEASE_IMAGE) $(REMOTE_IMAGE)
  docker push $(REMOTE_IMAGE)
  deisctl config publisher set image=$(REMOTE_IMAGE)

release: image
  docker tag $(DOCKER_IMAGE) $(RELEASE_IMAGE)
  docker push $(RELEASE_IMAGE)

restart: stop start

run: install start

start: check-deisctl
  deisctl start publisher

stop: check-deisctl
  deisctl stop publisher

test:
  @echo no unit tests

uninstall: check-deisctl
  deisctl scale publisher=0
