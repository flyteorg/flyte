# Flyte Versioning & Planning Updates

Hack MD: https://hackmd.io/ZwarNg5iSjeJz2l0YBKIZQ

## Background & Issues
This is the current [roadmap](https://github.com/flyteorg/flyte/blob/master/docs/community/roadmap.rst). The core Flyte maintainers team is proposing some updates to it, specifically to the Milestones and Releases section.

We are proposing:
* Changes to the pace of Flyte releases.
* Changes to the versioning scheme used across Flyte and its components
* Updates to the planning process. It has long since been our goal to make Flyte planning a community endeavor.

## Pace
The regular cadence of monthly releases made sense when Flyte was still an incubating project as APIs were still being upgraded and core experiences were being improved at a rapid pace. As a more stable graduated project, Flyte development needs to ensure stability and backwards compatibility while still delivering substantial feature improvements. We propose moving to quarterly releases, with the understanding that rather than being tied strictly to the calendar, we aim to have substantial features and improvements at each quarter. If features slated for a given release are delayed, then the release will be delayed as well. The increased time will give the Flyte development team more time to beta test each feature and release.


## Versioning Scheme
### Problems
The issue with the current versioning scheme in Flyte this proposal seeks to solve is a lack of clarity
* We don't have a well-defined one. We release Flyte once a month, but when that happens we release a subset of the individual components at [arbitrary versions](https://github.com/flyteorg/flyte/blob/2635cda85b24926632f0d7391975732ab221fe1b/charts/flyte/values.yaml#L19). Example of the most recent [Admin minor bump](https://github.com/flyteorg/flyteadmin/pull/410) and the most recent [Propeller minor bump](https://github.com/flyteorg/flytepropeller/pull/429).
* Because these aren't really connected, it raises questions like, "Which versions of Propeller work with which versions of Admin (which work with which versions of Console)? What about K8s versions?  And what about CRD versions?"
* Once a minor bump occurs, all the PRs just get patch versions. This can lead to a large amount of code changes accumulating in [patch releases](https://github.com/flyteorg/flytepropeller/compare/v1.1.16..v1.1.0), which is not really what patch releases were meant for.
* Not using patch versions as they were meant to means that we have only awkward ways of patching prior releases. For example [this release](https://github.com/flyteorg/flyte/commit/5570eff6bd636e07e40b22c79319e46f927519a3#diff-638f077159e277b61cb799423fd14a2c496b05799d05fb998823984bf979344aR19) uses Admin `v1.1.2`. If there were a security issue with that version, we'd have to release a `v1.1.2-hotfix` or something. The reason is because `v1.1.3` is already taken. This leads maintaining prior Flyte Helm charts more awkward as well.

### Version Numbering
At each quarterly release, major components of Flyte and the Flyte repository itself will be released with an incremented minor version number and the version number will be aligned across those components. The major version number will remain `1` for the foreseeable future. That is, if the current version of Flyte is `1.2.x`, the next quarter's release will be `1.3.0` for Flyte and the major components.

After each version is released, merges to master will be assigned beta releases of the next quarter's release. That is, if `flytepropeller` version `v1.2.0` was just released, the next merge to master will be tagged `v1.3.0b0`.

Not strictly forcing a time-constraint on the Flyte release cycle means that if a substantial number of changes is merged, perhaps due to a security issue or just a rapid pace of feature development, we can always bring up the timeline of the release.

### Components Included in Version Alignment
Not all of the Flyte repos will follow this naming semantic. We will only align these:

* Propeller
* Admin
* Console
* datacatalog
* flytectl
* flytesnacks
* Flytekit
* flytekit-java

The last two we are going to tie together for now, but realize that we may want to unpin in the future.

The above are the more user facing repos. The components that will not be tied to this scheme are:
* flyteidl
* flytestdlib
* flyteplugins
* flytecopilot


#### Helm Charts
The Helm charts should also follow the Flyte release version. Unlike the individual components however, the Helm version and the Flyte version should be identical down to the patch version. While Helm charts do have two separate versions, one for the chart and one for the app, according to their [guidelines](https://codefresh.io/docs/docs/new-helm/helm-best-practices/), "These are unrelated and can be bumped up in any manner that you see fit. You can sync them together or have them increase independently."


### Release Branches and Patching
After each minor release, a release branch will be created. There will be no alignment of patch versions across the components. That is, By the end of the `1.3.x` release cycle, `flyteadmin` may be on `1.3.8` and `flytepropeller` may be on `1.3.2`.

#### Patching older releases
When developing bug fixes, by default we will continue to develop off of master, which will not be the stable branch. After such bug fixes are merged, it will be the responsibility of the developer to ensure that the patches are also applied to prior releases. At the current time, we propose only supporting one release back. That is, if `flytepropeller` has a bug fix that results in `v1.3.0b0` that patch will be applied to the `v1.2.x` release, but not the `v1.1.x` release.

We also propose that beta patch versions be merged into the release branch when patching prior releases. For example, assuming no patches have yet to be made to the `v1.2.0` release, when porting a bug fix that resulted in `v1.3.0b0` onto the `release-v1.2` branch, the developer can first release `v1.2.1b0` for testing into `release-v1.2` before releasing the `v1.2.1` release (also to `release-v1.2` of course). Such beta releases should be made (always, when it's a big change, at the discretion of the developer????).

#### Why not develop on release branches?
A proposal was made that:
* Bug fixes should be made to the latest release branch when the big was discovered.
* Features should be developed against master.

We decided that this approach was less desirable because:
* Development now happens on two branches. This may be more confusing.
* The cost of forgetting to patch a prior release is lower than the cost of forgetting to merge a bug fix forward.
* Bug fixes are less likely to be made in ways that ongoing feature development prohibit.
* If we ever support multiple Flyte releases, developers will have to port both backwards (onto earlier release branches) and forwards (onto master).


### Beta Flyte releases
Whether or not a patch version of any of the Flyte components also creates a Flyte patch release shall be left to the discretion of the developer.

### Documentation Versioning
We also currently have an issue with our documentation versioning. While our readthedocs page does have versioning enabled and we publish the [docs version](https://github.com/flyteorg/flyte/blob/80c098f10334b1c916d1e4274ab9f204152d9d80/docs/conf.py#L33), all the [intersphinx mappings](https://github.com/flyteorg/flyte/blob/80c098f10334b1c916d1e4274ab9f204152d9d80/docs/conf.py#L219) just point to `latest`. Keep in mind that this mapping not only exists in this `flyte` repo, but also in all the other repos that that mapping points to. That is, to maintain an accurate mapping of different versions of documentation, we'll need to update the mapping in all the repos.

To remediate this, we propose the following:
* Documentation should be pinned only to Major.Minor on all the repos that have their versions "aligned".
    * This means that as we release patch versions of Admin, Propeller, etc., if we're on v1.1 for instance, as Admin code/auto-generated documentation changes, the v1.1 listing of readthedocs will automatically pick it up.
* Repos that are not aligned will just default to the "latest" documentation version.
    * Question: Will this affect navigation back though?  Like if we go from `flyte` (v1.1) to `flyteidl` (latest) and then back `flyte`, will navigation be reset to latest?

### Compatibility
For each flytekit version, we will also publish a minimum Helm version.


### Action Items
* GH action that does the `bump-version` needs to be updated:
    * Given the latest release, make the next version a beta release. So that if the most recent release is v1.3.0, the next merge to master should be v1.4.0b0. If the latest version is v1.2.0b5, then the next version should be v1.2.0b5.
    * Stop reading the #patch/minor/major hash tags when creating release versions.
    * When pushing to a `release-X.Y` branch, should we create a release automatically? If not, create a job that can create patch releases for these `release-X.Y` branches.
* Flyte repo level workflow that creates releases on all the repos listed in the section above.
    * This should also automatically create a `release-X.Y` branch for each repo, off of master.


## Planning
Planning is done after each minor release of Flyte, again aiming for a quarterly cadence. These will be publicly hosted meetings with the aim of drawing in additional community support and contributions.

Additionally, the Flyte team will also triage incoming issues on a weekly basis, and will internally check in on status every two weeks. Whether there's a benefit to the community from running these meetings publicly will probably depend on the level of public contributions.


