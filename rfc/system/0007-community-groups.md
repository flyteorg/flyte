# Update Governance model

**Authors:**

- David Espejo (@davidmirror-ops)

## 1. Executive Summary

The goal of this RFC is to introduce the concept of "community groups" as a mechanism to foster more directed collaboration among community members. The proposal includes the lifecycle and roles for Special Interest Groups/Working Groups.

## 2. Motivation

 The Flyte project is made up of several components, distributed in different repos and with the potential to be improved independently by a specialized group of individuals. Also, the set of use cases covered by Flyte and its integrations creates an opportunity to gather together community members with similar interests to collaborate towards common goals. 
   
   This proposal outlines a governance structure and corresponding contribution model that encourages distributed decision-making and code ownership.

## 3. Proposed Implementation
![](https://i.imgur.com/d0B0q9U.png)


## SIGs

Special Interest Groups are formed by members from multiple companies and organizations, with a common goal of advancing the project with respect a specific topic. Every identifiable subpart of the project (e.g. repository, subdirectory, API, test, issue, PR) is intended to be owned by some SIG.

Areas covered by SIGs are defined in terms of the level of cross-collaboration and code ownership required to meet their goals:

* **Horizontal SIGs**: deal with project-wide priorities as defined by the Technical Steering Committee, potentially owning multiple technical assets (repositories, etc)
* **Domain-specific SIGs**: more focused on particular components of the project, including work towards developing or improving integrations with other projects.

### Organizational structure

Every SIG should have one SIG chair at any given time, appointing a co-chair when considered necessary. SIG chairs are intended to be organizers and facilitators, responsible for the operation of the SIG and for communication and coordination with the other SIGs, the Steering Committee, and the broader community.

Each SIG must have a charter that specifies its scope (topics, subsystems, code repos and directories), responsibilities and areas of authority.

SIGs are voluntary groups which operate in the open and anyone can join.

## Working Groups

Designed to facilitate discussions/work regarding topics that are short-lived or that span multiple SIGs, Working Groups are distinct from SIGs in that they are intended to:

- facilitate collaboration across SIGs
- facilitate an exploration of a problem / solution through a group with minimal governmental overhead

### It is a Working Group if...

- It has a clear goal measured through a specific deliverable or deliverables
- It is temporary in nature and will be disbanded after reaching its stated goal(s)

## Processes

### How to start and operate a SIG
- [ ] Define and submit a charter including:
  - [ ] Scope. It should be a short statement describing SIG's goal at a high level
  - [ ] List of technical assets owned by this SIG (code repos, etc)
  - [ ] Non-goals or Out of scope items
  - [ ] Roles
    - [ ] SIG Chair(s)
        - Should be at least "Community members" as defined in [GOVERNANCE](https://github.com/flyteorg/community/blob/main/GOVERNANCE.md#community-roles-and-path-to-maintainership)
        - Should define how priorities and commitments are managed
        - Leads taking an extended leave of 1 or more months SHOULD coordinate with other leads to ensure the role is adequately staffed during their leave
        - Leads MAY decide to step down at anytime and propose a replacement.
   - [ ] Communication channels
       - [ ] Meetings should occur at least once per month. Use [Doodle](https://doodle.com/en/) to coordinate if needed
           - Meetings should be recorded and notes public
       - [ ] Slack channel to use or create

- [ ] Submit a PR to [/community](https://github.com/flyteorg/community) with the charter and request reviews from [Technical Steering Committee](https://github.com/flyteorg/community/blob/main/MAINTAINERS.md) Members
-  A new SIG is approved by a supermajority of TSC active members

### How to start and operate a WG

- [ ] Define and submit a charter including:
    - [ ] The problem this WG is trying to solve
    - [ ] The artifact(s) it will deliver
    - [ ] The SIG(s) it will collaborate with
    - [ ] Repo ownership (if any)
- [ ] Roles
    - [ ] Lead(s)
        - [ ] Should be at least “Community member” as defined in [GOVERNANCE](https://github.com/flyteorg/community/blob/main/GOVERNANCE.md#community-roles-and-path-to-maintainership)
        - [ ] Must coordinate project updates to the wider community.
- [ ] Submit a PR to [/community](https://github.com/flyteorg/community) with the charter and request reviews from [Technical Steering Committee](https://github.com/flyteorg/community/blob/main/MAINTAINERS.md) Members
-  A new WG is approved by a supermajority of TSC active members

## Community groups retirement process

### SIGs
In the event that the SIG is unable to regularly establish consistent quorum or otherwise fulfill its responsibilities

- after 3 or more months it SHOULD be retired
- after 6 or more months it MUST be retired

### WGs

Working Groups will be disbanded if either of the following is true:

- Its original goal was accomplished and results were communicated
- There is no longer a Lead (with a 6 week grace period)
- None of the communication channels for the Working Group have been in use for the goals outlined at the founding of the Working Group (with a 3 month grace period)

## 4. Drawbacks

At the beginning, it could cause some confusion especially around SIG/WG ownership on repositories, but once those community groups are formed, eventually most (ideally all) subcomponents of the project will be owned by a group.

## 5. Alternatives

There are community groups with less overhead (like User Groups) but typically they don't have deliverables and little sense of ownership.

## 6. Potential Impact and Dependencies

SIG/WG membership is open to any community member. Participation in such groups, though, doesn't confer anyone a  path for increased privileges  different from the requirements already documented in  [GOVERNANCE](https://github.com/flyteorg/community/blob/main/GOVERNANCE.md#community-roles-and-path-to-maintainership). 


## 7. Prior art

This proposal takes inspiration from:

- [Kubernetes governance](https://github.com/kubernetes/community/blob/master/governance.md)
- [CNCF Contributor ladder](https://github.com/cncf/project-template/blob/main/CONTRIBUTOR_LADDER.md#contributor-ladder-template)


