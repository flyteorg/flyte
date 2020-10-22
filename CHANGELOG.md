# Changelog

All notable changes to this project will be documented in this file. See
[Conventional Commits](https://conventionalcommits.org) for commit guidelines.

## [0.15.1](http://github.com/lyft/flyteconsole/compare/v0.15.0...v0.15.1) (2020-10-22)


### Bug Fixes

* don't skip legacy children check when isParentNode=false ([#111](http://github.com/lyft/flyteconsole/issues/111)) ([ed6ede3](http://github.com/lyft/flyteconsole/commit/ed6ede389815cd821ec5fe12e628eea1be80cbac))

# [0.15.0](http://github.com/lyft/flyteconsole/compare/v0.14.0...v0.15.0) (2020-10-21)


### Features

* render CronSchedule ([#107](http://github.com/lyft/flyteconsole/issues/107)) ([c3c5d5d](http://github.com/lyft/flyteconsole/commit/c3c5d5d347fb1921af3cc7311deecda0b0c72295))

# [0.14.0](http://github.com/lyft/flyteconsole/compare/v0.13.2...v0.14.0) (2020-10-20)


### Features

* Fetch Node Executions directly if possible ([#109](http://github.com/lyft/flyteconsole/issues/109)) ([c740b83](http://github.com/lyft/flyteconsole/commit/c740b83c08b0d9da94c3b54d88481fb5817cd71c))

## [0.13.2](http://github.com/lyft/flyteconsole/compare/v0.13.1...v0.13.2) (2020-10-16)


### Bug Fixes

* race condition with pagination token ([#105](http://github.com/lyft/flyteconsole/issues/105)) ([a5a9240](http://github.com/lyft/flyteconsole/commit/a5a92406567160dfb07238d30472786c15f34cb4))

## [0.13.1](http://github.com/lyft/flyteconsole/compare/v0.13.0...v0.13.1) (2020-10-13)


### Bug Fixes

* use correct filter for single task executions ([#104](http://github.com/lyft/flyteconsole/issues/104)) ([8529858](http://github.com/lyft/flyteconsole/commit/8529858abf5e67fc6b278bfd1209822671df8c11))

# [0.13.0](http://github.com/lyft/flyteconsole/compare/v0.12.1...v0.13.0) (2020-10-08)


### Features

* Improve UX for Single Task Executions ([#103](http://github.com/lyft/flyteconsole/issues/103)) ([d0335dc](http://github.com/lyft/flyteconsole/commit/d0335dc1f86b98b30011e63d686b8168262d3bbd)), closes [#96](http://github.com/lyft/flyteconsole/issues/96) [#97](http://github.com/lyft/flyteconsole/issues/97) [#99](http://github.com/lyft/flyteconsole/issues/99) [#100](http://github.com/lyft/flyteconsole/issues/100) [#101](http://github.com/lyft/flyteconsole/issues/101) [#102](http://github.com/lyft/flyteconsole/issues/102)

## [0.12.1](http://github.com/lyft/flyteconsole/compare/v0.12.0...v0.12.1) (2020-09-25)


### Bug Fixes

* Render timestamp of protobuf in UTC ([#98](http://github.com/lyft/flyteconsole/issues/98)) ([37604a0](http://github.com/lyft/flyteconsole/commit/37604a07f434878cb4177699be27b41923a3af00))

# [0.12.0](http://github.com/lyft/flyteconsole/compare/v0.11.0...v0.12.0) (2020-09-22)


### Features

* get full inputs/outputs from execution data ([#92](http://github.com/lyft/flyteconsole/issues/92)) ([78b250d](http://github.com/lyft/flyteconsole/commit/78b250d665cbd5aa95bae995da40d364ef96a509))

# [0.11.0](http://github.com/lyft/flyteconsole/compare/v0.10.0...v0.11.0) (2020-09-02)


### Features

* Allow cloning of running executions ([#93](http://github.com/lyft/flyteconsole/issues/93)) ([dfa7e26](http://github.com/lyft/flyteconsole/commit/dfa7e260372e7743efcc42bb5c9f03742e776fa8))

# [0.10.0](http://github.com/lyft/flyteconsole/compare/v0.9.0...v0.10.0) (2020-08-31)


### Features

* Expose caching status for NodeExecutions ([#91](http://github.com/lyft/flyteconsole/issues/91)) ([cb200bd](http://github.com/lyft/flyteconsole/commit/cb200bdd4d2708315a556e1e7d8b3d1df6af029b))

# [0.9.0](http://github.com/lyft/flyteconsole/compare/v0.8.1...v0.9.0) (2020-08-24)


### Features

* Authentication can be optionally disabled ([#89](http://github.com/lyft/flyteconsole/issues/89)) ([d92a127](http://github.com/lyft/flyteconsole/commit/d92a1277af27a31863103eeda231d71b4e037151))

## [0.8.1](http://github.com/lyft/flyteconsole/compare/v0.8.0...v0.8.1) (2020-08-13)


### Bug Fixes

* only use NE filter for Nodes tab ([#88](http://github.com/lyft/flyteconsole/issues/88)) ([f4ba80b](http://github.com/lyft/flyteconsole/commit/f4ba80b0498a0b37dc73ed3c3c8e72844101dcb5))

# [0.8.0](http://github.com/lyft/flyteconsole/compare/v0.7.3...v0.8.0) (2020-07-20)


### Features

* 0.8.0 release ([#87](http://github.com/lyft/flyteconsole/issues/87)) ([65f2a81](http://github.com/lyft/flyteconsole/commit/65f2a8102a5cdbf2456868502dc3e7628cfc1433)), closes [#81](http://github.com/lyft/flyteconsole/issues/81) [#82](http://github.com/lyft/flyteconsole/issues/82) [#84](http://github.com/lyft/flyteconsole/issues/84) [#86](http://github.com/lyft/flyteconsole/issues/86)

## [0.7.3](http://github.com/lyft/flyteconsole/compare/v0.7.2...v0.7.3) (2020-07-06)


### Bug Fixes

* prevents collapsing of execution errors when scrolled out of view ([#80](http://github.com/lyft/flyteconsole/issues/80)) ([ba39c1c](http://github.com/lyft/flyteconsole/commit/ba39c1c9c0783d69aee48de2f1b666879472e6b8))

## [0.7.2](http://github.com/lyft/flyteconsole/compare/v0.7.1...v0.7.2) (2020-07-01)


### Bug Fixes

* show dashed lines for empty values instead of removing them ([#79](http://github.com/lyft/flyteconsole/issues/79)) ([0eaaecd](http://github.com/lyft/flyteconsole/commit/0eaaecd2f7c5cf7e2ed33bebaf4f3a4fc57ba697))

## [0.7.1](http://github.com/lyft/flyteconsole/compare/v0.7.0...v0.7.1) (2020-07-01)


### Bug Fixes

* Expose cluster information for Workflow Executions ([#78](http://github.com/lyft/flyteconsole/issues/78)) ([d43a1e3](http://github.com/lyft/flyteconsole/commit/d43a1e393ef7b0651f9774460e3d7166a3b57f17))

# [0.7.0](http://github.com/lyft/flyteconsole/compare/v0.6.0...v0.7.0) (2020-06-30)


### Features

* improve user experience for nested NodeExecutions ([#77](http://github.com/lyft/flyteconsole/issues/77)) ([58ed1a4](http://github.com/lyft/flyteconsole/commit/58ed1a4176afeb0c13c34cf0cced13efa91ea7b4)), closes [#70](http://github.com/lyft/flyteconsole/issues/70) [#72](http://github.com/lyft/flyteconsole/issues/72) [#75](http://github.com/lyft/flyteconsole/issues/75) [#76](http://github.com/lyft/flyteconsole/issues/76)

# [0.6.0](http://github.com/lyft/flyteconsole/compare/v0.5.3...v0.6.0) (2020-06-30)


### Features

* Adding automatic image building and release management ([#69](http://github.com/lyft/flyteconsole/issues/69)) ([90362e2](http://github.com/lyft/flyteconsole/commit/90362e2fe5ef9a12ca3826ce66793df09c79a906))

# [0.6.0](http://github.com/lyft/flyteconsole/compare/v0.5.2...v0.6.0) (2020-05-13)


### Bug Fixes

* adds proper plugin to update package.json when releasing ([8640ac7](http://github.com/lyft/flyteconsole/commit/8640ac7f8e9620aa1484c30d24c50fe7862021ad))


### Features

* testing publishing of a release ([3c77418](http://github.com/lyft/flyteconsole/commit/3c774183c897a17c29481eba65b93a358b62cc8e))

# [0.6.0](http://github.com/lyft/flyteconsole/compare/v0.5.2...v0.6.0) (2020-05-13)


### Bug Fixes

* adds proper plugin to update package.json when releasing ([8640ac7](http://github.com/lyft/flyteconsole/commit/8640ac7f8e9620aa1484c30d24c50fe7862021ad))


### Features

* testing publishing of a release ([3c77418](http://github.com/lyft/flyteconsole/commit/3c774183c897a17c29481eba65b93a358b62cc8e))
