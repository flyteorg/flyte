<html>
<p align="center"> 
  <img src="rsts/images/flyte-and-lf.png" alt="Flyte and LF AI & Data Logo" width="300">
</p>

<h1 align="center">
  Flyte
</h1>

<p align="center">
  Flyte is a <b>workflow automation</b> platform for <b>complex</b>, <b>mission-critical data</b> and <b>ML processes</b> at scale 
</p>

<p align="center">
  <a href="https://github.com/flyteorg/flyte/releases/latest">
    <img src="https://img.shields.io/github/release/flyteorg/flyte.svg" alt="Current Release" />
  </a>
  <a href="https://github.com/flyteorg/flyte/actions/workflows/sandbox.yml">
    <img src="https://github.com/flyteorg/flyte/actions/workflows/sandbox.yml/badge.svg" alt="Sandbox Build" />
  </a>
  <a href="https://github.com/flyteorg/flyte/actions/workflows/tests.yml">
    <img src="https://github.com/flyteorg/flyte/actions/workflows/tests.yml/badge.svg" alt="End-to-End Tests" />
  </a>
  <a href="http://www.apache.org/licenses/LICENSE-2.0.html">
    <img src="https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg" alt="License" />
  </a>
  <img src="https://img.shields.io/github/commit-activity/w/flyteorg/flyte.svg?style=plastic" alt="Commit Activity" />
  <img src="https://img.shields.io/github/commits-since/flyteorg/flyte/latest.svg?style=plastic" alt="Commits since Last Release" />
  <img src="https://img.shields.io/github/milestones/closed/flyteorg/flyte?style=plastic" alt="GitHub Milestones Completed" />
  <img src="https://img.shields.io/github/milestones/progress-percent/flyteorg/flyte/18?style=plastic" alt="GitHub Next Milestone Percentage" />
  <a href="https://flyte.rtfd.io">
    <img src="https://readthedocs.org/projects/flyte/badge/?version=latest&style=plastic" alt="Docs" />
  </a>
  <a href="https://bestpractices.coreinfrastructure.org/projects/4670"><img src="https://bestpractices.coreinfrastructure.org/projects/4670/badge"></a> 
  <img src="https://img.shields.io/twitter/follow/flyteorg?label=Follow&style=social" alt="Twitter Follow" />
  <a href="https://artifacthub.io/packages/search?repo=flyte">
    <img src="https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/flyte" alt="Flyte Helm Chart" />
  </a> 
  <a href="https://slack.flyte.org">
    <img src="https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social" alt="Join Flyte Slack" />
  </a>
</p>

<h3 align="center">
  <a href="https://flyte.org">Home Page</a>
  <span> · </span>
  <a href="#quickstart">Quick Start</a>
  <span> · </span>
  <a href="https://docs.flyte.org/">Documentation</a>
  <span> · </span>
  <a href="#features">Features</a>
  <span> · </span>
  <a href="#community--resources">Community & Resources</a>
  <span> · </span>
  <a href="CHANGELOG/">Changelogs</a>
  <span> · </span>  
  <a href="#component-repos">Components</a>
</h3>

</html>

## 💥 Introduction

Flyte is a structured programming and distributed processing platform that enables highly concurrent, scalable and maintainable workflows for `Machine Learning` and `Data Processing`. It is a fabric that connects disparate computation backends using a type safe data dependency graph. It records all changes to a pipeline, making it possible to rewind time. It also stores
a history of all executions and provides an intuitive UI, CLI and REST/gRPC API to interact with the computation.

Flyte is more than a workflow engine -- it uses a `workflow` as a core concept and a `task` (a single unit of execution) as a top level concept. Multiple tasks arranged in a data
producer-consumer order create a workflow.

