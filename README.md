[![Community Project header](https://github.com/newrelic/open-source-office/raw/master/examples/categories/images/Community_Project.png)](https://github.com/newrelic/open-source-office/blob/master/examples/categories/index.md#community-project)

# New Relic integration for ECS

This integration collects metrics from ECS clusters and containers.

By itself, this integration just collects metadata of the ECS cluster, the
real value comes when combined with the [nri-docker][1] integration. The
recomended approach is to run the [infrastructre-bundle][3] which includes the
infrastructure agent and both integrations.

## Requirements

- Go 1.13
- ECS agent >= 1.21.

## Installation

Create a task definition that runs the [infrastructre-bundle][3] in your ECS
cluster. In our [official New Relic docs][2] you can find information on how to
set up your infrastructure automatically with CloudFormation, or generating the
task definition via command line or manually.

## Building

To generate the integration binary execute:

```
$ make compile
```

This will generate the binary `./bin/nri-ecs`.

# Development

A debug mode is provided to aid in development. It runs a special container
that simulates the metadata endpoints of the AWS ECS agent.

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

## Releasing

1. Release new version of nri-ecs binary to the downloads page.
1. Update the version of nri-ecs in the [infrastructre-bundle repo][3].
1. Release a new version of `infrastructre-bundle`.
1. Update the image version of `infrastructre-bundle` in the the tasks
  definitions.

## Support

New Relic hosts and moderates an online forum where customers can interact with
New Relic employees as well as other customers to get help and share best
practices. Like all official New Relic open source projects, there's a related
Community topic in the New Relic Explorers Hub. You can find this project's
topic/threads here:

https://discuss.newrelic.com/t/new-relic-ecs-integration/109092

## Contributing
Full details about how to contribute to
Contributions to improve the New Relic integration for ECS are encouraged! Keep
in mind when you submit your pull request, you'll need to sign the CLA via the
click-through using CLA-Assistant. You only have to sign the CLA one time per
project.
To execute our corporate CLA, which is required if your contribution is on
behalf of a company, or if you have any questions, please drop us an email at
opensource@newrelic.com.

## License
New Relic integration for ECS is licensed under the [Apache
2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.

The New Relic integration for ECS also uses source code from third party
libraries. Full details on which libraries are used and the terms under which
they are licensed can be found in the third party notices document.

[1]: https://github.com/newrelic/nri-docker
[2]: https://docs.newrelic.com/docs/integrations/elastic-container-service-integration/installation/install-ecs-integration
[3]: https://github.com/newrelic/infrastructure-bundle/blob/master/build/versions#L26
[4]: https://golang.org/pkg/testing/
