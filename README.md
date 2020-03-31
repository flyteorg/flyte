# Flyte
[![Current Release](https://img.shields.io/github/release/lyft/flyte.svg)](https://github.com/lyft/flyte/releases/latest)
[![Build Status](https://travis-ci.org/lyft/flyte.svg?branch=master)](https://travis-ci.org/lyft/flyte)
[![GoDoc](https://godoc.org/github.com/lyft/flyte?status.svg)](https://godoc.org/github.com/lyft/flyte)
[![License](https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
[![CodeCoverage](https://img.shields.io/codecov/c/github/lyft/flyte.svg)](https://codecov.io/gh/lyft/flyte)
[![Go Report Card](https://goreportcard.com/badge/github.com/lyft/flyte)](https://goreportcard.com/report/github.com/lyft/flyte)
![Commit activity](https://img.shields.io/github/commit-activity/w/lyft/flyte.svg?style=plastic)
![Commit since last release](https://img.shields.io/github/commits-since/lyft/flyte/latest.svg?style=plastic)

Flyte is an open source, K8s-native extensible orchestration engine that manages the core machine learning pipelines at Lyft: ETAs, pricing, incentives, mapping, vision, and more. 

# Community
Home:  https://flyte.org

Docs:  https://lyft.github.io/flyte

Slack:  [https://flyte-org.slack.com](https://docs.google.com/forms/d/e/1FAIpQLSf8bNuyhy7rkm77cOXPHIzCm3ApfL7Tdo7NUs6Ej2NOGQ1PYw/viewform)

Twitter: https://twitter.com/flyteorg

# Repos 

Repo | Language | Purpose
--- | --- | ---
[flyte](https://github.com/lyft/flyte) | RST | home, documentation, issues
[flyteidl](https://github.com/lyft/flyteidl) | Protobuf | interface definitions
[flytepropeller](https://github.com/lyft/flytepropeller) | Go | execution engine
[flyteadmin](https://github.com/lyft/flyteadmin) | Go | control plane
[flytekit](https://github.com/lyft/flytekit) | Python | python SDK and tools
[flyteconsole](https://github.com/lyft/flyteconsole) | Typescript | admin console
[datacatalog](https://github.com/lyft/datacatalog) | Go  | manage input & output artifacts
[flyteplugins](https://github.com/lyft/flyteplugins) | Go  | flyte plugins
[flytestdlib](https://github.com/lyft/flytestdlib) |  Go | standard library
[flytesnacks](https://github.com/lyft/flytesnacks) | Python | examples, tips, and tricks

# Production K8s Operators

Repo | Language | Purpose
--- | --- | ---
[Spark](https://github.com/GoogleCloudPlatform/spark-on-k8s-operator) | Go | Apache Spark batch
[Flink](https://github.com/lyft/flinkk8soperator) | Go | Apache Flink streaming
