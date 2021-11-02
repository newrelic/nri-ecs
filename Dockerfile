#### Please note that this dockerfile is not used to release the image,
#### It is used merely to create a test image to run tests or to be uploaded manually for testing pourpuses
ARG BASE_IMAGE=newrelic/infrastructure-bundle
ARG BASE_IMAGE_TAG=2.6.4
ARG GO_VERSION=1.17

FROM golang:${GO_VERSION}-alpine AS build
RUN apk add --no-cache --update git make

WORKDIR /go/src/github.com/newrelic/nri-ecs
COPY . .
ENV CGO_ENABLED=0
RUN make compile

FROM $BASE_IMAGE:$BASE_IMAGE_TAG as ec2

ENV NRIA_IS_CONTAINERIZED=true
ENV NRIA_PASSTHROUGH_ENVIRONMENT=ECS_CONTAINER_METADATA_URI

COPY --from=build /go/src/github.com/newrelic/nri-ecs/bin/nri-ecs /var/db/newrelic-infra/newrelic-integrations/bin/
COPY --from=build /go/src/github.com/newrelic/nri-ecs/newrelic-nri-ecs-config.yml /etc/newrelic-infra/integrations.d/

# Fargate has some extras on top of EC2
FROM ec2 as fargate

ENV NRIA_PASSTHROUGH_ENVIRONMENT=ECS_CONTAINER_METADATA_URI,FARGATE
ENV NRIA_IS_SECURE_FORWARD_ONLY=true
COPY --from=build /go/src/github.com/newrelic/nri-ecs/newrelic-nri-docker-config-fargate.yml /etc/newrelic-infra/integrations.d/docker-config.yml
ENV FARGATE=true
