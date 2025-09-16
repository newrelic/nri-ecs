.PHONY : ci/pull-builder-image
ci/pull-builder-image:
	@docker pull $(BUILDER_IMAGE)

.PHONY : ci/deps
ci/deps: ci/pull-builder-image

.PHONY : ci/debug-container
ci/debug-container: ci/deps
	@docker run --rm -it \
			-v $(CURDIR):/go/src/github.com/newrelic/nri-$(INTEGRATION) \
			-w /go/src/github.com/newrelic/nri-$(INTEGRATION) \
			-e PRERELEASE=true \
			-e GITHUB_TOKEN=$(GITHUB_TOKEN) \
			-e TAG \
			-e GPG_MAIL \
			-e GPG_PASSPHRASE \
			-e GPG_PRIVATE_KEY_BASE64 \
			$(BUILDER_IMAGE) bash

.PHONY : ci/prerelease-fips
ci/prerelease-fips: ci/deps
ifdef TAG
	@docker run --rm -t \
			--name "nri-$(INTEGRATION)-prerelease" \
			-v $(CURDIR):/go/src/github.com/newrelic/nri-$(INTEGRATION) \
			-w /go/src/github.com/newrelic/nri-$(INTEGRATION) \
			-e INTEGRATION \
			-e PRERELEASE=true \
			-e GITHUB_TOKEN \
			-e REPO_FULL_NAME \
			-e TAG \
			-e TAG_SUFFIX \
			-e GENERATE_PACKAGES \
			-e PRERELEASE \
			$(BUILDER_IMAGE) make release-fips
else
	@echo "===> $(INTEGRATION) ===  [ci/prerelease] TAG env variable expected to be set"
	exit 1
endif