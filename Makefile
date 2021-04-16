NATIVEOS	 := $(shell go version | awk -F '[ /]' '{print $$4}')
NATIVEARCH	 := $(shell go version | awk -F '[ /]' '{print $$5}')
INTEGRATION  := nri-ecs
BINARY_NAME   = $(INTEGRATION)
GO_PKGS      := $(shell go list ./... | grep -v "/vendor/")
RELEASE_VERSION := 1.0.0
RELEASE_TAG :=
RELEASE_STRING := ${RELEASE_VERSION}${RELEASE_TAG}

INFRA_BUNDLE_VERSION := 2.4.0
INFRA_BUNDLE_IMAGE := newrelic/infrastructure-bundle:$(INFRA_BUNDLE_VERSION)

# compile & package targets
OS := linux
ARCH := amd64

FILENAME_TARBALL := $(BINARY_NAME)_$(OS)_$(RELEASE_STRING)_$(ARCH).tar.gz
PACKAGE_DIR     := ./package/source
TARBALL_DIR     := ./package/tarball
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
S3_TARBALL_FOLDER := $(S3_BASE_FOLDER)/binaries/$(OS)/$(ARCH)

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

	# change all occurences of <INTEGRATION_IMAGE> to the released infra bundle
	find $(MANIFEST_DIR) -type f -exec sed -i "s|<INTEGRATION_IMAGE>|$(INFRA_BUNDLE_IMAGE)|" {} +

clean:
	@echo "=== $(INTEGRATION) === [ clean ]: Removing binaries and temporary files..."
	@rm -rfv bin package

compile:
	@echo "=== $(INTEGRATION) === [ compile ]: Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./cmd

compile_for:
	@echo "=== $(INTEGRATION) === [ compile ]: Building..."
	env GOOS=$(OS) GOARCH=$(ARCH) go build -o ./bin/$(INTEGRATION) ./cmd

package_for: compile_for
	@echo "=== $(INTEGRATION) === [ package ]: Packaging..."
	mkdir -p $(PACKAGE_DIR)/var/db/newrelic-infra/newrelic-integrations/bin
	mkdir -p $(PACKAGE_DIR)/var/db/newrelic-infra/integrations.d
	mkdir -p $(TARBALL_DIR)
	cp ./bin/$(INTEGRATION) $(PACKAGE_DIR)/var/db/newrelic-infra/newrelic-integrations/bin/$(INTEGRATION)
	cp newrelic-nri-ecs-config.yml $(PACKAGE_DIR)/var/db/newrelic-infra/integrations.d/nri-ecs-config.yml
	tar -czf $(TARBALL_DIR)/$(FILENAME_TARBALL) -C $(PACKAGE_DIR) ./

release_tarball_package_for: package_for
	@echo "=== $(INTEGRATION) === [ package ]: Releasing..."
	aws s3 cp $(TARBALL_DIR)/$(FILENAME_TARBALL) ${S3_TARBALL_FOLDER}/$(FILENAME_TARBALL)
	gh release upload "v$(RELEASE_VERSION)" "$(TARBALL_DIR)/$(FILENAME_TARBALL)" --repo "github.com/newrelic/nri-ecs" --clobber

test:
	@echo "=== $(INTEGRATION) === [ test ]: Running unit tests..."
	@go test -race $(GO_PKGS)

debug-mode:
	@echo "=== $(INTEGRATION) === [ debug ]: Running debug mode..."
	@-docker rm -f nri-ecs
	@docker build -t fsi/nri-ecs:debug -f Dockerfile.debug .
	@docker run -i -t -d --name nri-ecs -v /var/run/docker.sock:/var/run/docker.sock fsi/nri-ecs:debug
	@docker exec -it nri-ecs sh

buildThirdPartyNotice:
	@go list -m -json all | go-licence-detector -noticeOut=NOTICE.txt -rules ./assets/licence/rules.json  -noticeTemplate ./assets/licence/THIRD_PARTY_NOTICES.md.tmpl -noticeOut THIRD_PARTY_NOTICES.md -overrides ./assets/licence/overrides -includeIndirect

# configurator goals
include $(CURDIR)/configurator/terraform.mk

.PHONY: all build clean compile test buildLicenseNotice
