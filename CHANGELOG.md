# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

Unreleased section should follow
[Release Toolkit](https://github.com/newrelic/release-toolkit#render-markdown-and-update-markdown).

## Unreleased

## v1.10.1 - 2023-05-18

### ⛓️ Dependencies
- Updated newrelic/infrastructure-bundle to v3.2.0 - [Changelog 🔗](https://github.com/newrelic/infrastructure-bundle/releases/tag/3.2.0)

## v1.10.0 - 2023-05-16

### 🚀 Enhancements
- Add the automatic release of the integration

### ⛓️ Dependencies
- Updated amazon/amazon-ecs-local-container-endpoints to v1.4.2
- Updated golang to v1.20
- Updated newrelic/infrastructure-bundle to v3.1.8 - [Changelog 🔗](https://github.com/newrelic/infrastructure-bundle/releases/tag/3.1.8)

## 1.9.7
- Update to infrastructure-bundle 2.8.36 and dependencies

## 1.9.6
- Update to infrastructure-bundle 2.8.31 and dependencies

## 1.9.5
- Update to infrastructure-bundle 2.8.28 and dependencies

## 1.9.3
- Update to infrastructure-bundle 2.8.26, dependencies and go version
- Fixed the installation script removing --launch-type for the update

## 1.9.2
- Update to infrastructure-bundle 2.8.25

## 1.9.1
- Update to infrastructure-bundle 2.8.24

## 1.9.0
- Update go version to 1.18
- Update to infrastructure-bundle 2.8.23

## 1.8.1
- Update to infrastructure-bundle 2.8.21
- bump github.com/stretchr/testify from 1.7.1 to 1.8.0
- bump github.com/newrelic/infra-integrations-sdk from 3.7.2+incompatible to 3.7.3+incompatible

## 1.8.0
- Update to infrastructure-bundle 2.8.9
- bump github.com/stretchr/testify from 1.7.0 to 1.7.1
- bump github.com/newrelic/infra-integrations-sdk from 3.7.1+incompatible to 3.7.2+incompatible
- removes container service role from policy on deployment scripts

## 1.7.0
- Update to infrastructure-bundle 2.8.7

## 1.6.0
- Update to infrastructure-bundle 2.8.2

## 1.5.0
- Update to infrastructure-bundle 2.8.1

## 1.4.1
- Add runtime platform on Fargate task template

## 1.4.0
- Now the integration is added to the `nri-ecs` image, which is based
  on the `infrastructure-bundle` image
- Add support for External launch type instances.
- Update to infrastructure-bundle 2.7.4

## 1.3.1 - 2021-04-16
- Update to infrastructure-bundle 2.4.1

## 1.3.0 - 2021-03-31
- Update to infrastructure-bundle 2.2.3
- Update release pipeline to publish arm64 and arm binaries as well

## 1.2.0 - 2020-11-26
- Update to infrastructure-bundle 1.6.0

## 1.1.0 - 2020-08-17
- Update to infrastructure-bundle 1.5.0

## 1.0.1 - 2020-07-17
- Fix an issue that made the integration generate incorrect cluster ARNs from
  the task definition.

## 1.0.0 - 2020-06-10
### Changed
- Product GA
