# Flytekit Airflow Plugin
Airflow plugin allows you to seamlessly run Airflow tasks in the Flyte workflow without changing any code.

- Compile Airflow tasks to Flyte tasks
- Use Airflow sensors/operators in Flyte workflows
- Add support running Airflow tasks locally without running a cluster

## Example
```python
from airflow.sensors.filesystem import FileSensor
from flytekit import task, workflow

@task()
def t1():
    print("flyte")


@workflow
def wf():
    sensor = FileSensor(task_id="id", filepath="/tmp/1234")
    sensor >> t1()


if __name__ == '__main__':
    wf()
```


To install the plugin, run the following command:

```bash
pip install flytekitplugins-airflow
```
