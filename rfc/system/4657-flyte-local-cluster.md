# [RFC Template] Flyte Local Cluster

**Authors:**

- [@pingsutw](https://github.com/pingsutw)
- [@Future-Outlier](https://github.com/Future-Outlier)

## 1 Executive Summary

Flyte can support sandbox mode and single binary mode now.
We want to support local cluster mode, which won't start kubernetes when using flyte.
This aims to simplify the setup and use of Flyte for users who prefer or require a more localized environment.

## 2 Motivation

Currently, lots of companies deploy flyte in sandbox mode, which needs to run thekubernetes cluster.
However, there are companies want to deploy flyte without using kubernetes, so we want to propose a solution to run all components in local cluster, previously in kubernetes cluster. 

## 3 Proposed Implementation

There 3 mainly component in the kubernetes cluster we need to change it in flyte.

* Change database from postegres SQL to sqlite
* Change file system from minio and stow to local file system
* Change runtime from Kubernetes to Process

*Consider:*

- *using diagrams to help illustrate your ideas.*
- *including code examples if you're proposing an interface or system contract.*
- *linking to project briefs or wireframes that are relevant.*

## 4 Metrics & Dashboards

*What are the main metrics we should be measuring? For example, when interacting with an external system, it might be the external system latency. When adding a new table, how fast would it fill up?*

## 5 Drawbacks

* There are other priorities to be completed first
* Add the cost of maintaining flyte

## 6 Alternatives

Use sandbox and single binary mode only.

## 7 Potential Impact and Dependencies

We need to pay attention on how to overwrite functions so that it can work properly as single binary mode.

## 8 Unresolved questions

* Should local cluster mode have flyte console?

## 9 Conclusion

* Flyte local cluster mode can help companies to deploy flyte without using kubernetes, which is far more easier than before. 

## 10 RFC Process Guide, remove this section when done

*By writing an RFC, you're giving insight to your team on the direction you're taking. There may not be a right or better decision in many cases, but we will likely learn from it. By authoring, you're making a decision on where you want us to go and are looking for feedback on this direction from your team members, but ultimately the decision is yours.*

This document is a:

- thinking exercise, prototype with words.
- historical record, its value may decrease over time.
- way to broadcast information.
- mechanism to build trust.
- tool to empower.
- communication channel.

This document is not:

- a request for permission.
- the most up to date representation of any process or system

**Checklist:**

- [ ]  Copy template
- [ ]  Draft RFC (think of it as a wireframe)
- [ ]  Share as WIP with folks you trust to gut-check
- [ ]  Send pull request when comfortable
- [ ]  Label accordingly
- [ ]  Assign reviewers
- [ ]  Merge PR

**Recommendations**

- Tag RFC title with [WIP] if you're still ironing out details.
- Tag RFC title with [Newbie] if you're trying out something experimental or you're not entirely convinced of what you're proposing.
- Tag RFC title with [RR] if you'd like to schedule a review request to discuss the RFC.
- If there are areas that you're not convinced on, tag people who you consider may know about this and ask for their input.
- If you have doubts, ask on [#feature-discussions](https://slack.com/app_redirect?channel=CPQ3ZFQ84&team=TN89P6GGK) for help moving something forward.