`Workflows` and `Tasks` can be written in any language, with out of the box support for [Python](https://github.com/flyteorg/flytekit), [Java and Scala](https://github.com/spotify/flytekit-java).


## ⏳ Five Reasons to Use Flyte
- Kubernetes-Native Workflow Automation Platform
- Ergonomic SDK's in Python, Java & Scala
- Versioned & Auditable
- Reproducible Pipelines
- Strong Data Typing

<html>
<h2 id="quickstart"> 
  🚀 Quick Start
</h2>
</html>

With [Docker installed](https://docs.docker.com/get-docker/) and [Flytectl installed](https://docs.flyte.org/projects/flytectl/en/stable/index.html), run the following command:

```bash
  flytectl sandbox start
```

This creates a local Flyte sandbox. Once the sandbox is ready, you should see the following message: `Flyte is ready! Flyte UI is available at http://localhost:30081/console`.

Visit http://localhost:30081/console to view the Flyte dashboard.

Here's a quick visual tour of the console.

![Flyte console Example](https://github.com/flyteorg/flyte/raw/static-resources/img/first-run-console-2.gif)

To dig deeper into Flyte, refer to the [Documentation](https://docs.flyte.org/en/latest/index.html).

## ⭐️ Current Deployments & Contributors

- [Freenome](https://www.freenome.com/)
- [Gojek](https://www.gojek.io/)
- [Intel](https://www.intel.com/)
- [Lyft Rideshare, Mapping](https://www.lyft.com/)
- [Level 5 Global Autonomous (Woven Planet)](https://level-5.global/)
- [RunX.dev](https://runx.dev/)
- [Spotify](https://www.spotify.com/)
- [Striveworks](https://striveworks.us/)
- [Union.ai](https://union.ai/)
- [USU Group](https://www.usu.com/)
- [Wolt](https://www.wolt.com)

<html>
<h2 id="features"> 
  🔥 Features
</h2>
</html>

- Used at _Scale_ in production by **500+** users at Lyft with more than **1 million** executions and **40+ million** container executions per month
- A data aware platform
- Enables **collaboration across your organization** by:
  - Executing distributed data pipelines/workflows
  - Reusing tasks across projects, users, and workflows
  - Making it easy to stitch together workflows from different teams and domain experts
  - Backtracing to a specified workflow
  - Comparing results of training workflows over time and across pipelines
  - Sharing workflows and tasks across your teams
  - Simplifying the complexity of multi-step, multi-owner workflows
- **[Quick registration](https://docs.flyte.org/en/latest/getting_started.html)** -- start locally and scale to the cloud instantly
- **Centralized Inventory** constituting Tasks, Workflows and Executions
- **gRPC / REST** interface to define and execute tasks and workflows
- **Type safe** construction of pipelines -- each task has an interface which is characterized by its input and output, so illegal construction of pipelines fails during declaration rather than at runtime
- Supports multiple **[data types](https://docs.flyte.org/projects/cookbook/en/latest/auto/type_system/index.html)** for machine learning and data processing pipelines, such as Blobs (images, arbitrary files), Directories, Schema (columnar structured data), collections, maps, etc.
- Memoization and Lineage tracking
- Provides logging and observability
- Workflow features:
  - Start with one task, convert to a pipeline, attach **[multiple schedules](https://docs.flyte.org/projects/cookbook/en/latest/auto/deployment/workflow/lp_schedules.html)**, trigger using a programmatic API, or on-demand
  - Parallel step execution
  - Extensible backend to add **[customized plugin](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/extend_flyte/custom_task_plugin.html)** experience (with simplified user experience)
  - **[Branching](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/control_flow/run_conditions.html)**
  - Inline **[subworkflows](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/control_flow/subworkflows.html)** (a workflow can be embeded within one node of the top level workflow)
  - Distributed **remote child workflows** (a remote workflow can be triggered and statically verified at compile time)
  - **[Array Tasks](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/control_flow/map_task.html)** (map a function over a large dataset -- ensures controlled execution of thousands of containers)
  - **[Dynamic workflow](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/control_flow/dynamics.html)** creation and execution with runtime type safety
  - Container side [plugins](https://docs.flyte.org/projects/cookbook/en/latest/plugins.html) with first class support in Python
  - _PreAlpha_: Arbitrary flytekit-less containers supported ([RawContainer](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/containerization/raw_container.html))
- Guaranteed **[reproducibility](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/flyte_basics/task_cache.html)** of pipelines via:
  - Versioned data, code and models
  - Automatically tracked executions
  - Declarative pipelines
- **Multi cloud support** (AWS, GCP and others)
- Extensible core, modularized, and deep observability
- No single point of failure and is resilient by design
- Automated notifications to Slack, Email, and Pagerduty
- [Multi K8s cluster support](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/kubernetes/pod/index.html)
- Out of the box support to run **[Spark jobs on K8s](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/kubernetes/k8s_spark/index.html)**, **[Hive queries](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/external_services/hive/index.html)**, etc.
- Snappy Console
- Python CLI and Golang CLI (flytectl)
- Written in **Golang** and optimized for large running jobs' performance
- [Grafana templates](https://grafana.com/orgs/flyte) (user/system observability)

### In Progress

- Demos; Distributed Pytorch, feature engineering, etc.
- Integrations; Great Expectations, Feast
- Least-privilege Minimal Helm Chart
- Relaunch execution in recover mode
- Documentation as code

## 🔌 Available Plugins

- Containers
- [K8s Pods](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/kubernetes/pod/index.html)
- AWS Batch Arrays
- K8s Pod Arrays
- K8s Spark (native [Pyspark](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/kubernetes/k8s_spark/index.html) and Java/Scala)
- AWS Athena
- [Qubole Hive](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/external_services/hive/index.html)
- Presto Queries
- Distributed Pytorch (K8s Native) -- [Pytorch Operator](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/kubernetes/kfpytorch/index.html)
- Sagemaker ([builtin algorithms](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/aws/sagemaker_training/sagemaker_builtin_algo_training.html) & [custom models](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/aws/sagemaker_training/sagemaker_custom_training.html))
- Distributed Tensorflow (K8s Native) -- [TFOperator](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/kubernetes/kftensorflow/index.html)
- Papermill notebook execution ([Python](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/flytekit_plugins/papermilltasks/index.html) and Spark)
- [Type safe and data checking for Pandas dataframe](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/flytekit_plugins/pandera/index.html) using Pandera
- Versioned datastores using DoltHub and Dolt
- Use SQLAlchemy to query any relational database 
- Build your own plugins that use library containers


<html>
<h2 id="component-repos"> 
  📦 Component Repos 
</h2>
</html>

| Repo                                                         | Language      | Purpose                                        | Status           |
| ------------------------------------------------------------ | ------------- | ---------------------------------------------- | ---------------- |
| [flyte](https://github.com/flyteorg/flyte)                   | Kustomize,RST | deployment, documentation, issues              | Production-grade |
| [flyteidl](https://github.com/flyteorg/flyteidl)             | Protobuf      | interface definitions                          | Production-grade |
| [flytepropeller](https://github.com/flyteorg/flytepropeller) | Go            | execution engine                               | Production-grade |
| [flyteadmin](https://github.com/flyteorg/flyteadmin)         | Go            | control plane                                  | Production-grade |
| [flytekit](https://github.com/flyteorg/flytekit)             | Python        | python SDK and tools                           | Production-grade |
| [flyteconsole](https://github.com/flyteorg/flyteconsole)     | Typescript    | admin console                                  | Production-grade |
| [datacatalog](https://github.com/flyteorg/datacatalog)       | Go            | manage input & output artifacts                | Production-grade |
| [flyteplugins](https://github.com/flyteorg/flyteplugins)     | Go            | flyte plugins                                  | Production-grade |
| [flytestdlib](https://github.com/flyteorg/flytestdlib)       | Go            | standard library                               | Production-grade |
| [flytesnacks](https://github.com/flyteorg/flytesnacks)       | Python        | examples, tips, and tricks                     | Incubating       |
| [flytekit-java](https://github.com/flyteorg/flytekit-java)   | Java/Scala    | Java & scala SDK for authoring Flyte workflows | Incubating       |
| [flytectl](https://github.com/flyteorg/flytectl)             | Go            | A standalone Flyte CLI                         | Incomplete       |

## 🔩 Production K8s Operators

| Repo                                                                  | Language | Purpose                |
| --------------------------------------------------------------------- | -------- | ---------------------- |
| [Spark](https://github.com/GoogleCloudPlatform/spark-on-k8s-operator) | Go       | Apache Spark batch     |
| [Flink](https://github.com/lyft/flinkk8soperator)                 | Go       | Apache Flink streaming |

<html>
<h2 id="community--resources"> 
  🤝 Community & Resources
</h2>
</html>

Here are some resources to help you learn more about Flyte.

### Communication Channels

- [Slack](https://slack.flyte.org)
- [Email list](https://groups.google.com/u/0/a/flyte.org/g/users)
- [Twitter](https://twitter.com/flyteorg) 
- [LinkedIn Discussion Group](https://www.linkedin.com/groups/13962256/)
- [GitHub Discussions](https://github.com/flyteorg/flyte/discussions)

### Biweekly Community Sync

- 📣 **Flyte OSS Community Sync** Every other Tuesday, 9am-10am PDT. Checkout the [calendar](https://www.addevent.com/calendar/kE355955) and register to stay up-to-date with our meeting times. Or simply join us on [Zoom](https://us04web.zoom.us/j/71298741279?pwd=TDR1RUppQmxGaDRFdzBOa2lHN1dsZz09).
- Upcoming meeting agenda, previous meeting notes and a backlog of topics are captured in this [document](https://docs.google.com/document/d/1Jb6eOPOzvTaHjtPEVy7OR2O5qK1MhEs3vv56DX2dacM/edit#heading=h.c5ha25xc546e).
- If you'd like to revisit any previous community sync meetings, you can access the video recordings on [Flyte's YouTube channel](https://www.youtube.com/channel/UCNduEoLOToNo3nFVly-vUTQ).

### Conference Talks

- Kubecon 2019 - Flyte: Cloud Native Machine Learning and Data Processing Platform [video](https://www.youtube.com/watch?v=KdUJGSP1h9U) | [deck](https://kccncna19.sched.com/event/UaYY/flyte-cloud-native-machine-learning-data-processing-platform-ketan-umare-haytham-abuelfutuh-lyft)
- Kubecon 2019 - Running LargeScale Stateful workloads on Kubernetes at Lyft [video](https://www.youtube.com/watch?v=ECeVQoble0g)
- re:invent 2019 - Implementing ML workflows with Kubernetes and Amazon Sagemaker [video](https://youtu.be/G-wzIQQJKaE)
- Cloud-native machine learning at Lyft with AWS Batch and Amazon EKS [video](https://youtu.be/n_rRb8u1GSM)
- OSS + ELC NA 2020 [splash](https://ossna2020.sched.com/event/313cec91aa38a430a25f9571039874b8)
- Datacouncil [video](https://www.youtube.com/watch?v=1BjXC5TZAiI) | [splash](https://docs.google.com/document/d/1ZsCDOZ5ZJBPWzCNc45FhNtYQOxYHz0PAu9lrtDVnUpw/edit)
- FB AI@Scale [Making MLOps & DataOps a reality](https://www.facebook.com/atscaleevents/videos/ai-scale-flyte-making-mlops-and-dataops-a-reality/1047312585732459/)
- [GAIC 2020](http://www.globalbigdataconference.com/seattle/global-artificial-intelligence-virtual-conference-122/speaker-details/ketan-umare-113746.html)
- [OSPOCon 2021](https://events.linuxfoundation.org/ospocon/) Catch a variety of Flyte talks - final schedule and topics to be released soon.

### Blog Posts

- [Flyte blog site](https://blog.flyte.org/)

### Podcasts

- TWIML&AI - [Scalable and Maintainable ML Workflows at Lyft - Flyte](https://twimlai.com/twiml-talk-343-scalable-and-maintainable-workflows-at-lyft-with-flyte-w-haytham-abuelfutuh-and-ketan-umare/)
- Software Engineering Daily - [Flyte: Lyft Data Processing Platform](https://softwareengineeringdaily.com/2020/03/12/flyte-lyft-data-processing-platform-with-allyson-gale-and-ketan-umare/)
- MLOps Coffee session - [Flyte: an open-source tool for scalable, extensible, and portable workflows](https://anchor.fm/mlops/episodes/MLOps-Coffee-Sessions-12-Flyte-an-open-source-tool-for-scalable--extensible---and-portable-workflows-eksa5k)

## 💖 Top Contributors

A big thank you to the community for making Flyte possible!

- [@wild-endeavor](https://github.com/wild-endeavor)
- [@katrogan](https://github.com/katrogan)
- [@EngHabu](https://github.com/EngHabu)
- [@akhurana001](https://github.com/akhurana001)
- [@anandswaminathan](https://github.com/anandswaminathan)
- [@kanterov](https://github.com/kanterov)
- [@honnix](https://github.com/honnix)
- [@jeevb](https://github.com/jeevb)
- [@jonathanburns](https://github.com/jonathanburns)
- [@migueltol22](https://github.com/migueltol22)
- [@varshaparthay](https://github.com/varshaparthay)
- [@pingsutw](https://github.com/pingsutw)
- [@narape](https://github.com/narape)
- [@lu4nm3](https://github.com/lu4nm3)
- [@bnsblue](https://github.com/bnsblue)
- [@RubenBarragan](https://github.com/RubenBarragan)
- [@schottra](https://github.com/schottra)
- [@evalsocket](https://github.com/evalsocket)
- [@matthewphsmith](https://github.com/matthewphsmith)
- [@slai](https://github.com/slai)
- [@derwiki](https://github.com/derwiki)
- [@tnsetting](https://github.com/tnsetting)
- [@jbrambleDC](https://github.com/jbrambleDC)
- [@igorvalko](https://github.com/igorvalko)
- [@chanadian](https://github.com/chanadian)
- [@surindersinghp](https://github.com/surindersinghp)
- [@vsbus](https://github.com/vsbus)
- [@catalinii](https://github.com/catalinii)
- [@pmahindrakar-oss](https://github.com/pmahindrakar-oss)
- [@milton0825](https://github.com/milton0825)
- [@kumare3](https://github.com/kumare3)


<!-- CONTRIBUTORS START -->


<a href="https://github.com/katrogan">
    <img src="https://avatars.githubusercontent.com/u/953358?v=4" width="64" height="64" alt="953358" style="border-radius: 50px;">
</a>



<a href="https://github.com/lyft-metaservice-3">
    <img src="https://avatars.githubusercontent.com/u/37090125?v=4" width="64" height="64" alt="37090125" style="border-radius: 50px;">
</a>



<a href="https://github.com/matthewphsmith">
    <img src="https://avatars.githubusercontent.com/u/7597118?v=4" width="64" height="64" alt="7597118" style="border-radius: 50px;">
</a>



<a href="https://github.com/EngHabu">
    <img src="https://avatars.githubusercontent.com/u/27159?v=4" width="64" height="64" alt="27159" style="border-radius: 50px;">
</a>



<a href="https://github.com/goreleaserbot">
    <img src="https://avatars.githubusercontent.com/u/29843943?v=4" width="64" height="64" alt="29843943" style="border-radius: 50px;">
</a>






<a href="https://github.com/honnix">
    <img src="https://avatars.githubusercontent.com/u/158892?v=4" width="64" height="64" alt="158892" style="border-radius: 50px;">
</a>



<a href="https://github.com/anandswaminathan">
    <img src="https://avatars.githubusercontent.com/u/18408237?v=4" width="64" height="64" alt="18408237" style="border-radius: 50px;">
</a>



<a href="https://github.com/wild-endeavor">
    <img src="https://avatars.githubusercontent.com/u/2896568?v=4" width="64" height="64" alt="2896568" style="border-radius: 50px;">
</a>



<a href="https://github.com/evalsocket">
    <img src="https://avatars.githubusercontent.com/u/10830562?v=4" width="64" height="64" alt="10830562" style="border-radius: 50px;">
</a>



<a href="https://github.com/bnsblue">
    <img src="https://avatars.githubusercontent.com/u/1518524?v=4" width="64" height="64" alt="1518524" style="border-radius: 50px;">
</a>



<a href="https://github.com/flyte-bot">
    <img src="https://avatars.githubusercontent.com/u/78108056?v=4" width="64" height="64" alt="78108056" style="border-radius: 50px;">
</a>



<a href="https://github.com/samhita-alla">
    <img src="https://avatars.githubusercontent.com/u/27777173?v=4" width="64" height="64" alt="27777173" style="border-radius: 50px;">
</a>



<a href="https://github.com/kumare3">
    <img src="https://avatars.githubusercontent.com/u/16888709?v=4" width="64" height="64" alt="16888709" style="border-radius: 50px;">
</a>



<a href="https://github.com/brucearctor">
    <img src="https://avatars.githubusercontent.com/u/5032356?v=4" width="64" height="64" alt="5032356" style="border-radius: 50px;">
</a>



<a href="https://github.com/jonathanburns">
    <img src="https://avatars.githubusercontent.com/u/3880645?v=4" width="64" height="64" alt="3880645" style="border-radius: 50px;">
</a>



<a href="https://github.com/lu4nm3">
    <img src="https://avatars.githubusercontent.com/u/3936213?v=4" width="64" height="64" alt="3936213" style="border-radius: 50px;">
</a>



<a href="https://github.com/lyft-metaservice-2">
    <img src="https://avatars.githubusercontent.com/u/26174213?v=4" width="64" height="64" alt="26174213" style="border-radius: 50px;">
</a>



<a href="https://github.com/pmahindrakar-oss">
    <img src="https://avatars.githubusercontent.com/u/77798312?v=4" width="64" height="64" alt="77798312" style="border-radius: 50px;">
</a>



<a href="https://github.com/2uasimojo">
    <img src="https://avatars.githubusercontent.com/u/9142716?v=4" width="64" height="64" alt="9142716" style="border-radius: 50px;">
</a>



<a href="https://github.com/veggiemonk">
    <img src="https://avatars.githubusercontent.com/u/5487021?v=4" width="64" height="64" alt="5487021" style="border-radius: 50px;">
</a>



<a href="https://github.com/schottra">
    <img src="https://avatars.githubusercontent.com/u/1815175?v=4" width="64" height="64" alt="1815175" style="border-radius: 50px;">
</a>



<a href="https://github.com/migueltol22">
    <img src="https://avatars.githubusercontent.com/u/19375241?v=4" width="64" height="64" alt="19375241" style="border-radius: 50px;">
</a>



<a href="https://github.com/cosmicBboy">
    <img src="https://avatars.githubusercontent.com/u/2816689?v=4" width="64" height="64" alt="2816689" style="border-radius: 50px;">
</a>



<a href="https://github.com/pingsutw">
    <img src="https://avatars.githubusercontent.com/u/37936015?v=4" width="64" height="64" alt="37936015" style="border-radius: 50px;">
</a>



<a href="https://github.com/mayitbeegh">
    <img src="https://avatars.githubusercontent.com/u/6239450?v=4" width="64" height="64" alt="6239450" style="border-radius: 50px;">
</a>



<a href="https://github.com/milton0825">
    <img src="https://avatars.githubusercontent.com/u/6065051?v=4" width="64" height="64" alt="6065051" style="border-radius: 50px;">
</a>



<a href="https://github.com/slai">
    <img src="https://avatars.githubusercontent.com/u/70988?v=4" width="64" height="64" alt="70988" style="border-radius: 50px;">
</a>



<a href="https://github.com/surindersinghp">
    <img src="https://avatars.githubusercontent.com/u/16090976?v=4" width="64" height="64" alt="16090976" style="border-radius: 50px;">
</a>



<a href="https://github.com/hamersaw">
    <img src="https://avatars.githubusercontent.com/u/8888115?v=4" width="64" height="64" alt="8888115" style="border-radius: 50px;">
</a>



<a href="https://github.com/kosigz-lyft">
    <img src="https://avatars.githubusercontent.com/u/53313394?v=4" width="64" height="64" alt="53313394" style="border-radius: 50px;">
</a>



<a href="https://github.com/chanadian">
    <img src="https://avatars.githubusercontent.com/u/4967458?v=4" width="64" height="64" alt="4967458" style="border-radius: 50px;">
</a>



<a href="https://github.com/ckiosidis">
    <img src="https://avatars.githubusercontent.com/u/6562898?v=4" width="64" height="64" alt="6562898" style="border-radius: 50px;">
</a>



<a href="https://github.com/kanterov">
    <img src="https://avatars.githubusercontent.com/u/467927?v=4" width="64" height="64" alt="467927" style="border-radius: 50px;">
</a>



<a href="https://github.com/hanzo">
    <img src="https://avatars.githubusercontent.com/u/248688?v=4" width="64" height="64" alt="248688" style="border-radius: 50px;">
</a>



<a href="https://github.com/igorvalko">
    <img src="https://avatars.githubusercontent.com/u/1330233?v=4" width="64" height="64" alt="1330233" style="border-radius: 50px;">
</a>



<a href="https://github.com/vsbus">
    <img src="https://avatars.githubusercontent.com/u/5026554?v=4" width="64" height="64" alt="5026554" style="border-radius: 50px;">
</a>



<a href="https://github.com/tnsetting">
    <img src="https://avatars.githubusercontent.com/u/38207208?v=4" width="64" height="64" alt="38207208" style="border-radius: 50px;">
</a>



<a href="https://github.com/akhurana001">
    <img src="https://avatars.githubusercontent.com/u/34587798?v=4" width="64" height="64" alt="34587798" style="border-radius: 50px;">
</a>



<a href="https://github.com/catalinii">
    <img src="https://avatars.githubusercontent.com/u/8200209?v=4" width="64" height="64" alt="8200209" style="border-radius: 50px;">
</a>



<a href="https://github.com/chetcode">
    <img src="https://avatars.githubusercontent.com/u/43587819?v=4" width="64" height="64" alt="43587819" style="border-radius: 50px;">
</a>



<a href="https://github.com/jeevb">
    <img src="https://avatars.githubusercontent.com/u/10869815?v=4" width="64" height="64" alt="10869815" style="border-radius: 50px;">
</a>



<a href="https://github.com/regadas">
    <img src="https://avatars.githubusercontent.com/u/163899?v=4" width="64" height="64" alt="163899" style="border-radius: 50px;">
</a>



<a href="https://github.com/akashkatipally">
    <img src="https://avatars.githubusercontent.com/u/1316881?v=4" width="64" height="64" alt="1316881" style="border-radius: 50px;">
</a>



<a href="https://github.com/clairemcginty">
    <img src="https://avatars.githubusercontent.com/u/1360529?v=4" width="64" height="64" alt="1360529" style="border-radius: 50px;">
</a>



<a href="https://github.com/eapolinario">
    <img src="https://avatars.githubusercontent.com/u/653394?v=4" width="64" height="64" alt="653394" style="border-radius: 50px;">
</a>



<a href="https://github.com/derwiki">
    <img src="https://avatars.githubusercontent.com/u/155087?v=4" width="64" height="64" alt="155087" style="border-radius: 50px;">
</a>



<a href="https://github.com/th0114nd">
    <img src="https://avatars.githubusercontent.com/u/1399455?v=4" width="64" height="64" alt="1399455" style="border-radius: 50px;">
</a>



<a href="https://github.com/asottile">
    <img src="https://avatars.githubusercontent.com/u/1810591?v=4" width="64" height="64" alt="1810591" style="border-radius: 50px;">
</a>



<a href="https://github.com/SandraGH5">
    <img src="https://avatars.githubusercontent.com/u/80421934?v=4" width="64" height="64" alt="80421934" style="border-radius: 50px;">
</a>



<a href="https://github.com/fediazgon">
    <img src="https://avatars.githubusercontent.com/u/12219405?v=4" width="64" height="64" alt="12219405" style="border-radius: 50px;">
</a>



<a href="https://github.com/sbrunk">
    <img src="https://avatars.githubusercontent.com/u/3939659?v=4" width="64" height="64" alt="3939659" style="border-radius: 50px;">
</a>



<a href="https://github.com/ddhirajkumar">
    <img src="https://avatars.githubusercontent.com/u/6774758?v=4" width="64" height="64" alt="6774758" style="border-radius: 50px;">
</a>



<a href="https://github.com/max-hoffman">
    <img src="https://avatars.githubusercontent.com/u/18337807?v=4" width="64" height="64" alt="18337807" style="border-radius: 50px;">
</a>



<a href="https://github.com/michaels-lyft">
    <img src="https://avatars.githubusercontent.com/u/54334265?v=4" width="64" height="64" alt="54334265" style="border-radius: 50px;">
</a>



<a href="https://github.com/sonjaer">
    <img src="https://avatars.githubusercontent.com/u/9609986?v=4" width="64" height="64" alt="9609986" style="border-radius: 50px;">
</a>



<a href="https://github.com/AdrianoKF">
    <img src="https://avatars.githubusercontent.com/u/322624?v=4" width="64" height="64" alt="322624" style="border-radius: 50px;">
</a>



<a href="https://github.com/aalavian">
    <img src="https://avatars.githubusercontent.com/u/54333860?v=4" width="64" height="64" alt="54333860" style="border-radius: 50px;">
</a>



<a href="https://github.com/convexquad">
    <img src="https://avatars.githubusercontent.com/u/7005765?v=4" width="64" height="64" alt="7005765" style="border-radius: 50px;">
</a>



<a href="https://github.com/asahalyft">
    <img src="https://avatars.githubusercontent.com/u/48966647?v=4" width="64" height="64" alt="48966647" style="border-radius: 50px;">
</a>



<a href="https://github.com/apatel-fn">
    <img src="https://avatars.githubusercontent.com/u/77167782?v=4" width="64" height="64" alt="77167782" style="border-radius: 50px;">
</a>



<a href="https://github.com/lordnodd">
    <img src="https://avatars.githubusercontent.com/u/31381038?v=4" width="64" height="64" alt="31381038" style="border-radius: 50px;">
</a>



<a href="https://github.com/frsann">
    <img src="https://avatars.githubusercontent.com/u/7358951?v=4" width="64" height="64" alt="7358951" style="border-radius: 50px;">
</a>



<a href="https://github.com/jbrambleDC">
    <img src="https://avatars.githubusercontent.com/u/6984748?v=4" width="64" height="64" alt="6984748" style="border-radius: 50px;">
</a>



<a href="https://github.com/tekumara">
    <img src="https://avatars.githubusercontent.com/u/125105?v=4" width="64" height="64" alt="125105" style="border-radius: 50px;">
</a>



<a href="https://github.com/rubenbarragan">
    <img src="https://avatars.githubusercontent.com/u/4308533?v=4" width="64" height="64" alt="4308533" style="border-radius: 50px;">
</a>



<a href="https://github.com/sushrut111">
    <img src="https://avatars.githubusercontent.com/u/11269256?v=4" width="64" height="64" alt="11269256" style="border-radius: 50px;">
</a>



<a href="https://github.com/Tat-V">
    <img src="https://avatars.githubusercontent.com/u/61228633?v=4" width="64" height="64" alt="61228633" style="border-radius: 50px;">
</a>



<a href="https://github.com/ThomVett">
    <img src="https://avatars.githubusercontent.com/u/8817639?v=4" width="64" height="64" alt="8817639" style="border-radius: 50px;">
</a>



<a href="https://github.com/datability-io">
    <img src="https://avatars.githubusercontent.com/u/17309187?v=4" width="64" height="64" alt="17309187" style="border-radius: 50px;">
</a>



<a href="https://github.com/varshaparthay">
    <img src="https://avatars.githubusercontent.com/u/57967031?v=4" width="64" height="64" alt="57967031" style="border-radius: 50px;">
</a>



<a href="https://github.com/ybubnov">
    <img src="https://avatars.githubusercontent.com/u/1778407?v=4" width="64" height="64" alt="1778407" style="border-radius: 50px;">
</a>



<a href="https://github.com/ajsalow">
    <img src="https://avatars.githubusercontent.com/u/12450632?v=4" width="64" height="64" alt="12450632" style="border-radius: 50px;">
</a>



<a href="https://github.com/apps/dependabot">
    <img src="https://avatars.githubusercontent.com/in/29110?v=4" width="64" height="64" alt="49699333" style="border-radius: 50px;">
</a>



<a href="https://github.com/service-github-lyft-semantic-release">
    <img src="https://avatars.githubusercontent.com/u/65977800?v=4" width="64" height="64" alt="65977800" style="border-radius: 50px;">
</a>



<a href="https://github.com/csirius">
    <img src="https://avatars.githubusercontent.com/u/85753828?v=4" width="64" height="64" alt="85753828" style="border-radius: 50px;">
</a>



<a href="https://github.com/Pianist038801">
    <img src="https://avatars.githubusercontent.com/u/26953709?v=4" width="64" height="64" alt="26953709" style="border-radius: 50px;">
</a>



<a href="https://github.com/jsonporter">
    <img src="https://avatars.githubusercontent.com/u/84735036?v=4" width="64" height="64" alt="84735036" style="border-radius: 50px;">
</a>



<a href="https://github.com/aviaviavi">
    <img src="https://avatars.githubusercontent.com/u/1388071?v=4" width="64" height="64" alt="1388071" style="border-radius: 50px;">
</a>



<a href="https://github.com/Professional0321">
    <img src="https://avatars.githubusercontent.com/u/58770001?v=4" width="64" height="64" alt="58770001" style="border-radius: 50px;">
</a>



<a href="https://github.com/snyk-bot">
    <img src="https://avatars.githubusercontent.com/u/19733683?v=4" width="64" height="64" alt="19733683" style="border-radius: 50px;">
</a>



<a href="https://github.com/jimbobby5">
    <img src="https://avatars.githubusercontent.com/u/18363301?v=4" width="64" height="64" alt="18363301" style="border-radius: 50px;">
</a>



<a href="https://github.com/pradithya">
    <img src="https://avatars.githubusercontent.com/u/4023015?v=4" width="64" height="64" alt="4023015" style="border-radius: 50px;">
</a>



<a href="https://github.com/vvasavada-fn">
    <img src="https://avatars.githubusercontent.com/u/67166843?v=4" width="64" height="64" alt="67166843" style="border-radius: 50px;">
</a>



<a href="https://github.com/ariefrahmansyah">
    <img src="https://avatars.githubusercontent.com/u/8122852?v=4" width="64" height="64" alt="8122852" style="border-radius: 50px;">
</a>



<a href="https://github.com/haoyuez">
    <img src="https://avatars.githubusercontent.com/u/25364490?v=4" width="64" height="64" alt="25364490" style="border-radius: 50px;">
</a>



<a href="https://github.com/narape">
    <img src="https://avatars.githubusercontent.com/u/7515359?v=4" width="64" height="64" alt="7515359" style="border-radius: 50px;">
</a>



<a href="https://github.com/arturdryomov">
    <img src="https://avatars.githubusercontent.com/u/200401?v=4" width="64" height="64" alt="200401" style="border-radius: 50px;">
</a>



<a href="https://github.com/felixwang9817">
    <img src="https://avatars.githubusercontent.com/u/24739949?v=4" width="64" height="64" alt="24739949" style="border-radius: 50px;">
</a>



<a href="https://github.com/juandiegopalomino">
    <img src="https://avatars.githubusercontent.com/u/10430635?v=4" width="64" height="64" alt="10430635" style="border-radius: 50px;">
</a>



<a href="https://github.com/georgesnelling">
    <img src="https://avatars.githubusercontent.com/u/405480?v=4" width="64" height="64" alt="405480" style="border-radius: 50px;">
</a>



<a href="https://github.com/dschaller">
    <img src="https://avatars.githubusercontent.com/u/1004789?v=4" width="64" height="64" alt="1004789" style="border-radius: 50px;">
</a>



<a href="https://github.com/hoyajigi">
    <img src="https://avatars.githubusercontent.com/u/1335881?v=4" width="64" height="64" alt="1335881" style="border-radius: 50px;">
</a>



<a href="https://github.com/NitinAgg">
    <img src="https://avatars.githubusercontent.com/u/4830700?v=4" width="64" height="64" alt="4830700" style="border-radius: 50px;">
</a>



<a href="https://github.com/noobkid2411">
    <img src="https://avatars.githubusercontent.com/u/69161722?v=4" width="64" height="64" alt="69161722" style="border-radius: 50px;">
</a>



<a href="https://github.com/stephen37">
    <img src="https://avatars.githubusercontent.com/u/6506810?v=4" width="64" height="64" alt="6506810" style="border-radius: 50px;">
</a>



<a href="https://github.com/eanakhl">
    <img src="https://avatars.githubusercontent.com/u/14992189?v=4" width="64" height="64" alt="14992189" style="border-radius: 50px;">
</a>



<a href="https://github.com/adinin">
    <img src="https://avatars.githubusercontent.com/u/1175392?v=4" width="64" height="64" alt="1175392" style="border-radius: 50px;">
</a>



<a href="https://github.com/anton-malakhov">
    <img src="https://avatars.githubusercontent.com/u/7475946?v=4" width="64" height="64" alt="7475946" style="border-radius: 50px;">
</a>



<a href="https://github.com/kinow">
    <img src="https://avatars.githubusercontent.com/u/304786?v=4" width="64" height="64" alt="304786" style="border-radius: 50px;">
</a>



<a href="https://github.com/Daeruin">
    <img src="https://avatars.githubusercontent.com/u/24402505?v=4" width="64" height="64" alt="24402505" style="border-radius: 50px;">
</a>



<a href="https://github.com/JakeNeyer">
    <img src="https://avatars.githubusercontent.com/u/16461847?v=4" width="64" height="64" alt="16461847" style="border-radius: 50px;">
</a>



<a href="https://github.com/akumor">
    <img src="https://avatars.githubusercontent.com/u/2538760?v=4" width="64" height="64" alt="2538760" style="border-radius: 50px;">
</a>



<a href="https://github.com/jeremydonahue">
    <img src="https://avatars.githubusercontent.com/u/14008978?v=4" width="64" height="64" alt="14008978" style="border-radius: 50px;">
</a>



<a href="https://github.com/lsena">
    <img src="https://avatars.githubusercontent.com/u/19229049?v=4" width="64" height="64" alt="19229049" style="border-radius: 50px;">
</a>



<a href="https://github.com/Dread1982">
    <img src="https://avatars.githubusercontent.com/u/7548823?v=4" width="64" height="64" alt="7548823" style="border-radius: 50px;">
</a>



<a href="https://github.com/ilikedata">
    <img src="https://avatars.githubusercontent.com/u/580328?v=4" width="64" height="64" alt="580328" style="border-radius: 50px;">
</a>



<a href="https://github.com/orf">
    <img src="https://avatars.githubusercontent.com/u/1027207?v=4" width="64" height="64" alt="1027207" style="border-radius: 50px;">
</a>



<a href="https://github.com/leorleor">
    <img src="https://avatars.githubusercontent.com/u/1568889?v=4" width="64" height="64" alt="1568889" style="border-radius: 50px;">
</a>



<a href="https://github.com/moose007">
    <img src="https://avatars.githubusercontent.com/u/937967?v=4" width="64" height="64" alt="937967" style="border-radius: 50px;">
</a>



<a href="https://github.com/nicklofaso">
    <img src="https://avatars.githubusercontent.com/u/25391173?v=4" width="64" height="64" alt="25391173" style="border-radius: 50px;">
</a>



<a href="https://github.com/v01dXYZ">
    <img src="https://avatars.githubusercontent.com/u/14996868?v=4" width="64" height="64" alt="14996868" style="border-radius: 50px;">
</a>



<a href="https://github.com/kylewaynebenson">
    <img src="https://avatars.githubusercontent.com/u/1043051?v=4" width="64" height="64" alt="1043051" style="border-radius: 50px;">
</a>



<a href="https://github.com/mouuff">
    <img src="https://avatars.githubusercontent.com/u/1174730?v=4" width="64" height="64" alt="1174730" style="border-radius: 50px;">
</a>



<a href="https://github.com/idivyanshbansal">
    <img src="https://avatars.githubusercontent.com/u/86911142?v=4" width="64" height="64" alt="86911142" style="border-radius: 50px;">
</a>



<a href="https://github.com/vglocus">
    <img src="https://avatars.githubusercontent.com/u/697033?v=4" width="64" height="64" alt="697033" style="border-radius: 50px;">
</a>



<a href="https://github.com/stormy-ua">
    <img src="https://avatars.githubusercontent.com/u/5732047?v=4" width="64" height="64" alt="5732047" style="border-radius: 50px;">
</a>



<a href="https://github.com/gdungca-fn">
    <img src="https://avatars.githubusercontent.com/u/71284190?v=4" width="64" height="64" alt="71284190" style="border-radius: 50px;">
</a>



<a href="https://github.com/ttanay">
    <img src="https://avatars.githubusercontent.com/u/26265392?v=4" width="64" height="64" alt="26265392" style="border-radius: 50px;">
</a>



<a href="https://github.com/pradyunsg">
    <img src="https://avatars.githubusercontent.com/u/3275593?v=4" width="64" height="64" alt="3275593" style="border-radius: 50px;">
</a>



<a href="https://github.com/apps/pre-commit-ci">
    <img src="https://avatars.githubusercontent.com/in/68672?v=4" width="64" height="64" alt="66853113" style="border-radius: 50px;">
</a>



<a href="https://github.com/kmike">
    <img src="https://avatars.githubusercontent.com/u/107893?v=4" width="64" height="64" alt="107893" style="border-radius: 50px;">
</a>



<a href="https://github.com/sirosen">
    <img src="https://avatars.githubusercontent.com/u/1300022?v=4" width="64" height="64" alt="1300022" style="border-radius: 50px;">
</a>



<a href="https://github.com/hugovk">
    <img src="https://avatars.githubusercontent.com/u/1324225?v=4" width="64" height="64" alt="1324225" style="border-radius: 50px;">
</a>



<a href="https://github.com/bastimeyer">
    <img src="https://avatars.githubusercontent.com/u/467294?v=4" width="64" height="64" alt="467294" style="border-radius: 50px;">
</a>



<a href="https://github.com/drewyh">
    <img src="https://avatars.githubusercontent.com/u/20280470?v=4" width="64" height="64" alt="20280470" style="border-radius: 50px;">
</a>



<a href="https://github.com/dvarrazzo">
    <img src="https://avatars.githubusercontent.com/u/199429?v=4" width="64" height="64" alt="199429" style="border-radius: 50px;">
</a>



<a href="https://github.com/dbitouze">
    <img src="https://avatars.githubusercontent.com/u/1032633?v=4" width="64" height="64" alt="1032633" style="border-radius: 50px;">
</a>



<a href="https://github.com/sethmlarson">
    <img src="https://avatars.githubusercontent.com/u/18519037?v=4" width="64" height="64" alt="18519037" style="border-radius: 50px;">
</a>



<a href="https://github.com/stonecharioteer">
    <img src="https://avatars.githubusercontent.com/u/11478411?v=4" width="64" height="64" alt="11478411" style="border-radius: 50px;">
</a>



<a href="https://github.com/estan">
    <img src="https://avatars.githubusercontent.com/u/86675?v=4" width="64" height="64" alt="86675" style="border-radius: 50px;">
</a>



<a href="https://github.com/pseudomuto">
    <img src="https://avatars.githubusercontent.com/u/4748863?v=4" width="64" height="64" alt="4748863" style="border-radius: 50px;">
</a>



<a href="https://github.com/htdvisser">
    <img src="https://avatars.githubusercontent.com/u/181308?v=4" width="64" height="64" alt="181308" style="border-radius: 50px;">
</a>



<a href="https://github.com/jacobtolar">
    <img src="https://avatars.githubusercontent.com/u/1390277?v=4" width="64" height="64" alt="1390277" style="border-radius: 50px;">
</a>



<a href="https://github.com/ezimanyi">
    <img src="https://avatars.githubusercontent.com/u/1391982?v=4" width="64" height="64" alt="1391982" style="border-radius: 50px;">
</a>



<a href="https://github.com/lpabon">
    <img src="https://avatars.githubusercontent.com/u/3880001?v=4" width="64" height="64" alt="3880001" style="border-radius: 50px;">
</a>



<a href="https://github.com/ArcEye">
    <img src="https://avatars.githubusercontent.com/u/770392?v=4" width="64" height="64" alt="770392" style="border-radius: 50px;">
</a>



<a href="https://github.com/mingrammer">
    <img src="https://avatars.githubusercontent.com/u/6178510?v=4" width="64" height="64" alt="6178510" style="border-radius: 50px;">
</a>



<a href="https://github.com/aschrijver">
    <img src="https://avatars.githubusercontent.com/u/5111931?v=4" width="64" height="64" alt="5111931" style="border-radius: 50px;">
</a>



<a href="https://github.com/panzerfahrer">
    <img src="https://avatars.githubusercontent.com/u/873434?v=4" width="64" height="64" alt="873434" style="border-radius: 50px;">
</a>



<a href="https://github.com/masterzen">
    <img src="https://avatars.githubusercontent.com/u/20242?v=4" width="64" height="64" alt="20242" style="border-radius: 50px;">
</a>



<a href="https://github.com/glasser">
    <img src="https://avatars.githubusercontent.com/u/16724?v=4" width="64" height="64" alt="16724" style="border-radius: 50px;">
</a>



<a href="https://github.com/murph0">
    <img src="https://avatars.githubusercontent.com/u/17330872?v=4" width="64" height="64" alt="17330872" style="border-radius: 50px;">
</a>



<a href="https://github.com/zetaron">
    <img src="https://avatars.githubusercontent.com/u/419419?v=4" width="64" height="64" alt="419419" style="border-radius: 50px;">
</a>



<a href="https://github.com/sunfmin">
    <img src="https://avatars.githubusercontent.com/u/1014?v=4" width="64" height="64" alt="1014" style="border-radius: 50px;">
</a>



<a href="https://github.com/guozheng">
    <img src="https://avatars.githubusercontent.com/u/504507?v=4" width="64" height="64" alt="504507" style="border-radius: 50px;">
</a>



<a href="https://github.com/nagytech">
    <img src="https://avatars.githubusercontent.com/u/8542033?v=4" width="64" height="64" alt="8542033" style="border-radius: 50px;">
</a>



<a href="https://github.com/suusan2go">
    <img src="https://avatars.githubusercontent.com/u/8841470?v=4" width="64" height="64" alt="8841470" style="border-radius: 50px;">
</a>



<a href="https://github.com/mhaberler">
    <img src="https://avatars.githubusercontent.com/u/901479?v=4" width="64" height="64" alt="901479" style="border-radius: 50px;">
</a>



<a href="https://github.com/s4ichi">
    <img src="https://avatars.githubusercontent.com/u/6400253?v=4" width="64" height="64" alt="6400253" style="border-radius: 50px;">
</a>



<a href="https://github.com/dreampuf">
    <img src="https://avatars.githubusercontent.com/u/353644?v=4" width="64" height="64" alt="353644" style="border-radius: 50px;">
</a>



<a href="https://github.com/UnicodingUnicorn">
    <img src="https://avatars.githubusercontent.com/u/12421077?v=4" width="64" height="64" alt="12421077" style="border-radius: 50px;">
</a>



<a href="https://github.com/philiptzou">
    <img src="https://avatars.githubusercontent.com/u/809865?v=4" width="64" height="64" alt="809865" style="border-radius: 50px;">
</a>



<a href="https://github.com/timabell">
    <img src="https://avatars.githubusercontent.com/u/19378?v=4" width="64" height="64" alt="19378" style="border-radius: 50px;">
</a>


<!-- CONTRIBUTORS END -->