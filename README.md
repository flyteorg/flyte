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

Flyte is a structured programming and distributed processing platform that enables highly concurrent, scalable, and maintainable workflows for `Machine Learning` and `Data Processing`. It is a fabric that connects disparate computation backends using a type-safe data dependency graph. It records all changes to a pipeline, making it possible to rewind time. It also stores
a history of all executions and provides an intuitive UI, CLI, and REST/gRPC API to interact with the computation.

Flyte is more than a workflow engine -- it uses `workflow` as a core concept and `task` (a single unit of execution) as a top-level concept. Multiple tasks arranged in a data
producer-consumer order creates a workflow.

`Workflows` and `Tasks` can be written in any language, with out-of-the-box support for [Python](https://github.com/flyteorg/flytekit), [Java and Scala](https://github.com/spotify/flytekit-java).


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
** Please maintain the order alphabetically **

- [Blackshark.ai](https://blackshark.ai/)
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
- A data-aware platform
- Enables **collaboration across your organization** by:
  - Executing distributed data pipelines/workflows
  - Reusing tasks across projects, users, and workflows
  - Making it easy to stitch together workflows from different teams and domain experts
  - Backtracing to a specified workflow
  - Comparing results of training workflows over time and across pipelines
  - Sharing workflows and tasks across your teams
  - Simplifying the complexity of multi-step, multi-owner workflows
- **[Quick registration](https://docs.flyte.org/en/latest/getting_started.html)** -- start locally and scale to the cloud instantly
- **Centralized Inventory** constituting Tasks, Workflows, and Executions
- **gRPC / REST** interface to define and execute tasks and workflows
- **Type safe** construction of pipelines -- each task has an interface that is characterized by its input and output, so illegal construction of pipelines fails during declaration rather than at runtime
- Supports multiple **[data types](https://docs.flyte.org/projects/cookbook/en/latest/auto/type_system/index.html)** for machine learning and data processing pipelines, such as Blobs (images, arbitrary files), Directories, Schema (columnar structured data), collections, maps, etc.
- Memoization and Lineage tracking
- Provides logging and observability
- Workflow features:
  - Start with one task, convert to a pipeline, attach **[multiple schedules](https://docs.flyte.org/projects/cookbook/en/latest/auto/deployment/workflow/lp_schedules.html)**, trigger using a programmatic API, or on-demand
  - Parallel step execution
  - Extensible backend to add **[customized plugin](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/extend_flyte/custom_task_plugin.html)** experience (with simplified user experience)
  - **[Branching](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/control_flow/run_conditions.html)**
 - Inline **[subworkflows](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/control_flow/subworkflows.html)** (a workflow can be embedded within one node of the top-level workflow)
  - Distributed **remote child workflows** (a remote workflow can be triggered and statically verified at compile time)
  - **[Array Tasks](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/control_flow/map_task.html)** (map a function over a large dataset -- ensures controlled execution of thousands of containers)
  - **[Dynamic workflow](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/control_flow/dynamics.html)** creation and execution with runtime type safety
  - Container side [plugins](https://docs.flyte.org/projects/cookbook/en/latest/plugins.html) with first-class support in Python
  - _PreAlpha_: Arbitrary flytekit-less containers supported ([RawContainer](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/containerization/raw_container.html))
- Guaranteed **[reproducibility](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/flyte_basics/task_cache.html)** of pipelines via:
  - Versioned data, code, and models
  - Automatically tracked executions
  - Declarative pipelines
- **Multi-cloud support** (AWS, GCP, and others)
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
- Distributed Tensorflow (K8s Native)
- Papermill notebook execution ([Python](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/flytekit_plugins/papermilltasks/index.html) and Spark)
- [Type safe and data checking for Pandas dataframe](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/flytekit_plugins/pandera_examples/index.html) using Pandera
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

- 📣 **Flyte OSS Community Sync** Every other Tuesday, 9am-10am PDT. Check out the [calendar](https://www.addevent.com/calendar/kE355955) and register to stay up-to-date with our meeting times. Or simply join us on [Zoom](https://us04web.zoom.us/j/71298741279?pwd=TDR1RUppQmxGaDRFdzBOa2lHN1dsZz09).
- Upcoming meeting agenda, previous meeting notes, and a backlog of topics are captured in this [document](https://docs.google.com/document/d/1Jb6eOPOzvTaHjtPEVy7OR2O5qK1MhEs3vv56DX2dacM/edit#heading=h.c5ha25xc546e).
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

## 💖 All Contributors

A big thank you to the community for making Flyte possible!

<!-- CONTRIBUTORS START -->
[![953358](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/953358?v=4&w=50&h=50&mask=circle)](https://github.com/katrogan)[![37090125](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/37090125?v=4&w=50&h=50&mask=circle)](https://github.com/lyft-metaservice-3)[![7597118](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/7597118?v=4&w=50&h=50&mask=circle)](https://github.com/matthewphsmith)[![27159](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/27159?v=4&w=50&h=50&mask=circle)](https://github.com/EngHabu)[![29843943](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/29843943?v=4&w=50&h=50&mask=circle)](https://github.com/goreleaserbot)[![158892](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/158892?v=4&w=50&h=50&mask=circle)](https://github.com/honnix)[![18408237](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/18408237?v=4&w=50&h=50&mask=circle)](https://github.com/anandswaminathan)[![2896568](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/2896568?v=4&w=50&h=50&mask=circle)](https://github.com/wild-endeavor)[![10830562](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/10830562?v=4&w=50&h=50&mask=circle)](https://github.com/evalsocket)[![1518524](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1518524?v=4&w=50&h=50&mask=circle)](https://github.com/bnsblue)[![78108056](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/78108056?v=4&w=50&h=50&mask=circle)](https://github.com/flyte-bot)[![27777173](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/27777173?v=4&w=50&h=50&mask=circle)](https://github.com/samhita-alla)[![16888709](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/16888709?v=4&w=50&h=50&mask=circle)](https://github.com/kumare3)[![5032356](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/5032356?v=4&w=50&h=50&mask=circle)](https://github.com/brucearctor)[![8122852](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/8122852?v=4&w=50&h=50&mask=circle)](https://github.com/ariefrahmansyah)[![3880645](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/3880645?v=4&w=50&h=50&mask=circle)](https://github.com/jonathanburns)[![3936213](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/3936213?v=4&w=50&h=50&mask=circle)](https://github.com/lu4nm3)[![26174213](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/26174213?v=4&w=50&h=50&mask=circle)](https://github.com/lyft-metaservice-2)[![77798312](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/77798312?v=4&w=50&h=50&mask=circle)](https://github.com/pmahindrakar-oss)[![9142716](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/9142716?v=4&w=50&h=50&mask=circle)](https://github.com/2uasimojo)[![5487021](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/5487021?v=4&w=50&h=50&mask=circle)](https://github.com/veggiemonk)[![1815175](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1815175?v=4&w=50&h=50&mask=circle)](https://github.com/schottra)[![19375241](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/19375241?v=4&w=50&h=50&mask=circle)](https://github.com/migueltol22)[![2816689](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/2816689?v=4&w=50&h=50&mask=circle)](https://github.com/cosmicBboy)[![37936015](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/37936015?v=4&w=50&h=50&mask=circle)](https://github.com/pingsutw)[![6239450](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/6239450?v=4&w=50&h=50&mask=circle)](https://github.com/mayitbeegh)[![6065051](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/6065051?v=4&w=50&h=50&mask=circle)](https://github.com/milton0825)[![70988](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/70988?v=4&w=50&h=50&mask=circle)](https://github.com/slai)[![16090976](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/16090976?v=4&w=50&h=50&mask=circle)](https://github.com/surindersinghp)[![8888115](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/8888115?v=4&w=50&h=50&mask=circle)](https://github.com/hamersaw)[![53313394](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/53313394?v=4&w=50&h=50&mask=circle)](https://github.com/kosigz-lyft)[![4967458](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/4967458?v=4&w=50&h=50&mask=circle)](https://github.com/chanadian)[![6562898](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/6562898?v=4&w=50&h=50&mask=circle)](https://github.com/ckiosidis)[![467927](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/467927?v=4&w=50&h=50&mask=circle)](https://github.com/kanterov)[![248688](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/248688?v=4&w=50&h=50&mask=circle)](https://github.com/hanzo)[![1330233](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1330233?v=4&w=50&h=50&mask=circle)](https://github.com/igorvalko)[![5026554](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/5026554?v=4&w=50&h=50&mask=circle)](https://github.com/vsbus)[![38207208](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/38207208?v=4&w=50&h=50&mask=circle)](https://github.com/tnsetting)[![34587798](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/34587798?v=4&w=50&h=50&mask=circle)](https://github.com/akhurana001)[![8200209](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/8200209?v=4&w=50&h=50&mask=circle)](https://github.com/catalinii)[![43587819](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/43587819?v=4&w=50&h=50&mask=circle)](https://github.com/chetcode)[![10869815](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/10869815?v=4&w=50&h=50&mask=circle)](https://github.com/jeevb)[![163899](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/163899?v=4&w=50&h=50&mask=circle)](https://github.com/regadas)[![1316881](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1316881?v=4&w=50&h=50&mask=circle)](https://github.com/akashkatipally)[![1360529](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1360529?v=4&w=50&h=50&mask=circle)](https://github.com/clairemcginty)[![653394](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/653394?v=4&w=50&h=50&mask=circle)](https://github.com/eapolinario)[![155087](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/155087?v=4&w=50&h=50&mask=circle)](https://github.com/derwiki)[![1399455](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1399455?v=4&w=50&h=50&mask=circle)](https://github.com/th0114nd)[![1810591](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1810591?v=4&w=50&h=50&mask=circle)](https://github.com/asottile)[![80421934](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/80421934?v=4&w=50&h=50&mask=circle)](https://github.com/SandraGH5)[![12219405](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/12219405?v=4&w=50&h=50&mask=circle)](https://github.com/fediazgon)[![3939659](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/3939659?v=4&w=50&h=50&mask=circle)](https://github.com/sbrunk)[![6774758](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/6774758?v=4&w=50&h=50&mask=circle)](https://github.com/ddhirajkumar)[![18337807](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/18337807?v=4&w=50&h=50&mask=circle)](https://github.com/max-hoffman)[![54334265](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/54334265?v=4&w=50&h=50&mask=circle)](https://github.com/michaels-lyft)[![9609986](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/9609986?v=4&w=50&h=50&mask=circle)](https://github.com/sonjaer)[![322624](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/322624?v=4&w=50&h=50&mask=circle)](https://github.com/AdrianoKF)[![54333860](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/54333860?v=4&w=50&h=50&mask=circle)](https://github.com/aalavian)[![7005765](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/7005765?v=4&w=50&h=50&mask=circle)](https://github.com/convexquad)[![48966647](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/48966647?v=4&w=50&h=50&mask=circle)](https://github.com/asahalyft)[![77167782](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/77167782?v=4&w=50&h=50&mask=circle)](https://github.com/apatel-fn)[![31381038](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/31381038?v=4&w=50&h=50&mask=circle)](https://github.com/lordnodd)[![7358951](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/7358951?v=4&w=50&h=50&mask=circle)](https://github.com/frsann)[![6984748](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/6984748?v=4&w=50&h=50&mask=circle)](https://github.com/jbrambleDC)[![125105](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/125105?v=4&w=50&h=50&mask=circle)](https://github.com/tekumara)[![4308533](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/4308533?v=4&w=50&h=50&mask=circle)](https://github.com/rubenbarragan)[![11269256](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/11269256?v=4&w=50&h=50&mask=circle)](https://github.com/sushrut111)[![61228633](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/61228633?v=4&w=50&h=50&mask=circle)](https://github.com/Tat-V)[![8817639](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/8817639?v=4&w=50&h=50&mask=circle)](https://github.com/ThomVett)[![17309187](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/17309187?v=4&w=50&h=50&mask=circle)](https://github.com/datability-io)[![57967031](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/57967031?v=4&w=50&h=50&mask=circle)](https://github.com/varshaparthay)[![1778407](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1778407?v=4&w=50&h=50&mask=circle)](https://github.com/ybubnov)[![12450632](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/12450632?v=4&w=50&h=50&mask=circle)](https://github.com/ajsalow)[![49699333](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/in/29110?v=4&w=50&h=50&mask=circle)](https://github.com/apps/dependabot)[![65977800](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/65977800?v=4&w=50&h=50&mask=circle)](https://github.com/service-github-lyft-semantic-release)[![85753828](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/85753828?v=4&w=50&h=50&mask=circle)](https://github.com/csirius)[![26953709](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/26953709?v=4&w=50&h=50&mask=circle)](https://github.com/Pianist038801)[![84735036](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/84735036?v=4&w=50&h=50&mask=circle)](https://github.com/jsonporter)[![1388071](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1388071?v=4&w=50&h=50&mask=circle)](https://github.com/aviaviavi)[![58770001](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/58770001?v=4&w=50&h=50&mask=circle)](https://github.com/Professional0321)[![19733683](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/19733683?v=4&w=50&h=50&mask=circle)](https://github.com/snyk-bot)[![18363301](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/18363301?v=4&w=50&h=50&mask=circle)](https://github.com/jimbobby5)[![4023015](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/4023015?v=4&w=50&h=50&mask=circle)](https://github.com/pradithya)[![67166843](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/67166843?v=4&w=50&h=50&mask=circle)](https://github.com/vvasavada-fn)[![25364490](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/25364490?v=4&w=50&h=50&mask=circle)](https://github.com/haoyuez)[![7515359](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/7515359?v=4&w=50&h=50&mask=circle)](https://github.com/narape)[![200401](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/200401?v=4&w=50&h=50&mask=circle)](https://github.com/arturdryomov)[![24739949](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/24739949?v=4&w=50&h=50&mask=circle)](https://github.com/felixwang9817)[![10430635](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/10430635?v=4&w=50&h=50&mask=circle)](https://github.com/juandiegopalomino)[![405480](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/405480?v=4&w=50&h=50&mask=circle)](https://github.com/georgesnelling)[![1004789](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1004789?v=4&w=50&h=50&mask=circle)](https://github.com/dschaller)[![1335881](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1335881?v=4&w=50&h=50&mask=circle)](https://github.com/hoyajigi)[![4830700](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/4830700?v=4&w=50&h=50&mask=circle)](https://github.com/NitinAgg)[![69161722](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/69161722?v=4&w=50&h=50&mask=circle)](https://github.com/noobkid2411)[![6506810](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/6506810?v=4&w=50&h=50&mask=circle)](https://github.com/stephen37)[![14992189](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/14992189?v=4&w=50&h=50&mask=circle)](https://github.com/eanakhl)[![1175392](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1175392?v=4&w=50&h=50&mask=circle)](https://github.com/adinin)[![7475946](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/7475946?v=4&w=50&h=50&mask=circle)](https://github.com/anton-malakhov)[![304786](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/304786?v=4&w=50&h=50&mask=circle)](https://github.com/kinow)[![24402505](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/24402505?v=4&w=50&h=50&mask=circle)](https://github.com/Daeruin)[![86911142](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/86911142?v=4&w=50&h=50&mask=circle)](https://github.com/idivyanshbansal)[![10345184](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/10345184?v=4&w=50&h=50&mask=circle)](https://github.com/hasukmistry)[![16461847](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/16461847?v=4&w=50&h=50&mask=circle)](https://github.com/JakeNeyer)[![2538760](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/2538760?v=4&w=50&h=50&mask=circle)](https://github.com/akumor)[![14008978](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/14008978?v=4&w=50&h=50&mask=circle)](https://github.com/jeremydonahue)[![1633460](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1633460?v=4&w=50&h=50&mask=circle)](https://github.com/jmcarp)[![19229049](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/19229049?v=4&w=50&h=50&mask=circle)](https://github.com/lsena)[![7548823](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/7548823?v=4&w=50&h=50&mask=circle)](https://github.com/Dread1982)[![580328](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/580328?v=4&w=50&h=50&mask=circle)](https://github.com/ilikedata)[![1027207](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1027207?v=4&w=50&h=50&mask=circle)](https://github.com/orf)[![1568889](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1568889?v=4&w=50&h=50&mask=circle)](https://github.com/leorleor)[![937967](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/937967?v=4&w=50&h=50&mask=circle)](https://github.com/moose007)[![25391173](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/25391173?v=4&w=50&h=50&mask=circle)](https://github.com/nicklofaso)[![14996868](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/14996868?v=4&w=50&h=50&mask=circle)](https://github.com/v01dXYZ)[![1043051](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1043051?v=4&w=50&h=50&mask=circle)](https://github.com/kylewaynebenson)[![1174730](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1174730?v=4&w=50&h=50&mask=circle)](https://github.com/mouuff)[![697033](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/697033?v=4&w=50&h=50&mask=circle)](https://github.com/vglocus)[![5732047](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/5732047?v=4&w=50&h=50&mask=circle)](https://github.com/stormy-ua)[![71284190](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/71284190?v=4&w=50&h=50&mask=circle)](https://github.com/gdungca-fn)[![26265392](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/26265392?v=4&w=50&h=50&mask=circle)](https://github.com/ttanay)[![3275593](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/3275593?v=4&w=50&h=50&mask=circle)](https://github.com/pradyunsg)[![66853113](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/in/68672?v=4&w=50&h=50&mask=circle)](https://github.com/apps/pre-commit-ci)[![107893](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/107893?v=4&w=50&h=50&mask=circle)](https://github.com/kmike)[![1300022](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1300022?v=4&w=50&h=50&mask=circle)](https://github.com/sirosen)[![1324225](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1324225?v=4&w=50&h=50&mask=circle)](https://github.com/hugovk)[![467294](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/467294?v=4&w=50&h=50&mask=circle)](https://github.com/bastimeyer)[![20280470](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/20280470?v=4&w=50&h=50&mask=circle)](https://github.com/drewyh)[![199429](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/199429?v=4&w=50&h=50&mask=circle)](https://github.com/dvarrazzo)[![1032633](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1032633?v=4&w=50&h=50&mask=circle)](https://github.com/dbitouze)[![18519037](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/18519037?v=4&w=50&h=50&mask=circle)](https://github.com/sethmlarson)[![11478411](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/11478411?v=4&w=50&h=50&mask=circle)](https://github.com/stonecharioteer)[![86675](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/86675?v=4&w=50&h=50&mask=circle)](https://github.com/estan)[![4748863](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/4748863?v=4&w=50&h=50&mask=circle)](https://github.com/pseudomuto)[![181308](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/181308?v=4&w=50&h=50&mask=circle)](https://github.com/htdvisser)[![1390277](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1390277?v=4&w=50&h=50&mask=circle)](https://github.com/jacobtolar)[![1391982](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1391982?v=4&w=50&h=50&mask=circle)](https://github.com/ezimanyi)[![3880001](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/3880001?v=4&w=50&h=50&mask=circle)](https://github.com/lpabon)[![770392](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/770392?v=4&w=50&h=50&mask=circle)](https://github.com/ArcEye)[![6178510](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/6178510?v=4&w=50&h=50&mask=circle)](https://github.com/mingrammer)[![5111931](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/5111931?v=4&w=50&h=50&mask=circle)](https://github.com/aschrijver)[![873434](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/873434?v=4&w=50&h=50&mask=circle)](https://github.com/panzerfahrer)[![20242](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/20242?v=4&w=50&h=50&mask=circle)](https://github.com/masterzen)[![16724](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/16724?v=4&w=50&h=50&mask=circle)](https://github.com/glasser)[![17330872](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/17330872?v=4&w=50&h=50&mask=circle)](https://github.com/murph0)[![419419](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/419419?v=4&w=50&h=50&mask=circle)](https://github.com/zetaron)[![1014](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/1014?v=4&w=50&h=50&mask=circle)](https://github.com/sunfmin)[![504507](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/504507?v=4&w=50&h=50&mask=circle)](https://github.com/guozheng)[![8542033](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/8542033?v=4&w=50&h=50&mask=circle)](https://github.com/nagytech)[![8841470](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/8841470?v=4&w=50&h=50&mask=circle)](https://github.com/suusan2go)[![901479](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/901479?v=4&w=50&h=50&mask=circle)](https://github.com/mhaberler)[![6400253](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/6400253?v=4&w=50&h=50&mask=circle)](https://github.com/s4ichi)[![353644](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/353644?v=4&w=50&h=50&mask=circle)](https://github.com/dreampuf)[![12421077](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/12421077?v=4&w=50&h=50&mask=circle)](https://github.com/UnicodingUnicorn)[![809865](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/809865?v=4&w=50&h=50&mask=circle)](https://github.com/philiptzou)[![19378](https://images.weserv.nl/?url=https://avatars.githubusercontent.com/u/19378?v=4&w=50&h=50&mask=circle)](https://github.com/timabell)
<!-- CONTRIBUTORS END -->
