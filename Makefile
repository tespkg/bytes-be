
GH_CI_USER:=$(GH_CI_USER)
GH_CI_TOKEN:=$(GH_CI_TOKEN)
GL_CI_USER:=$(GL_CI_USER)
GL_CI_TOKEN:=$(GL_CI_TOKEN)

.PHONY: settoken
settoken:
ifneq ('$(GH_CI_TOKEN)','')
	@git config --global url."https://$(GH_CI_USER):$(GH_CI_TOKEN)@github.com/tespkg".insteadOf "https://github.com/tespkg"
endif
ifneq ('$(GL_CI_TOKEN)','')
	@git config --global url."https://$(GL_CI_USER):$(GL_CI_TOKEN)@gitlab.com/".insteadOf "https://gitlab.com/"
endif

.PHONY: run
run:
	@./bin/bytes-be staff -c ./contrib/server.yaml

.PHONY: swag
swag:
	./bin/swag init -g main.go

.PHONY: build-with-vendor
build-with-vendor:
	@if test -d "./vendor/"; then echo "checking vendor folder... [OK]"; else echo "checking vendor folder... [FAILED]"; echo "run go mod vendor first"; exit 2; fi;
	@mkdir -p bin
	@echo "start building bytes-be..."
	@GOPRIVATE=github.com/tespkg/*,gitlab.com/target-digital-transformation/*,tespkg.in/* go build -mod vendor -o bin/bytes-be github.com/tespkg/bytes-be
	@echo "building done"

.PHONY: build
build: settoken
	@mkdir -p bin
	@echo "start building bytes-be..."
	@go mod tidy
	@GOPRIVATE=github.com/tespkg/*,gitlab.com/target-digital-transformation/*,tespkg.in/* go build -o bin/bytes-be github.com/tespkg/bytes-be
	@echo "building done"

.PHONY: build-image
build-image: settoken
	@echo "start building docker image"
	@docker build -t local/bytes-be -f ./Dockerfile .
	@echo "building done"

.PHONY: build-image-with-vendor
build-image-with-vendor: settoken
	@if test -d "./vendor/"; then echo "checking vendor folder... [OK]"; else echo "checking vendor folder... [FAILED]"; echo "run go mod vendor first"; exit 2; fi;
	@echo "start building docker image with vendor"
	@docker build -t local/bytes-be -f ./docker/vendor.Dockerfile .
	@echo "building done"
