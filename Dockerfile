ARG BASE_IMAGE=newrelic/infrastructure-bundle:3.2.74

FROM $BASE_IMAGE AS base

# Set by docker automatically
# If building with `docker build`, make sure to set GOOS/GOARCH explicitly when calling make:
# `make compile GOOS=something GOARCH=something`
# Otherwise the makefile will not append them to the binary name and docker build will fail.
ARG TARGETOS
ARG TARGETARCH

# Add the nri-ecs integration binary to the default folders.
ADD --chmod=755 bin/nri-ecs-${TARGETOS}-${TARGETARCH} /var/db/newrelic-infra/newrelic-integrations/bin/
RUN mv /var/db/newrelic-infra/newrelic-integrations/bin/nri-ecs-${TARGETOS}-${TARGETARCH} \
       /var/db/newrelic-infra/newrelic-integrations/bin/nri-ecs

RUN rm /etc/newrelic-infra/integrations.d/docker-config.yml

# Activates the nri-ecs integration in the image by default.
# Some Envars needed to configure the integration are set in the deployment task
# and added to NRIA_PASSTHROUGH_ENVIRONMENT.
ADD nri-ecs-config.yml /var/db/newrelic-infra/integrations.d/nri-ecs-config.yml
ADD nri-docker-config.yml /var/db/newrelic-infra/integrations.d/nri-docker-config.yml
