# Union.ai Actors


Setup
Make sure you have a virtual environment set up and activated.  Use uv.

Go into the folder this readme is in (`cd fasttask/worker-v2` from base of repo). Then,
```bash
uv venv --python 3.13
source .venv/bin/activate
```
Other setup is needed. Should add below as people try them.


Iteration cycle

`make build-wheels` - to build the wheels. This compiles all the Rust code 

From `examples` folder of unionv2, using a different virtual environment. Make sure to update the wheel folder in the
example code.

```bash
flyte -c ~/.flyte/config-k3d.yaml run -d development reuse/oomer_reuse.py failure_recovery
```

Note when hitting Ctrl-C during the make command, the docker process doesn't stop. You'll have to docker stop/rm the container
if you don't want to wait around for the build to finish.

For releasing
```bash
make build-wheels SETUPTOOLS_SCM_PRETEND_VERSION=0.1.0b0
```
but wheels need to be renamed to this pattern
```bash
mv unionai_reuse-0.1.4b0-cp38-abi3-linux_aarch64.whl unionai_reuse-0.1.4b0-cp38-abi3-manylinux_2_28_aarch64.whl
mv unionai_reuse-0.1.4b0-cp38-abi3-linux_x86_64.whl unionai_reuse-0.1.4b0-cp38-abi3-manylinux_2_28_x86_64.whl
```

Need to figure out how to make this automatic.
