Golang Dockerfile
~~~~~~~~~~~~~~~~~

Provides a Dockerfile that produces a small image.

**To Enable:**

Add ``flyteorg/golang_dockerfile`` to your ``boilerplate/update.cfg`` file.

Create and configure a ``make linux_compile`` target that compiles your go binaries to the ``/artifacts`` directory ::

  .PHONY: linux_compile
  linux_compile:
    RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /artifacts {{ packages }}

All binaries compiled to ``/artifacts`` will be available at ``/bin`` in your final image.
