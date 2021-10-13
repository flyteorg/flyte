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
  <a href="https://forms.gle/UVuek9WfBoweiqcJA">
    <img src="https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social" alt="Slack Status" />
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

- [Slack](https://forms.gle/UVuek9WfBoweiqcJA)
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


<!-- NOTE: the contributors.yaml action fills in the section below with the contributors list -->
## Contributors

A big thank you to the community for making Flyte possible!

<!-- Flyte Contributors -->

<table>
<tr>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/kumare3>
            <img src=https://avatars.githubusercontent.com/u/16888709?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Ketan Umare/>
            <br />
            <sub style="font-size:10px"><b>Ketan Umare</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/EngHabu>
            <img src=https://avatars.githubusercontent.com/u/27159?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Haytham Abuelfutuh/>
            <br />
            <sub style="font-size:10px"><b>Haytham Abuelfutuh</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/wild-endeavor>
            <img src=https://avatars.githubusercontent.com/u/2896568?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Yee Hing Tong/>
            <br />
            <sub style="font-size:10px"><b>Yee Hing Tong</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/katrogan>
            <img src=https://avatars.githubusercontent.com/u/953358?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Katrina Rogan/>
            <br />
            <sub style="font-size:10px"><b>Katrina Rogan</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/SandraGH5>
            <img src=https://avatars.githubusercontent.com/u/80421934?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Sandra Youssef/>
            <br />
            <sub style="font-size:10px"><b>Sandra Youssef</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/evalsocket>
            <img src=https://avatars.githubusercontent.com/u/10830562?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Yuvraj />
            <br />
            <sub style="font-size:10px"><b>Yuvraj </b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/cosmicBboy>
            <img src=https://avatars.githubusercontent.com/u/2816689?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Niels Bantilan/>
            <br />
            <sub style="font-size:10px"><b>Niels Bantilan</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/samhita-alla>
            <img src=https://avatars.githubusercontent.com/u/27777173?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Samhita Alla/>
            <br />
            <sub style="font-size:10px"><b>Samhita Alla</b></sub>
        </a>
    </td>
</tr>
<tr>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/flyte-bot>
            <img src=https://avatars.githubusercontent.com/u/78108056?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Flyte Bot/>
            <br />
            <sub style="font-size:10px"><b>Flyte Bot</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/jeevb>
            <img src=https://avatars.githubusercontent.com/u/10869815?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Jeev B/>
            <br />
            <sub style="font-size:10px"><b>Jeev B</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/honnix>
            <img src=https://avatars.githubusercontent.com/u/158892?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Honnix/>
            <br />
            <sub style="font-size:10px"><b>Honnix</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/georgesnelling>
            <img src=https://avatars.githubusercontent.com/u/405480?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=George Snelling/>
            <br />
            <sub style="font-size:10px"><b>George Snelling</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/pmahindrakar-oss>
            <img src=https://avatars.githubusercontent.com/u/77798312?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=pmahindrakar-oss/>
            <br />
            <sub style="font-size:10px"><b>pmahindrakar-oss</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/sbrunk>
            <img src=https://avatars.githubusercontent.com/u/3939659?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Sören Brunk/>
            <br />
            <sub style="font-size:10px"><b>Sören Brunk</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/akhurana001>
            <img src=https://avatars.githubusercontent.com/u/34587798?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=akhurana001/>
            <br />
            <sub style="font-size:10px"><b>akhurana001</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/dschaller>
            <img src=https://avatars.githubusercontent.com/u/1004789?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Derek/>
            <br />
            <sub style="font-size:10px"><b>Derek</b></sub>
        </a>
    </td>
</tr>
<tr>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/jonathanburns>
            <img src=https://avatars.githubusercontent.com/u/3880645?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Johnny Burns/>
            <br />
            <sub style="font-size:10px"><b>Johnny Burns</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/pingsutw>
            <img src=https://avatars.githubusercontent.com/u/37936015?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Kevin Su/>
            <br />
            <sub style="font-size:10px"><b>Kevin Su</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/migueltol22>
            <img src=https://avatars.githubusercontent.com/u/19375241?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Miguel Toledo/>
            <br />
            <sub style="font-size:10px"><b>Miguel Toledo</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/schottra>
            <img src=https://avatars.githubusercontent.com/u/1815175?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Randy Schott/>
            <br />
            <sub style="font-size:10px"><b>Randy Schott</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/hamersaw>
            <img src=https://avatars.githubusercontent.com/u/8888115?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Dan Rammer/>
            <br />
            <sub style="font-size:10px"><b>Dan Rammer</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/hoyajigi>
            <img src=https://avatars.githubusercontent.com/u/1335881?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Hyunseok Cho/>
            <br />
            <sub style="font-size:10px"><b>Hyunseok Cho</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/NitinAgg>
            <img src=https://avatars.githubusercontent.com/u/4830700?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Nitin Aggarwal/>
            <br />
            <sub style="font-size:10px"><b>Nitin Aggarwal</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/noobkid2411>
            <img src=https://avatars.githubusercontent.com/u/69161722?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Sonakshi Mishra/>
            <br />
            <sub style="font-size:10px"><b>Sonakshi Mishra</b></sub>
        </a>
    </td>
</tr>
<tr>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/stephen37>
            <img src=https://avatars.githubusercontent.com/u/6506810?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Stephen Batifol/>
            <br />
            <sub style="font-size:10px"><b>Stephen Batifol</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/igorvalko>
            <img src=https://avatars.githubusercontent.com/u/1330233?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Igor Valko/>
            <br />
            <sub style="font-size:10px"><b>Igor Valko</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/mayitbeegh>
            <img src=https://avatars.githubusercontent.com/u/6239450?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Sean Lin/>
            <br />
            <sub style="font-size:10px"><b>Sean Lin</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/anandswaminathan>
            <img src=https://avatars.githubusercontent.com/u/18408237?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Anand Swaminathan/>
            <br />
            <sub style="font-size:10px"><b>Anand Swaminathan</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/eanakhl>
            <img src=https://avatars.githubusercontent.com/u/14992189?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Anastasia Khlebnikova/>
            <br />
            <sub style="font-size:10px"><b>Anastasia Khlebnikova</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/adinin>
            <img src=https://avatars.githubusercontent.com/u/1175392?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Andrei Dinin/>
            <br />
            <sub style="font-size:10px"><b>Andrei Dinin</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/anton-malakhov>
            <img src=https://avatars.githubusercontent.com/u/7475946?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Anton Malakhov/>
            <br />
            <sub style="font-size:10px"><b>Anton Malakhov</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/aviaviavi>
            <img src=https://avatars.githubusercontent.com/u/1388071?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Avi Press/>
            <br />
            <sub style="font-size:10px"><b>Avi Press</b></sub>
        </a>
    </td>
</tr>
<tr>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/kinow>
            <img src=https://avatars.githubusercontent.com/u/304786?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Bruno P. Kinoshita/>
            <br />
            <sub style="font-size:10px"><b>Bruno P. Kinoshita</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/bnsblue>
            <img src=https://avatars.githubusercontent.com/u/1518524?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Chang-Hong Hsu/>
            <br />
            <sub style="font-size:10px"><b>Chang-Hong Hsu</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/Daeruin>
            <img src=https://avatars.githubusercontent.com/u/24402505?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Daeruin/>
            <br />
            <sub style="font-size:10px"><b>Daeruin</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/kanterov>
            <img src=https://avatars.githubusercontent.com/u/467927?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Gleb Kanterov/>
            <br />
            <sub style="font-size:10px"><b>Gleb Kanterov</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/JakeNeyer>
            <img src=https://avatars.githubusercontent.com/u/16461847?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=JakeNeyer/>
            <br />
            <sub style="font-size:10px"><b>JakeNeyer</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/akumor>
            <img src=https://avatars.githubusercontent.com/u/2538760?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Jan Kumor/>
            <br />
            <sub style="font-size:10px"><b>Jan Kumor</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/jeremydonahue>
            <img src=https://avatars.githubusercontent.com/u/14008978?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Jeremy Donahue/>
            <br />
            <sub style="font-size:10px"><b>Jeremy Donahue</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/juandiegopalomino>
            <img src=https://avatars.githubusercontent.com/u/10430635?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Juan Diego Palomino/>
            <br />
            <sub style="font-size:10px"><b>Juan Diego Palomino</b></sub>
        </a>
    </td>
</tr>
<tr>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/lu4nm3>
            <img src=https://avatars.githubusercontent.com/u/3936213?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Luis Medina/>
            <br />
            <sub style="font-size:10px"><b>Luis Medina</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/lsena>
            <img src=https://avatars.githubusercontent.com/u/19229049?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Luis Sena/>
            <br />
            <sub style="font-size:10px"><b>Luis Sena</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/Dread1982>
            <img src=https://avatars.githubusercontent.com/u/7548823?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Manuel Rombach/>
            <br />
            <sub style="font-size:10px"><b>Manuel Rombach</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/narape>
            <img src=https://avatars.githubusercontent.com/u/7515359?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Nelson Arapé/>
            <br />
            <sub style="font-size:10px"><b>Nelson Arapé</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/rubenbarragan>
            <img src=https://avatars.githubusercontent.com/u/4308533?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Rubén Barragán/>
            <br />
            <sub style="font-size:10px"><b>Rubén Barragán</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/ilikedata>
            <img src=https://avatars.githubusercontent.com/u/580328?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Tai/>
            <br />
            <sub style="font-size:10px"><b>Tai</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/orf>
            <img src=https://avatars.githubusercontent.com/u/1027207?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Tom Forbes/>
            <br />
            <sub style="font-size:10px"><b>Tom Forbes</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/varshaparthay>
            <img src=https://avatars.githubusercontent.com/u/57967031?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Varsha Parthasarathy/>
            <br />
            <sub style="font-size:10px"><b>Varsha Parthasarathy</b></sub>
        </a>
    </td>
</tr>
<tr>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/ajsalow>
            <img src=https://avatars.githubusercontent.com/u/12450632?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=ajsalow/>
            <br />
            <sub style="font-size:10px"><b>ajsalow</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/pradithya>
            <img src=https://avatars.githubusercontent.com/u/4023015?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=aria/>
            <br />
            <sub style="font-size:10px"><b>aria</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/brucearctor>
            <img src=https://avatars.githubusercontent.com/u/5032356?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=brucearctor/>
            <br />
            <sub style="font-size:10px"><b>brucearctor</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/leorleor>
            <img src=https://avatars.githubusercontent.com/u/1568889?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=leorleor/>
            <br />
            <sub style="font-size:10px"><b>leorleor</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/moose007>
            <img src=https://avatars.githubusercontent.com/u/937967?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=mustaafa/>
            <br />
            <sub style="font-size:10px"><b>mustaafa</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/nicklofaso>
            <img src=https://avatars.githubusercontent.com/u/25391173?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=nicklofaso/>
            <br />
            <sub style="font-size:10px"><b>nicklofaso</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 135.0; height: 135.0">
        <a href=https://github.com/v01dXYZ>
            <img src=https://avatars.githubusercontent.com/u/14996868?v=4 width="90;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=v01dXYZ/>
            <br />
            <sub style="font-size:10px"><b>v01dXYZ</b></sub>
        </a>
    </td>
</tr>
</table>
