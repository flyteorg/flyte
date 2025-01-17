from collections import OrderedDict

import flytekit.configuration
from flytekit.configuration import Image, ImageConfig
from flytekit.tools.translator import get_serializable

from .test_task import tk as not_tk


def test_sql_lhs():
    assert not_tk.lhs == "tk"


def test_sql_command():
    default_img = Image(name="default", fqn="test", tag="tag")
    serialization_settings = flytekit.configuration.SerializationSettings(
        project="project",
        domain="domain",
        version="version",
        env=None,
        image_config=ImageConfig(default_image=default_img, images=[default_img]),
    )
    srz_t = get_serializable(OrderedDict(), serialization_settings, not_tk)
    assert srz_t.template.container.args[-5:] == [
        "--resolver",
        "flytekit.core.python_customized_container_task.default_task_template_resolver",
        "--",
        "{{.taskTemplatePath}}",
        "flytekitplugins.sqlalchemy.task.SQLAlchemyTaskExecutor",
    ]
