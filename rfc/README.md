# RFC Process

>*"To think, you have to write. If you're thinking without writing, you only think you're thinking"*.  Leslie Lamport

## Table of Contents
- [Introduction](#introduction)
- [When to Follow this Process](#when-to-follow-this-process)
- [Before Creating an RFC](#before-creating-an-rfc)
- [Where to Store RFCs](#where-to-store-rfcs)
- [The Process](#the-process)
  - [Opening a Pull Request](#opening-a-pull-request)
  - [Acceptance Criteria](#acceptance-criteria)
- [Implementing RFCs](#implementing-rfcs)
- [RFC Postponement](#rfc-postponement)
- [Contributions](#contributions)
- [Prior Art](#prior-art)

## Introduction

An RFC - Request For Comments - is a document issued mainly to recommend changes to the language and standard libraries. The RFC process aims to provide a consistent and controlled path for changes to enter the project, ensuring that all stakeholders can review and approve them before they are implemented.

## When to Follow this Process

You need to follow this process if you intend to make a significant change to the system or changes to the user experience. What constitutes a "significant" change is evolving based on community feedback, but may include the following:

- Semantics or syntax changes, other than bugfixes (e.g. introducing a new Flytekit language feature)
- Core-Language updates (e.g. FlyteIdl changes that introduce a new wire-format)
- Documentation restructuring
- Service-to-Service changes (e.g. new gRPC APIs)
- Breaking changes
- Governance changes

Changes not requiring RFCs include:

- Documentation rephrasing or reorganizing
- Addition of images, tables, etc. to supplement documentation
- Updates to code samples, links, examples, etc.
- Bug fixes that do not introduce changes listed above (use your judgment).

**NOTE:** New feature implementation will almost certainly require an RFC; PRs attempting to do so will not be approved without a merged RFC.

## Before Creating an RFC

Some good practices when starting an RFC include:

- Gathering feedback from team or community members first, to confirm that the changes will indeed be useful
- Starting a Discussion at the [RFC Incubator](https://github.com/flyteorg/flyte/discussions/new?category=rfc-incubator) to gauge interest. Once you have received positive feedback, especially from Maintainers or Steering Committee members, please proceed to:
  - Ask in the Discussion for confirmation to submit an RFC
  - If there's no objection (silence is approval) then create an Issue from the Discussion ([see how](https://docs.github.com/en/issues/tracking-your-work-with-issues/creating-an-issue#creating-an-issue-from-discussion))
  - Proceed to [open a PR](#opening-a-pull-request)
- Discussing the topic on the [#contribute](https://flyte-org.slack.com/archives/C04NJPLRWUX) Slack channel
- Adding the topic to the Contributor's [meeting agenda](https://hackmd.io/@davidmirror/rkqCpbK1n) to make sure it aligns with roadmap
- Taking the time to produce a well-written, well-thought-of document by using the template located [here](https://github.com/flyteorg/flyte/blob/RFC-Process/rfc/RFC-0000-Template.md).

## Where to Store RFCs

The Flyte repo on GitHub has an RFC folder with 3 directories:

- **Core language:** Proposals to `FlyteIdl` that change the wire-format in any way are considered significant changes that require revision and approval.
  
  Reviewers: At least one of [Flyte maintainers](https://github.com/flyteorg/community/blob/main/MAINTAINERS.md) and one of [Technical Steering Committee](https://github.com/flyteorg/community/blob/main/MAINTAINERS.md).
  
- **General System:** Changes to other repos that introduce a significant change of behavior or user-impacting features.
  
  Reviewers: At least one of [Flyte maintainers](https://github.com/flyteorg/community/blob/main/MAINTAINERS.md) and one of [Technical Steering Committee](https://github.com/flyteorg/community/blob/main/MAINTAINERS.md).
  
- **CI-CD:** Significant changes to CI-CD System that have an impact across different repositories.
  
  Reviewer: At least one of [Flyte maintainers](https://github.com/flyteorg/community/blob/main/MAINTAINERS.md).

## The Process

### Opening a Pull Request

- Fork the [flyte repository](https://github.com/flyteorg/flyte).
- Copy `rfc/RFC-0000-Template.md` to `rfc/<category>/0000-my-feature.md` (where "my-feature" is descriptive). Don't assign an RFC number yet; this is going to be the PR number and we'll rename the file accordingly if the RFC is accepted.
- Fill in the RFC. Put care into the details.
- Submit a pull request. As a pull request, the RFC will receive design feedback from the larger community, and the author should be prepared to revise it in response.
- Now that your RFC has an open pull request, use the pull request number (e.g. `123` from `https://github.com/flyteorg/flyte/pull/123`) to update your 0000- prefix to that number.

Once a pull request is opened, the RFC is now in development and the following will happen:

- It will be introduced in a future Contributor's meetup, happening every other week, except otherwise informed.
- The proposal will be discussed as much as possible in the RFC pull request directly. Any outside discussion will be summarized in the comment thread.
- When deemed "ready," a maintainer or TSC member will propose a "motion for Final Comment Period (FCP)" along with a disposition of the outcome (merge, close, or postpone). This step is taken when enough discussions of the tradeoffs have taken place and the community is in a position to make a decision. 
- The proposal enters FCP unless there's any objection (lazy consensus).
- The Final Comment Period will last 7 days. If there's no objection, the FCP can close early.
- If no substantial new arguments or ideas are raised, the FCP will follow the outcome decided.
- The RFC author could also withdraw the proposal if they decide to accept a competing proposal as a better alternative or if they deem the original idea as not applicable anymore. 

### Acceptance Criteria

A proposal is considered Accepted when it has:
- Completed the FCP with no significant objections
- Received an approval vote from a supermajority (2/3) of the [Technical Steering Committee](https://github.com/flyteorg/community/blob/main/MAINTAINERS.md)'s members

## Implementing RFCs

RFCs for vital features may need to be implemented immediately, whereas other RFCs may need to wait for the right developer to be available. The author of an RFC may not always be the one to implement it, but if they do, they are welcome to submit an implementation for review after the RFC has been accepted. 

## RFC Postponement

Rejected RFCs may be labeled as "postponed." This means the PR has been closed without implementation, because either:
- the proposed change will need to wait for prerequisite features, 
- the business model may not be presently able to support it,
- or other more urgent changes may need to be implemented first. 

Although "postponed" RFCs have already gone through the approval process, it is always recommended that they be reviewed again prior to implementation, to make sure the existing ecosystem will still benefit from the change.

Alternatively, if an RFC is not likely to be implemented, it should be labeled as "rejected" and permanently closed.

## Contributions

Any contribution to FlyteOrg repos will be licensed with [Apache-2.0](https://github.com/flyteorg/flyte/blob/master/LICENSE) license without any additional terms or conditions.

## Prior Art

- [Rust RFC process](https://github.com/rust-lang/rfcs) 
- [Python PEP process](https://peps.python.org/pep-0001/)
- [ASF community design](https://community.apache.org/committers/lazyConsensus.html)
