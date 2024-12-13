# Flyte v1.11.0 Release Notes

We're excited to announce the release of Flyte v1.11.0! This version brings a host of improvements, bug fixes, and new features designed to enhance your experience with Flyte. From operational enhancements to documentation updates, this release aims to make Flyte more robust, user-friendly, and feature-rich.

## Highlights

- **Agents hit General Availability (GA):** Agents, now in General Availability, are long-running, stateless services that facilitate asynchronous job launches on platforms like Databricks or Snowflake and enable external service calls. They are versatile, supporting implementations in any language through a protobuf interface, enhancing Flyte's flexibility and operational efficiency.
- **Improved Caching:** Support for loading cached sublists with multiple data types has been introduced, eliminating issues related to cache retrieval across varied data formats.
- **Tracing and Observability:** The introduction of opentelemetry BlobstoreClientTracer in flyteadmin enhances observability, allowing for better monitoring and troubleshooting.
- **Security Enhancements:** Added securityContext configuration to Flyte-core charts, strengthening the security posture of Flyte deployments.
- **Documentation Overhaul:** Continuous improvements and updates have been made to the documentation, fixing broken links and updating content for better clarity and usability.
- **Operational Improvements:** This release introduces enhancements such as adding a service account for V1 Ray Jobs, caching console assets in a single binary, and conditional mounting of secrets to improve the operational efficiency of Flyte. Additionally, we are removing `kustomize` from our deployment process to simplify the configuration and management of Flyte instances, making it easier for users to maintain and streamline their deployment workflows.


## Bug Fixes

- **Fixed Literal in Launchplan:** Added fixed_literal to the launchplan template, addressing issues with hardcoded values in workflows.
- **Corrected Metadata and Resources:** Fixes have been applied to correct IsParent metadata in ArrayNode eventing and to address invalid "resources" scope issues in deployment configurations.
- **Enhanced Stability and Performance:** Numerous bug fixes have been implemented to address stability and performance issues, including fixes for data catalog errors, yaml comment errors in pod template examples, and more.

## Documentation and Guides

- **Comprehensive Guides:** New guides and documentation updates have been added, including a ChatGPT Agent Setup guide and an Airflow migration guide. Improvements in documentation for developing agents have been integrated into the broader enhancements for this release.
- **Updated Troubleshooting and Configuration Docs:** New troubleshooting guides for spark task execution and updates to deployment configuration documents enhance the knowledge base for Flyte users.

## Contributors

We extend our deepest gratitude to all the contributors who made this release possible. Special shoutouts to @neilisaur, @lowc1012, @MortalHappiness, @novahow, and @pryce-turner for making their first contributions!

**For a full list of changes, enhancements, and bug fixes, visit our [changelog](https://github.com/flyteorg/flyte/compare/v1.10.7...v1.11.0).**

Thank you for your continued support of Flyte. We look forward to hearing your feedback on this release!
