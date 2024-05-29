<html>
    <p align="center"> 
        <img src="https://raw.githubusercontent.com/flyteorg/static-resources/main/common/flyte_circle_gradient_1_4x4.png" alt="Flyte Logo" width="100">
    </p>
    <h1 align="center">
        FlyteCTL
    </h1>
    <p align="center">
       Flyte's official command-line interface
    </p>
    <h3 align="center">
        <a href="https://flytectl.rtfd.io">Documentation</a>
        <span> · </span>
        <a href="https://docs.flyte.org/en/latest/flytectl/contribute.html">Contribution Guide</a>
    </h3>
</html>

[![Docs](https://readthedocs.org/projects/flytectl/badge/?version=latest&style=plastic)](https://flytectl.rtfd.io)
[![Current Release](https://img.shields.io/github/release/flyteorg/flyte/flytectl.svg)](https://github.com/flyteorg/flytectl/releases/latest)
![Master](https://github.com/flyteorg/flyte/flytectl/workflows/Master/badge.svg)
[![GoDoc](https://godoc.org/github.com/flyteorg/flyte/flytectl?status.svg)](https://pkg.go.dev/mod/github.com/flyteorg/flytectl)
[![License](https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
[![CodeCoverage](https://img.shields.io/codecov/c/github/flyteorg/flyte/flytectl.svg)](https://codecov.io/gh/flyteorg/flytectl)
[![Go Report Card](https://goreportcard.com/badge/github.com/flyteorg/flyte/flytectl)](https://goreportcard.com/report/github.com/flyteorg/flytectl)
![Commit activity](https://img.shields.io/github/commit-activity/w/lyft/flytectl.svg?style=plastic)
![Commit since last release](https://img.shields.io/github/commits-since/lyft/flytectl/latest.svg?style=plastic)
[![Slack](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](https://slack.flyte.org)

Flytectl was designed as a portable and lightweight command-line interface to work with Flyte. It is written in Golang and accesses [FlyteAdmin](https://github.com/flyteorg/flyteadmin/), the control plane for Flyte.

## 🚀 Quick Start

1. Install Flytectl with bash or shell script.

    * Bash
        ```bash
        $ brew install flyteorg/homebrew-tap/flytectl
        ```
    * Shell script
        ```bash
        $ curl -sL https://ctl.flyte.org/install | bash
        ```
2. (Optional) `flytectl upgrade` provides a general interface to upgrading Flytectl; run the command in the output.

3. Start Sandbox using Flytectl.
    ```bash
    $ flytectl sandbox start 
    ```

4. Register examples.
    ```bash
    # Register core workflows 
    $ flytectl register examples -d development -p flytesnacks
    ```

<html>
    <h2 id="contribution-guide"> 
        📖 How to Contribute to Flytectl
    </h2>
</html>

You can find the detailed contribution guide [here](docs/source/contribute.rst).

<html>
    <h2 id="file-an-issue"> 
        🐞 File an Issue
    </h2>
</html>

Refer to the [issues](https://docs.flyte.org/en/latest/community/contribute.html#file-an-issue) section in the contribution
guide if you'd like to file an issue.
