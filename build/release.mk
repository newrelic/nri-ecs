BUILD_DIR    := ./bin/
GORELEASER_VERSION ?= v2.4.4
GORELEASER_BIN ?= bin/goreleaser

bin:
	@mkdir -p $(BUILD_DIR)

$(GORELEASER_BIN): bin
	@echo "===> $(INTEGRATION) === [$(GORELEASER_BIN)] Installing goreleaser $(GORELEASER_VERSION)"
	@(wget -qO /tmp/goreleaser.tar.gz https://github.com/goreleaser/goreleaser/releases/download/$(GORELEASER_VERSION)/goreleaser_$(OS_DOWNLOAD)_x86_64.tar.gz)
	@(tar -xf  /tmp/goreleaser.tar.gz -C bin/)
	@(rm -f /tmp/goreleaser.tar.gz)
	@echo "===> $(INTEGRATION) === [$(GORELEASER_BIN)] goreleaser downloaded"

.PHONY : release/clean
release/clean:
	@echo "===> $(INTEGRATION) === [release/clean] remove build metadata files"
	rm -fv $(CURDIR)/cmd/versioninfo.json
	rm -fv $(CURDIR)/cmd/resource.syso

.PHONY : release/deps
release/deps: $(GORELEASER_BIN)
	@echo "===> $(INTEGRATION) === [release/deps] installing deps"
	@go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
	@go mod tidy

.PHONY : release/build-fips
release/build-fips: release/deps release/clean
ifeq ($(GENERATE_PACKAGES), true)
	@echo "===> $(INTEGRATION) === [release/build] PRERELEASE/RELEASE compiling fips binaries, creating packages, archives"
	# TAG_SUFFIX should be set as "-pre" during prereleases
	@$(GORELEASER_BIN) release --config $(CURDIR)/.goreleaser-fips.yml --skip=validate --clean
else
	@echo "===> $(INTEGRATION) === [release/build-fips] build compiling fips binaries"
	# release/build with PRERELEASE unset is actually called only from push/pr pipeline to check everything builds correctly
	@$(GORELEASER_BIN) build --config $(CURDIR)/.goreleaser-fips.yml --skip=validate --snapshot --clean
endif

.PHONY : release/fix-archive
release/fix-archive:
	@echo "===> $(INTEGRATION) === [release/fix-archive] fixing tar.gz archives internal structure"
	@bash $(CURDIR)/build/nix/fix_archives.sh $(CURDIR)

.PHONY : release-fips
release-fips: release/build-fips release/fix-archive release/clean
	@echo "===> $(INTEGRATION) === [release-fips] fips pre-release cycle complete for nix"

OS := $(shell uname -s)
ifeq ($(OS), Darwin)
	OS_DOWNLOAD := "darwin"
	TAR := gtar
else
	OS_DOWNLOAD := "linux"
endif