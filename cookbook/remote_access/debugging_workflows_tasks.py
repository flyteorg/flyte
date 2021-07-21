"""
Debugging Workflow and Task Executions
--------------------------------------

The inspection of task and workflow execution would provide you log links to debug things further

Using ``--details`` flag would shows you node executions view with log links. ::


        └── n1 - FAILED - 2021-06-30 08:51:07.3111846 +0000 UTC - 2021-06-30 08:51:17.192852 +0000 UTC
        │   ├── Attempt :0
        │       └── Task - FAILED - 2021-06-30 08:51:07.3111846 +0000 UTC - 2021-06-30 08:51:17.192852 +0000 UTC
        │       └── Logs :
        │           └── Name :Kubernetes Logs (User)
        │           └── URI :http://localhost:30082/#/log/flytectldemo-development/f3a5a4034960f4aa1a09-n1-0/pod?namespace=flytectldemo-development

Additionally you can check the pods launched by flyte in <project>-<domain> namespace ::

    kubectl get pods -n <project>-<domain>

The launched pods will have a prefix of execution name along with suffix of nodeId ::

        NAME                        READY   STATUS             RESTARTS   AGE
        f65009af77f284e50959-n0-0   0/1     ErrImagePull       0          18h

So here the investigation can move ahead by describing the pod and checking the issue with Image pull.

"""
