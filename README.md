# Flyte

![](https://github.com/lyft/flyte/workflows/tests/badge.svg)

Flyte is a K8s-native extensible orchestration engine that manages the core machine learning pipelines at Lyft: etas, pricing, incentives, mapping, vision, and more. 

Home:  https://flyte.org

Docs:  https://lyft.github.io/flyte

# Repos 

Repo | Lang | Purpose
--- | --- | ---
[flyte](https://github.com/lyft/flyte) | RST | home, documentation, issues
[flyteidl](https://github.com/lyft/flyteidl) | IDL | interface definitions
[flytepropeller](https://github.com/lyft/flytepropeller) | Golang | execution engine
[flyteadmin](https://github.com/lyft/flyteadmin) | Golang | control plane
[flytekit](https://github.com/lyft/flytekit) | Python | python SDK and tools
[flyteconsole](https://github.com/lyft/flyteconsole) | Typescript | admin console
[flyteplugins](https://github.com/lyft/flyteplugins) | Golang | Flyte plugins
[flytestdlib](https://github.com/lyft/flytestdlib) |  Golang | standard library
[flytesnacks](https://github.com/lyft/flytesnacks) | Python | examples, tips, and tricks

# K8s Operators tested in production in Flyte Pipelines

[Spark](https://github.com/GoogleCloudPlatform/spark-on-k8s-operator) | Go | Apache Spark 
[Flink](https://github.com/lyft/flinkk8soperator) | Go | Apache Flink streaming
