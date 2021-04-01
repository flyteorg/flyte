# General
* Added CoPilot configuration to sandbox manifest
* Updated and streamlined documentation navigation and themes.  Better content organization coming soon!

# Performance
* Improved execution performance including:
  - Reduced cache lookups
  - Improved GetWorkflowExecution performance
  - Capped max number of nodes in each propeller round
  - Misc. propeller performance tweaks
* TaskTemplate offloading

# Housekeeping
* Migrated Datacatalog protobuf definitions to flyteidl [thanks @tnsetting]
* Upgraded stow version used in flytestdlib
* Moved off lyft kubernetes forks and onto official kubernetes library dependencies
* Revamped pod tasks to use official kubernetes python client library for defining PodSpecs

# Events
* Richer event metadata for task executions
* Better merging of custom info across task events

# Bug fixes
* Resolved non-backwards protobuf role changes that prevented launching single task executions [thanks @kanterov]
* Better handling of large workflows and errors in flytepropeller

# Flytekit (Python)
* Access to secrets
* Bug fixes around the 0.16 release.
  * Use original FlyteFile/FlyteDirectory 
  * Fix serialization of pod specs in pod plugin [thanks @jeevb]
  * Accept auth role arg in single task execution
  * Fixed task resolver in map task
  * Requests and limits added to ContainerTask [thanks @migueltol22]

