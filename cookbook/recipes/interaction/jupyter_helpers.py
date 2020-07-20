from IPython.display import Image, display
from flytekit.configuration import set_flyte_config_file, platform


def display_images(paths, format=None):
    for p in paths:
        display(Image(p, format=format))


def config_load():
    set_flyte_config_file("notebook.config")
    print("Connected to {}".format(platform.URL.get()))


def print_console_url(exc):
    print("http://{}/console/projects/{}/domains/{}/executions/{}".format(platform.URL.get(), exc.id.project, exc.id.domain, exc.id.name))
