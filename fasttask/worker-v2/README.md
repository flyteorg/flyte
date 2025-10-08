# Union.ai Actors

## Developer Setup

Make sure you have a virtual environment set up and activated.  Use uv.

Go into the folder this readme is in (`cd fasttask/worker-v2` from base of repo). Then,
```bash
uv venv --python 3.13
source .venv/bin/activate
```

To build wheels containing the Rust code you'll also need this for now.
`make build-builders` to build the builder images

This is because the `manylinux` by default don't come with rust installed. Installing them into the manylinux
image every time is too time-consuming, so these are really lightweight images that just install rust/cargo/maturin.

We can get rid of these once we find better ways to build multi-arch wheels, or find maintained public images. I'm sure
they exist, just haven't looked in depth.

### End to end iterating
`make build-wheels` to build the wheels.  This compiles all the Rust code

Do something like this on the Python side

```python
actor_dist_folder = Path("/Users/ytong/go/src/github.com/unionai/flyte/fasttask/worker-v2/dist")
wheel_layer = PythonWheels(wheel_dir=actor_dist_folder, package_name="unionai-reuse")
base = flyte.Image.from_debian_base()
actor_image = base.clone(addl_layer=wheel_layer)
```

and also run `make dist` as you would normally when iterating just in the flyte-sdk.

From `examples` folder of `flyte-sdk`, using a different virtual environment. Make sure to update the wheel folder in the
example code.

```bash
flyte -c ~/.flyte/config-k3d.yaml run -d development reuse/oomer_reuse.py failure_recovery
```

Note when hitting Ctrl-C during the make command, the docker process doesn't stop. You'll have to docker stop/rm the container
if you don't want to wait around for the build to finish.

### Executor iteration

To make a change to the python side
* Just make sure that the flyte-sdk library is installed in editable mode and changes should get picked up. Test with a
  print statement to confirm.

To make a change to the rust side
These instructions may be incomplete

Run the test server so that the executor has something to attach to.
```bash
cargo run -p unionai_actor_bridge --bin test_server
```

Another test service is available called `ping_test` - it doesn't ping, it calls devbox one.

```bash
cargo run -p unionai_actor_bridge --bin ping_test
cargo run -p unionai_actor_bridge --bin ping_test -- --test-cancel --cancel-wait 10
```

Run the executor (see below for building)
```bash
unionai-actor-executor --executor-registration-addr 127.0.0.1:15606 --id 0 --num-workers 1
```

This adds local flyte-sdk changes to the path. Also adds the examples folder to the path so that the manual override added (need to document),
can be picked up.

When using this, you'll also need to pull this test branch in flyte-sdk, which helps short-circuit part of the task execution process.
https://github.com/flyteorg/flyte-sdk/compare/rusty-test-abort?expand=1

```bash
PYTHONPATH=/Users/ytong/go/src/github.com/flyteorg/flyte-sdk/examples:/Users/ytong/go/src/github.com/flyteorg/flyte-sdk/src: unionai-actor-executor --executor-registration-addr 127.0.0.1:15606 --id 0 --num-workers 1
```

Will need to `uv pip install -e .` from base folder (the dir this readme is in) to pick up changes - need to investigate how to run maturin develop from the executor/ folder, which doesn't seem to work currently. Changes don't get pick up.


### Bridge iteration
fill this in next time someone is working only on the bridge side.


### SDK side
In addition to the test branch above, to get the flyte-sdk to pick up the whls that are build as part of the `make build-wheels` target,
you can use this Python snippet as the image in the `TaskEnvironment`

```python
from pathlib import Path
from flyte._image import PythonWheels

actor_dist_folder = Path("/Users/yourusername/go/src/github.com/unionai/flyte/fasttask/worker-v2/dist")
wheel_layer = PythonWheels(wheel_dir=actor_dist_folder, package_name="unionai-reuse")
base = flyte.Image.from_debian_base()
actor_image = base.clone(addl_layer=wheel_layer)
```


## Releasing
Most of this section should not exist - still awaiting proper CI.
```bash
make build-wheels SETUPTOOLS_SCM_PRETEND_VERSION=0.1.7a0b0
SETUPTOOLS_SCM_PRETEND_VERSION=0.1.7a0 python -m build --wheel
```
but wheels need to be renamed to this pattern
```bash
mv unionai_reuse-0.1.7a0-cp38-abi3-linux_aarch64.whl unionai_reuse-0.1.7a0-cp38-abi3-manylinux_2_28_aarch64.whl;
mv unionai_reuse-0.1.7a0-cp38-abi3-linux_x86_64.whl unionai_reuse-0.1.7a0-cp38-abi3-manylinux_2_28_x86_64.whl
```

Need to figure out how to make this automatic.
