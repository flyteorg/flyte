# RFC Process

[![hackmd-github-sync-badge](https://hackmd.io/0F2E71vhTk-zPE5qAmBFZA/badge)](https://hackmd.io/0F2E71vhTk-zPE5qAmBFZA)

An RFC - Request For Change - is a document issued mainly to recommend new features to the language and standard libraries. It provides the opportunity for all stakeholders to review and approve the changes before they are implemented. 

Not all changes will require an RFC. Bug fixes and documentation improvements can be carried out using GitHub pull requests (PRs).

Other more extensive changes need to be carefully tracked and reviewed, which is where RFCs come in.

## When you need to follow this process
[When you need to follow this process]: #when-you-need-to-follow-this-process

Significant changes to the system or changes to the user experience would require an RFC. Examples of changes can be:

  - Semantics or syntax changes, other than bugfixes (e.g. introducing a new Flytekit language feature)
  - Core-Language updates (e.g. FlyteIdl changes that introuce new wire-format)
  - Documentation restructuring
  - Service-2-Service changes (e.g. new gRPC APIs)

Changes not requiring RFCs include:

  - Documentation rephrasing or reorganizing
  - Addition of images, tables, etc. to supplement documentation
  - Updates to code samples, links, examples, etc.
  - Bug fixes that do not introduce changes listed above (use your judgement).

New feature implementation will almost certainly require an RFC. PRs attempting to do so will not be approved without a merged RFC.

## Before creating an RFC
[Before creating an RFC]: #before-creating-an-rfc

Some good practices when starting an RFC include:
- Gathering feedback from team or community members first, to confirm that the changes will indeed be useful.
- Discussing the topic on [Slack](http://slack.flyte.org/) or [GitHub Discussions](https://github.com/flyteorg/flyte/discussions/categories/) to gauge interest
- Making sure the changes align with short-term roadmaps (Brigning up the change in the Flyte Sync-Up meeting).
- Taking the time to produce a well-written, well-thought-of document. A suggested RFC template can be found [here](https://github.com/flyteorg/flyte/blob/RFC-Process/rfc/RFC-0000-Template.md).

## The process

[![](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggVERcbiAgICBBW1N0YXJ0XSAtLT58V3JpdGUgUkZDfCBCKE9wZW4gYSBQUilcbiAgICBCIC0tPiBDe1BSIFJldmlld31cbiAgICBDIC0tPnxBcHByb3ZlZHwgRFtNZXJnZV1cbiAgICBDIC0tPnxGZWVkYmFja3wgQ1xuICAgIEMgLS0-fFJlamVjdGVkfCBGW1BSIENsb3NlZF0iLCJtZXJtYWlkIjp7InRoZW1lIjoiZGVmYXVsdCJ9LCJ1cGRhdGVFZGl0b3IiOmZhbHNlLCJhdXRvU3luYyI6dHJ1ZSwidXBkYXRlRGlhZ3JhbSI6ZmFsc2V9)](https://mermaid-js.github.io/mermaid-live-editor/edit##eyJjb2RlIjoiZ3JhcGggVERcbiAgICBBW1N0YXJ0XSAtLT58V3JpdGUgUkZDfCBCKE9wZW4gYSBQUilcbiAgICBCIC0tPiBDe1BSIFJldmlld31cbiAgICBDIC0tPnxBcHByb3ZlZHwgRFtNZXJnZV1cbiAgICBDIC0tPnxGZWVkYmFja3wgQ1xuICAgIEMgLS0-fFJlamVjdGVkfCBGW1BSIENsb3NlXSIsIm1lcm1haWQiOiJ7XG4gIFwidGhlbWVcIjogXCJkZWZhdWx0XCJcbn0iLCJ1cGRhdGVFZGl0b3IiOmZhbHNlLCJhdXRvU3luYyI6dHJ1ZSwidXBkYXRlRGlhZ3JhbSI6ZmFsc2V9)

### The bare-minimum

RFCs must be introduced as a `.md` (Markdown-format) file in a PR into the flyte repo. Details of the location to store them can be [found below](#Where-to-store-RFCs).

### Opening a PullRequest

* Fork the [flyte repository](https://github.com/flyteorg/flyte).
* Copy `rfc/0000-template.md` to `rfc/<category>/0000-my-feature.md` (where "my-feature" is descriptive). Don't assign an RFC number yet; This is going to be the PR number and we'll rename the file accordingly if the RFC is accepted.
* Fill in the RFC. Put care into the details.
* Submit a pull request. As a pull request the RFC will receive design feedback from the larger community, and the author should be prepared to revise it in response.
* Now that your RFC has an open pull request, use the issue number of the PR to update your 0000- prefix to that number.

### The recommended path

At Flyte, we have been using [HackMD](https://hackmd.io) to author RFCs.

#### Why HackMD?

It is recommended that the initial review of an RFC takes place on HackMD, before pushing to GitHub. Several benefits of this include:
- PR authors typically have a GitHub account, which is all that's needed to sign in to HackMD. 
- Ease of commenting, iterating and adjusting on HackMD before pushing to GutHub, as opposed to requesting a PR for every change made to the RFC on GitHub. Visit https://hackmd.io/s/how-to-use-comments for more information on commenting.
- Ease of document tracking, compared to Google docs, since RFCs will be stored in GitHub making the RFCs indexed, searchable and versioned.

#### The HackMD process

The Flyte repo on GitHub contains an RFC folder where RFCs are to be pushed to. 

For those new to HackMD, here is how to begin:
- Visit https://hackmd.io/
- Click `Sign in` and use GitHub to login
- Click on `+ New Note`
- Fire away your RFC! 
- Make sure the link to the hackmd document is included within the body of the document.

Note: HackMD documents are written in Markdown. RFC templates can be found [here](https://github.com/flyteorg/flyte/blob/RFC-Process/rfc/RFC-0000-Template.md). 

After completing your RFC on HackMD, it is time to push to GitHub:
- On the top right, locate the ellipsis (...) next to your avatar
- Click on "Versions and GitHub Sync"
- On the top of the pop-out menu, select "Push to GitHub"
- A new pop-out menu will appear. Make the following selections:
    |         Option             | Value                                  |
    | -------------------------- | -------------------------------------- |
    | Select Repo:               | <your fork>                         |
    | Select branch:             | <Insert a new or existing branch name> |
    | Select file:               | rfcs/<category>/awesomefeature.md      |
    | Choose version(s) to push: | Fill in a descriptive name and a self-link to the hackmd note.  |
- Go ahead and Push to GitHub

## Where to store RFCs

The Flyte repo on GitHub has an RFC folder with 3 directories:
- Core language: Changes to FlyteIdl that change the wire-format in any way are considered significant changes that require revision and approval.
  
  Reviewers: At least one of [flyte core maintainers](https://github.com/orgs/flyteorg/teams/flyte-core-maintainers) and one of [flyte maintainers](https://github.com/orgs/flyteorg/teams/flyte-maintainers/members).
- General System: Changes to other repos that introduce signficant change of behavior or user-impacting features.
  
  Reviewer: At least one of [flyte maintainers](https://github.com/orgs/flyteorg/teams/flyte-maintainers/members).
- CI-CD: Significant changes to CI-CD System that have impact across different repositories.
  
  Reviewer: At least one of [flyte maintainers](https://github.com/orgs/flyteorg/teams/flyte-maintainers/members).
  
## Reviewing RFCs

Once the RFC is created in a PR, reviewers can take up to 3 weeks to review, test, and either approve or recommend adjustments. They may call for meetings, demos, or other measures deemed necessary.

### Instructions for Reviewers

1. If you are set as a reviewer for a PR, you must submit a review within 3 weeks of the PR open date.
1. We are not looking for consensus-based decision making. Strong objections to proposals must be considered and a decision one way or another should be taken.
1. Being curtious and responsive to comments and responses help contributors stay engaged and smoothen out the overall process.
1. If there is a link to HackMD in the PR, please use that to leave comments and have discussions on the RFC. If there isn't one, feel free to request one or submit your comments in the GitHub PR directly.

Once approved, the RFC will be pushed to GitHub to be implemented.

RFC PRs whose authors become unresponsive to reviewers' comments within 5 business days will be closed.

## Implementing RFCs

RFCs for vital features may need to be implemented immediately, whereas other RFCs may need to wait for the right developer to be available. 

The author of an RFC may not always be the one to implement it, but if they do, they are welcome to submit an implementation for review after the RFc has been accepted. 

## RFC Postponement

Rejected RFCs may be labelled as "postponed". This means the PR has been closed without implementation, because either:
- the proposed change will need to wait for prerequisite features, 
- the business model may not be presently able to support it,
- or other more urgent changes may need to be implemented first. 

Although "postponed" RFCs have already gone through the approval process, it is always recommended that they be reviewed again prior to implementation, to make sure the existing ecosystem will still benefit from the change.

Alternatively, if an RFC is not likely to be implemented, it should be labelled as "rejected" and permanently closed.

## Contributions
[Contributions]: #contributions

Any contribution to FlyteOrg repos will be licensed with [Apache-2.0](https://github.com/flyteorg/flyte/blob/master/LICENSE) license without any additional terms or conditions.

Adapted from: 
[Rust RFCs - Active RFC List](https://github.com/rust-lang/rfcs/blob/master/README.md) 