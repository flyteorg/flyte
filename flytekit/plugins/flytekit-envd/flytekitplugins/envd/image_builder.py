import os
import pathlib
import shutil
import subprocess
from importlib import metadata

import click
from packaging.version import Version

from flytekit.configuration import DefaultImages
from flytekit.core import context_manager
from flytekit.core.constants import REQUIREMENTS_FILE_NAME
from flytekit.image_spec.image_spec import _F_IMG_ID, ImageBuildEngine, ImageSpec, ImageSpecBuilder


class EnvdImageSpecBuilder(ImageSpecBuilder):
    """
    This class is used to build a docker image using envd.
    """

    def execute_command(self, command):
        click.secho(f"Run command: {command} ", fg="blue")
        p = subprocess.Popen(command.split(), stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        result = []
        for line in iter(p.stdout.readline, ""):
            if p.poll() is not None:
                break
            if line.decode().strip() != "":
                output = line.decode().strip()
                click.secho(output, fg="blue")
                result.append(output)

        if p.returncode != 0:
            _, stderr = p.communicate()
            raise Exception(f"failed to run command {command} with error {stderr}")

        return result

    def build_image(self, image_spec: ImageSpec):
        cfg_path = create_envd_config(image_spec)

        if image_spec.registry_config:
            bootstrap_command = f"envd bootstrap --registry-config {image_spec.registry_config}"
            self.execute_command(bootstrap_command)

        build_command = f"envd build --path {pathlib.Path(cfg_path).parent}  --platform {image_spec.platform}"
        if image_spec.registry:
            build_command += f" --output type=image,name={image_spec.image_name()},push=true"
        self.execute_command(build_command)


def _create_str_from_package_list(packages):
    if packages is None:
        return ""
    return ", ".join(f'"{name}"' for name in packages)


def create_envd_config(image_spec: ImageSpec) -> str:
    base_image = DefaultImages.default_image() if image_spec.base_image is None else image_spec.base_image
    if image_spec.cuda:
        if image_spec.python_version is None:
            raise Exception("python_version is required when cuda and cudnn are specified")
        base_image = "ubuntu20.04"

    python_packages = _create_str_from_package_list(image_spec.packages)
    conda_packages = _create_str_from_package_list(image_spec.conda_packages)
    run_commands = _create_str_from_package_list(image_spec.commands)
    conda_channels = _create_str_from_package_list(image_spec.conda_channels)
    apt_packages = _create_str_from_package_list(image_spec.apt_packages)
    env = {"PYTHONPATH": "/root", _F_IMG_ID: image_spec.image_name()}

    if image_spec.env:
        env.update(image_spec.env)
    pip_index = "https://pypi.org/simple" if image_spec.pip_index is None else image_spec.pip_index

    envd_config = f"""# syntax=v1

def build():
    base(image="{base_image}", dev=False)
    run(commands=[{run_commands}])
    install.python_packages(name=[{python_packages}])
    install.apt_packages(name=[{apt_packages}])
    runtime.environ(env={env}, extra_path=['/root'])
    config.pip_index(url="{pip_index}")
"""
    ctx = context_manager.FlyteContextManager.current_context()
    cfg_path = ctx.file_access.get_random_local_path("build.envd")
    pathlib.Path(cfg_path).parent.mkdir(parents=True, exist_ok=True)

    if conda_packages:
        envd_config += "    install.conda(use_mamba=True)\n"
        envd_config += f"    install.conda_packages(name=[{conda_packages}], channel=[{conda_channels}])\n"

    if image_spec.requirements:
        requirement_path = f"{pathlib.Path(cfg_path).parent}{os.sep}{REQUIREMENTS_FILE_NAME}"
        shutil.copyfile(image_spec.requirements, requirement_path)
        envd_config += f'    install.python_packages(requirements="{REQUIREMENTS_FILE_NAME}")\n'

    if image_spec.python_version:
        # Indentation is required by envd
        envd_config += f'    install.python(version="{image_spec.python_version}")\n'

    if image_spec.cuda:
        cudnn = image_spec.cudnn if image_spec.cudnn else ""
        envd_config += f'    install.cuda(version="{image_spec.cuda}", cudnn="{cudnn}")\n'

    if image_spec.source_root:
        shutil.copytree(image_spec.source_root, pathlib.Path(cfg_path).parent, dirs_exist_ok=True)

        envd_version = metadata.version("envd")
        # Indentation is required by envd
        if Version(envd_version) <= Version("0.3.37"):
            envd_config += '    io.copy(host_path="./", envd_path="/root")'
        else:
            envd_config += '    io.copy(source="./", target="/root")'

    with open(cfg_path, "w+") as f:
        f.write(envd_config)

    return cfg_path


ImageBuildEngine.register("envd", EnvdImageSpecBuilder())
