
.PHONY: build-crd-controller
build-crd-controller:
	 docker build . -f Dockerfile.crd -t zmalikshxil/kudo-crd-controller:0.0.1-alpha

.PHONY: push-crd-controller
push-crd-controller: build-crd-controller
	docker push zmalikshxil/kudo-crd-controller:0.0.1-alpha

.PHONY: build-bridge-controller
build-bridge-controller:
	 docker build . -f Dockerfile.bridge -t zmalikshxil/kudo-bridge-controller:0.0.1-alpha

.PHONY: push-bridge-controller
push-bridge-controller: build-bridge-controller
	docker push zmalikshxil/kudo-bridge-controller:0.0.1-alpha

.PHONY: build.all
build.all: push-crd-controller build-bridge-controller

.PHONY: push.all
push.all: push-crd-controller push-bridge-controller

generate:
ifneq ($(shell go list -f '{{.Version}}' -m sigs.k8s.io/controller-tools), $(shell controller-gen --version 2>/dev/null | cut -b 10-))
	@echo "(Re-)installing controller-gen. Current version:  $(controller-gen --version 2>/dev/null | cut -b 10-). Need $(go list -f '{{.Version}}' -m sigs.k8s.io/controller-tools)"
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@$$(go list -f '{{.Version}}' -m sigs.k8s.io/controller-tools)
endif
	controller-gen crd paths=./bridge-controller/pkg/apis/... output:crd:dir=config/crds output:stdout
ifeq (, $(shell which go-bindata))
	go get github.com/go-bindata/go-bindata/go-bindata@$$(go list -f '{{.Version}}' -m github.com/go-bindata/go-bindata)
endif
	./hack/update_codegen.sh
