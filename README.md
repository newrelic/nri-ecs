<a href="https://opensource.newrelic.com/oss-category/#community-plus"><picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/dark/Community_Plus.png"><source media="(prefers-color-scheme: light)" srcset="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/Community_Plus.png"><img alt="New Relic Open Source community plus project banner." src="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/Community_Plus.png"></picture></a>

# New Relic integration for Amazon ECS

This integration collects metrics from ECS clusters and containers in AWS.

By itself, this integration just collects metadata of the ECS cluster, that is used to decorate all the metrics collected by the [nri-docker][1] integration , the Infrastructure Agent and the on host integrations that has been activated.

This repo generates the [newrelic/nri-ecs][5] image which is based on the [infrastructure-bundle][3] that contains the Agent and the others on host integrations.

## Table of contents

- [Table of contents](#table-of-contents)
- [Requirements](#requirements)
- [Installation](#installation)
- [Building](#building)
- [Support](#support)
- [Contributing](#contribute)
- [License](#license)

## Requirements

- Go 1.19
- ECS agent version 1.21 or greater.

## Installation

Create a task definition that runs [newrelic/nri-ecs][5] in your ECS cluster. In our [docs][2] you can find information on how to
set up your infrastructure automatically with CloudFormation, or generating the task definition via command line or manually.

## Building

To generate the integration image execute:

```
$ make image NRI_ECS_IMAGE_REPO=myrepo/nri-ecs
```

This will generate the integration docker image for Linux amd64.

# Development

A debug mode is provided to aid in development. It runs a special container that simulates the metadata endpoints of the AWS ECS agent.

To build this container and get a shell into, run `make debug-mode`.

# Testing

To execute unit tests, run this command:

```
$ make test
```

You can run a specific test by invoking go (which is also how you can run tests on Windows):

```
$ go test -race -run ''      # Run all tests.
$ go test -race -run Foo     # Run top-level tests matching "Foo", such as "TestFooBar".
$ go test -race -run Foo/A=  # For top-level tests matching "Foo", run subtests matching "A=".
$ go test -race -run /A=1    # For all top-level tests, run subtests matching "A=1".
```

For more information, see [Testing][4] in the official Go docs.

## Support

Should you need assistance with New Relic products, you are in good hands with several support diagnostic tools and support channels.



If the issue has been confirmed as a bug or is a feature request, file a GitHub issue.

**Support Channels**

- [New Relic Documentation](https://docs.newrelic.com): Comprehensive guidance for using our platform
- [New Relic Community](https://discuss.newrelic.com): The best place to engage in troubleshooting questions
- [New Relic Developer](https://developer.newrelic.com/): Resources for building a custom observability applications
- [New Relic University](https://learn.newrelic.com/): A range of online training for New Relic users of every level
- [New Relic Technical Support](https://support.newrelic.com/) 24/7/365 ticketed support. Read more about our [Technical Support Offerings](https://docs.newrelic.com/docs/licenses/license-information/general-usage-licenses/support-plan).

## Privacy

At New Relic we take your privacy and the security of your information seriously, and are committed to protecting your information. We must emphasize the importance of not sharing personal data in public forums, and ask all users to scrub logs and diagnostic information for sensitive information, whether personal, proprietary, or otherwise.

We define “Personal Data” as any information relating to an identified or identifiable individual, including, for example, your name, phone number, post code or zip code, Device ID, IP address, and email address.

For more information, review [New Relic’s General Data Privacy Notice](https://newrelic.com/termsandconditions/privacy).

## Contribute

We encourage your contributions to improve this project! Keep in mind that when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.

If you have any questions, or to execute our corporate CLA (which is required if your contribution is on behalf of a company), drop us an email at opensource@newrelic.com.

**A note about vulnerabilities**

As noted in our [security policy](../../security/policy), New Relic is committed to the privacy and security of our customers and their data. We believe that providing coordinated disclosure by security researchers and engaging with the security community are important means to achieve our security goals.

If you believe you have found a security vulnerability in this project or any of New Relic's products or websites, we welcome and greatly appreciate you reporting it to New Relic through [HackerOne](https://hackerone.com/newrelic).

If you would like to contribute to this project, review [these guidelines](./CONTRIBUTING.md).

To all contributors, we thank you!  Without your contribution, this project would not be what it is today.

## License

nri-ecs is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.

The New Relic integration for ECS also uses source code from third party libraries. Full details on which libraries are used and the terms under which they are licensed can be found in the third party notices document.

[1]: https://github.com/newrelic/nri-docker
[2]: https://docs.newrelic.com/docs/integrations/elastic-container-service-integration/installation/install-ecs-integration
[3]: https://github.com/newrelic/infrastructure-bundle/blob/master/build/versions#L26
[4]: https://golang.org/pkg/testing/
[5]: https://hub.docker.com/r/newrelic/nri-ecs
