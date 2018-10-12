# Copyright 2018 Intel Corporation.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# --------- General build rules

.PHONY: all
all: nff-go-nat client/client

client/client: .check-env Makefile client/client.go
	cd client && go build

nff-go-nat: .check-env Makefile nat.go $(wildcard nat/*.go)
	go build

.PHONY: clean
clean:
	-rm nat
	-rm client/client

# --------- Docker images build rules

IMAGENAME=nff-go-nat
BASEIMAGE=nff-go-base
# Add user name to generated images
ifdef NFF_GO_IMAGE_PREFIX
WORKIMAGENAME=$(NFF_GO_IMAGE_PREFIX)/$(USER)/$(IMAGENAME)
IMAGE_PREFIX=$(NFF_GO_IMAGE_PREFIX)/$(USER)
BASEIMAGENAME=$(NFF_GO_IMAGE_PREFIX)/$(USER)/$(BASEIMAGE)
else
WORKIMAGENAME=$(USER)/$(IMAGENAME)
IMAGE_PREFIX=$(USER)
BASEIMAGENAME=$(USER)/$(BASEIMAGE)
endif

.PHONY: .check-base-image
.check-base-image:
	@if ! docker images '$(BASEIMAGENAME)' | grep '$(BASEIMAGENAME)' > /dev/null; then		\
		echo "!!! You need to build $(BASEIMAGENAME) docker image in $(NFF_GO) repository";	\
		exit 1;											\
	fi

.PHONY: images
images: .check-base-image Dockerfile all
	docker build --build-arg USER_NAME=$(IMAGE_PREFIX) -t $(WORKIMAGENAME) .

.PHONY: clean-images
clean-images: clean
	-docker rmi $(WORKIMAGENAME)

# --------- Docker deployment rules

.PHONY: .check-deploy-env
.check-deploy-env: .check-defined-NFF_GO_HOSTS

.PHONY: deploy
deploy: .check-deploy-env images
	$(eval TMPNAME=tmp-$(IMAGENAME).tar)
	docker save $(WORKIMAGENAME) > $(TMPNAME)
	for host in `echo $(NFF_GO_HOSTS) | tr ',' ' '`; do			\
		if ! docker -H tcp://$$host load < $(TMPNAME); then break; fi;	\
	done
	rm $(TMPNAME)

.PHONY: cleanall
cleanall: .check-deploy-env clean-images
	-for host in `echo $(NFF_GO_HOSTS) | tr ',' ' '`; do	\
		docker -H tcp://$$host rmi -f $(WORKIMAGENAME);	\
	done

# --------- Test execution rules

.PHONY: .check-test-env
.check-test-env: .check-defined-NFF_GO .check-defined-NFF_GO_HOSTS $(NFF_GO)/test/framework/main/tf

.PHONY: test-stability
test-stability: .check-test-env test/stability-nat.json
	$(NFF_GO)/test/framework/main/tf -directory nat-stabilityresults -config test/stability-nat.json -hosts $(NFF_GO_HOSTS)

.PHONY: test-stability-vlan
test-stability-vlan: .check-test-env test/stability-nat-vlan.json
	$(NFF_GO)/test/framework/main/tf -directory nat-vlan-stabilityresults -config test/stability-nat-vlan.json -hosts $(NFF_GO_HOSTS)

.PHONY: test-performance
test-performance: .check-test-env test/perf-nat.json
	$(NFF_GO)/test/framework/main/tf -directory nat-perfresults -config test/perf-nat.json -hosts $(NFF_GO_HOSTS)

.PHONY: test-performance-vlan
test-performance-vlan: .check-test-env test/perf-nat-vlan.json
	$(NFF_GO)/test/framework/main/tf -directory nat-vlan-perfresults -config test/perf-nat-vlan.json -hosts $(NFF_GO_HOSTS)

# --------- Utility rules

.PHONY: .check-env
.check-env: 					\
	.check-defined-RTE_TARGET		\
	.check-defined-RTE_SDK			\
	.check-defined-CGO_LDFLAGS_ALLOW	\
	.check-defined-CGO_CFLAGS		\
	.check-defined-CGO_LDFLAGS

.PHONY: .check-defined-%
.check-defined-%:
	@if [ -z '${${*}}' ]; then echo "!!! Variable $* is undefined" && exit 1; fi
