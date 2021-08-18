# RFC Process

[![hackmd-github-sync-badge](https://hackmd.io/0F2E71vhTk-zPE5qAmBFZA/badge)](https://hackmd.io/0F2E71vhTk-zPE5qAmBFZA)

An RFC - Request For Change - is a document issued mainly to recommend new features to the language and standard libraries. It provides the opportunity for all stakeholders to review and approve the changes before they are implemented. 

Not all changes will require an RFC. Bug fixes and documentation improvements can be carried out using GitHub pull requests (PRs).

Other more extensive changes need to be carefully tracked and reviewed, which is where RFCs come in.


## Table of Contents
[Table of Contents]: #table-of-contents
[toc]

## When you need to follow this process
[When you need to follow this process]: #when-you-need-to-follow-this-process

Extensive changes that would benefit from an RFC will depend on the organization, community norms and expectations, and where the changes will be implemented. Such changes can be :

  - Semantics or syntax changes, other than bugfixes (e.g. introducing a new flytekit language feature)
  - Core-Language updates (e.g. FlyteIdl changes that introuce new wire-format)
  - Documentation restructuring
  - Service-2-Service changes (e.g. new gRPC APIs)

Changes not requiring RFCs include:

  - Documentation rephrasing or reorganizing
  - Addition of images, tables, etc. to supplement documentation
  - Updates to code samples, links, examples, etc.
  - Bug fixes

New feature implementation will almost certainly require an RFC. PRs attempting to do so may be closed with a request to submit an RFC instead. 

## Before creating an RFC
[Before creating an RFC]: #before-creating-an-rfc

Some good practices when starting an RFC include:
- Gathering feedback from team or community members first, to confirm that the changes will indeed be useful.
- Discussing the topic on [Slack](http://slack.flyte.org/) or [GitHub Discussions](https://github.com/flyteorg/flyte/discussions/categories/) to gauge interest
- Making sure the changes align with short-term roadmaps (Brigning up the change in the Flyte Sync-Up meeting).
- Taking the time to produce a well-written, well-thought-of document. A suggested RFC template can be found [here](insert link to RFC template) 

Generally speaking, proposals met with encouraging feedback from long-standing project developers and community members have a better chance of their RFC being approved.

At Flyte, we encourage RFCs to be initiated on HackMD (https://hackmd.io/), where they can be iterated on and when ready, pushed to GitHub as a PR, then approved for implementation. Details of the process can be found below. 

## Why HackMD?
[Why HackMD?]: #why-hackmd?

It is recommended that the initial review of an RFC takes place on HackMD, before pushing to GitHub. Several benefits of this include:
- Ease of commenting, iterating and adjusting on HackMD before pushing to GutHub, as opposed to requesting a PR for every change made to the RFC on GitHub. Visit https://hackmd.io/s/how-to-use-comments for more information on commenting.
- Ease of document tracking, compared to Google docs, since RFCs will be linked to github making the RFCs indexed, searchable and versioned.

## The process
[The process]: #the-process

The Flyte repo on GitHub contains an RFC folder where RFCs are to be pushed to. 

For those new to HackMD, here is how to begin:
- Visit https://hackmd.io/
- Enter your email (preferably one associated with GitHub) and sign up
- On the top left, locate "My Workspace". Use the drop-down arrow to select "Flyte"
- Click on "+ New Team Note"
- Fire away your RFC! 

Note: HackMD documents are written in Markdown. RFC templates can be found here (insert link to the GitHub template). 

After completing your RFC on HackMD, it is time to push to GitHub:
- On the top right, locate the ellipsis (...) next to your avatar
- Click on "Versions and GitHub Sync"
- On the top of the pop-out menu, select "Push to GitHub"
- A new pop-out menu will appear. Make the following selections:
    |         Option             | Value                                  |
    | -------------------------- | -------------------------------------- |
    | Select Repo:               | flyteorg/flyte                         |
    | Select branch:             | <Insert a new or existing branch name> |
    | Select file:               | rfcs/<category>/awesomefature.md       |
    | Choose version(s) to push: | Fill in a descriptive name and a self-link to the hackmd note.  |
- Go ahead and Push to GitHub

(replace above starred points with one screenshot featuring correct selections)

The Flyte repo on GitHub has an RFC folder with 3 directories:
- Core language: Changes to FlyteIdl that change the wire-format in any way are considered significant changes that require revision and approval.

  Lead reviewers: At least one of [flyte core maintainers](https://github.com/orgs/flyteorg/teams/flyte-core-maintainers) and one of [flyte maintainers](https://github.com/orgs/flyteorg/teams/flyte-maintainers/members).
- General System: Changes to other repos that introduce signficant change of behavior or user-impacting features.

  Lead reviewer: At least one of [flyte maintainers](https://github.com/orgs/flyteorg/teams/flyte-maintainers/members).
- CI-CD: Significant changes to CI-CD System that have impact across different repositories.

  Lead reviewer: At least one of [flyte maintainers](https://github.com/orgs/flyteorg/teams/flyte-maintainers/members).

## Reviewing RFCs
[Reviewing RFCs]: #reviewing-rfcs

Once the RFC is created in a PR, reviewers can take up to three weeks to review, test, and either approve or recommend adjustments. They may call for meetings, demos, or other measures deemed necessary. 

Once approved, the RFC will be pushed to GitHub to be implementated.

RFC PRs whose authors become unresponsive to reviewers comments will be closed.

## Implementing RFCs
[Implementing RFCs]: #implementing-rfcs

RFCs for vital features may need to be implemented immediately, whereas other RFCs may need to wait for the right developer to be available. 

The author of an RFC may not always be the one to implement it, but if they do, they are welcome to submit an implementation for review after the RFc has been accepted. 

## RFC Postponement
[RFC Postponement]: #rfc-postponement

Rejected RFCs may be labelled as "postponed". This means the PR has been closed without implementation, because either:
- the proposed change will need to wait for prerequisite features, 
- the business model may not be presently able to support it,
- or other more urgent changes may need to be implemented first. 

Although "postponed" RFCs have already gone through the approval process, it is always recommended that they be reviewed again prior to implementation, to make sure the existing ecosystem will still benefit from the change.

Alternatively, if an RFC is not likely to be implemented, it should be labelled as "rejected" and permanently closed.

## Contributions
[Contributions]: #contributions

Unless you explicitly state otherwise, any contribution intentionally submitted for inclusion in the work by you, as defined in the Apache-2.0 license, shall be dual licensed as above, without any additional terms or conditions.

(do we need to include some sort of a liability or disclaimer statement like this?)

Adapted from: 
[Rust RFCs - Active RFC List](https://github.com/rust-lang/rfcs/blob/master/README.md) 