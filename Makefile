INTEGRATION  := nri-ecs
BINARY_NAME   = $(INTEGRATION)

# Version of the integration without 'v' prefix. Populated from release tag on CI/CD.
RELEASE_VERSION ?= "0.0.0"
RELEASE_STRING := ${RELEASE_VERSION}

COMMIT ?= $(shell git rev-parse HEAD || echo "unknown")
LD_FLAGS ?= "-X 'main.integrationVersion=$(RELEASE_VERSION)' -X 'main.gitCommit=$(COMMIT)'"

NRI_ECS_IMAGE_REPO ?= newrelic/nri-ecs
NRI_ECS_IMAGE_TAG ?= "dev"
NRI_ECS_IMAGE := $(NRI_ECS_IMAGE_REPO):$(NRI_ECS_IMAGE_TAG)

MANIFEST_DIR    := ./package/manifests

EXAMPLE_MANIFEST_DIR = ./deploy

INSTALLER_SCRIPT_FILE = newrelic-infra-ecs-installer.sh

TASK_DEFINITION_FILE_TEMPLATE := newrelic-infra-ecs-ec2-<VERSION>.json
EXAMPLE_TASK_DEFINITION_FILE := task_definition_example.json

FARGATE_SIDECAR_FILE_TEMPLATE := newrelic-infra-ecs-fargate-example-<VERSION>.json
EXAMPLE_FARGATE_SIDECAR_FILE := fargate_sidecar_example.json

S3_BASE_FOLDER ?= s3://nr-downloads-main/infrastructure_agent
S3_ECS_FOLDER := $(S3_BASE_FOLDER)/integrations/ecs
S3_CLOUDFORMATION_FOLDER := $(S3_ECS_FOLDER)/cloudformation

all: build

build: clean test compile

upload_manifests: prepare_manifests
	@echo "=== $(INTEGRATION) === [ upload manifests ]: uploading manifests to S3"

	# newrelic-infra-ecs-ec2-$(RELEASE_STRING).json
	aws s3 cp $(MANIFEST_DIR)/$(EXAMPLE_TASK_DEFINITION_FILE) \
		${S3_ECS_FOLDER}/$(subst <VERSION>,${RELEASE_STRING},${TASK_DEFINITION_FILE_TEMPLATE})

	# newrelic-infra-ecs-ec2-latest.json
	aws s3 cp $(MANIFEST_DIR)/$(EXAMPLE_TASK_DEFINITION_FILE) \
		${S3_ECS_FOLDER}/$(subst <VERSION>,latest,${TASK_DEFINITION_FILE_TEMPLATE})

	# newrelic-infra-ecs-fargate-example-$(RELEASE_STRING).json
	aws s3 cp $(MANIFEST_DIR)/$(EXAMPLE_FARGATE_SIDECAR_FILE) \
		${S3_ECS_FOLDER}/$(subst <VERSION>,${RELEASE_STRING},${FARGATE_SIDECAR_FILE_TEMPLATE})

	# newrelic-infra-ecs-fargate-example-latest.json
	aws s3 cp $(MANIFEST_DIR)/$(EXAMPLE_FARGATE_SIDECAR_FILE) \
		${S3_ECS_FOLDER}/$(subst <VERSION>,latest,${FARGATE_SIDECAR_FILE_TEMPLATE})

	# installer shell script
	aws s3 cp $(MANIFEST_DIR)/$(INSTALLER_SCRIPT_FILE) \
		$(S3_ECS_FOLDER)/$(INSTALLER_SCRIPT_FILE)

	# all cloudformation files
	aws s3 sync $(MANIFEST_DIR)/cloudformation/ \
		$(S3_CLOUDFORMATION_FOLDER)


