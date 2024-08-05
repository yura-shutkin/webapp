.PHONY: help
help: ## 				Show this help
	@sed -e '/__hidethis__/d; /^\.PHONY.*/d; /[A-Z0-9#]?*/!d; /^\t/d; s/:.##/\t/g; s/^####.*//; s/#/-/g; s/^\([A-Z0-9_]*=.*\)/| \1/g; s/^\([a-zA-Z0-9]\)/* \1/g; s/^| \(.*\)/\1/' $(MAKEFILE_LIST)

################################################################################
### WebApp ######################################################################
################################################################################
REGISTRY_ADDR="localhost:5000"
DISTRO=ubuntu
GO_VER="1.22.5"
VERSION=$(shell cat version)
####

.PHONY: build
build: ##			Build image for different platforms
	@export DOCKER_CLI_EXPERIMENTAL=enabled && \
		docker buildx create --use --name=crossplat --node=crossplat && \
		docker buildx build --rm \
			--platform linux/amd64,linux/arm64 \
			--output "type=image,push=false" \
			--build-arg=GOLANG_VER=${GO_VER} \
			-t "${REGISTRY_ADDR}/web-app:${VERSION}-${DISTRO}" \
			-f docker/Dockerfile-${DISTRO} .

.PHONY: publish
publish: ##				Build and push image for different platforms
#	@docker push "${REGISTRY_ADDR}/web-app:${VERSION}-${DISTRO}"
	@export DOCKER_CLI_EXPERIMENTAL=enabled && \
    	docker buildx create --use --name=crossplat --node=crossplat && \
    	docker buildx build --rm \
    		--platform linux/amd64,linux/arm64 \
    		--output "type=image,push=true" \
    		--build-arg=GOLANG_VER=${GO_VER} \
    		-t "${REGISTRY_ADDR}/web-app:${VERSION}-${DISTRO}" \
    		-f docker/Dockerfile-${DISTRO} .

.PHONY: build-n-push
build-n-push: ##			Build app image and push into registry
	@$(MAKE) build
	@$(MAKE) push