prepare_manifests:
	@echo "=== $(INTEGRATION) === [ prepare manifests ]: building package for releasing manifests"

	@mkdir -p $(MANIFEST_DIR)
	# copy example manifest files to manifest package directory
	cp $(EXAMPLE_MANIFEST_DIR)/$(EXAMPLE_TASK_DEFINITION_FILE) $(MANIFEST_DIR)
	cp $(EXAMPLE_MANIFEST_DIR)/$(EXAMPLE_FARGATE_SIDECAR_FILE) $(MANIFEST_DIR)
	cp $(EXAMPLE_MANIFEST_DIR)/$(INSTALLER_SCRIPT_FILE) $(MANIFEST_DIR)
	cp -r $(EXAMPLE_MANIFEST_DIR)/cloudformation $(MANIFEST_DIR)

	# change all occurences of <INTEGRATION_IMAGE> to the released image
	find $(MANIFEST_DIR) -type f -exec sed -i "s|<INTEGRATION_IMAGE>|$(NRI_ECS_IMAGE)|" {} +

clean:
	@echo "=== $(INTEGRATION) === [ clean ]: Removing binaries and temporary files..."
	@rm -rfv bin package

compile: CGO_ENABLED=0
compile:
	@echo "=== $(INTEGRATION) === [ compile ]: Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME)-$(GOOS)-$(GOARCH) -ldflags $(LD_FLAGS) ./cmd
 
compile-multiarch:
	$(MAKE) compile GOOS=linux GOARCH=amd64
	$(MAKE) compile GOOS=linux GOARCH=arm64
	$(MAKE) compile GOOS=linux GOARCH=arm

## GOOS and GOARCH are manually set so the output BINARY_NAME includes them as suffixes.
## Additionally, DOCKER_BUILDKIT is set since it's needed for Docker to populate TARGETOS and TARGETARCH ARGs.
## Here we call $(MAKE) build instead of using a dependency because the latter would, for some reason, prevent
## the BINARY_NAME conditional from working.
## Multi-Arch image building happens on the CI workflow. This target is for testing only.
image: GOOS := $(if $(GOOS),$(GOOS),linux)
image: GOARCH := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
image: ## Builds metrics-adapter Docker image.
	@if [ "$(GOOS)" != "linux" ]; then echo "'make image' must be called with GOOS=linux (or empty), found '$(GOOS)'"; exit 1; fi
	$(MAKE) compile GOOS=$(GOOS) GOARCH=$(GOARCH)
	DOCKER_BUILDKIT=1 docker build --rm=true -t $(NRI_ECS_IMAGE_REPO) .	

test:
	@echo "=== $(INTEGRATION) === [ test ]: Running unit tests..."
	@go test -race ./...

debug-mode:
	@echo "=== $(INTEGRATION) === [ debug ]: Running debug mode..."
	@-docker rm -f nri-ecs
	@docker build -t fsi/nri-ecs:debug -f Dockerfile.debug .
	@docker run -i -t -d --name nri-ecs -v /var/run/docker.sock:/var/run/docker.sock -p 80:80 fsi/nri-ecs:debug
	@docker exec -it nri-ecs sh

buildThirdPartyNotice:
	@go list -m -json all | go-licence-detector -noticeOut=NOTICE.txt -rules ./assets/licence/rules.json  -noticeTemplate ./assets/licence/THIRD_PARTY_NOTICES.md.tmpl -noticeOut THIRD_PARTY_NOTICES.md -overrides ./assets/licence/overrides -includeIndirect

# rt-update-changelog runs the release-toolkit run.sh script by piping it into bash to update the CHANGELOG.md.
# It also passes down to the script all the flags added to the make target. To check all the accepted flags,
# see: https://github.com/newrelic/release-toolkit/blob/main/contrib/ohi-release-notes/run.sh
#  e.g. `make rt-update-changelog -- -v`
rt-update-changelog:
	curl "https://raw.githubusercontent.com/newrelic/release-toolkit/v1/contrib/ohi-release-notes/run.sh" | bash -s -- $(filter-out $@,$(MAKECMDGOALS))

# configurator goals
include $(CURDIR)/configurator/terraform.mk

.PHONY: all build clean image compile compile-multiarch test buildLicenseNotice
