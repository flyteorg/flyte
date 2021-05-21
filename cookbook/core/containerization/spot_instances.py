"""
Using Spot/Preemptible Instances
--------------------------------

"""
# %%
# What Are Spot/Preemptible Instances?
# ====================================
# Spot instances are unused EC2 capacity in AWS. `Spot instances <https://aws.amazon.com/ec2/spot/?cards.sort-by=item.additionalFields.startDateTime&cards.sort-order=asc>`_ can result in up to 90% savings on on-demand prices. The caveat is that these instances can be preempted at any point and no longer be available for use. This can happen due to:
#
# * Price – The spot price is greater than your maximum price.
# * Capacity – If there are not enough unused EC2 instances to meet the demand for spot instances, Amazon EC2 interrupts spot instances. Amazon EC2 determines the order in which the instances are interrupted.
# * Constraints – If your request includes a constraint such as a launch group or an Availability Zone group, these spot instances are terminated as a group when the constraint can no longer be met.
#
# Generally, most spot instances are obtained for around 2 hours (median), with the floor being about 20 minutes and the ceiling of unbounded duration.
#
# .. note::
#
#   Spot Instances are called ``Preemptible Instances`` in the GCP terminology.
#
# Setting up Spot Instances
# ^^^^^^^^^^^^^^^^^^^^^^^^^
# - AWS: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/spot-requests.html
# - GCP: https://cloud.google.com/compute/docs/instances/create-start-preemptible-instance
#
# If an auto-scaling group (ASG) is set up, you may want to isolate the tasks you want to trigger on spot/preemptible instances from the regular workloads.
# This can be done by setting taints and tolerations using the `config <https://github.com/flyteorg/flyteplugins/blob/60b94c688ef2b98aa53a9224b529ac672af04540/go/tasks/pluginmachinery/flytek8s/config/config.go#L84-L92>`__ available at ``flyteorg/flyteplugins`` repo.
#
# .. admonition:: What's an ASG for a spot/preemptible instance?
#
#   When your spot/preemptible instance is terminated, ASG attempts to launch a replacement instance to maintain the desired capacity for the group.

# %%
# What Are Interruptible Tasks?
# =============================
#
# If specified, the ``interruptible flag`` is added to the task definition and signals to the Flyte engine that it may be scheduled on machines that may be preempted, such as AWS spot instances. This is low-hanging fruit for any cost-savings initiative. 
#
# Setting Interruptible
# ^^^^^^^^^^^^^^^^^^^^^
# To run your workload on a spot/preemptible instance, you can set interruptible to ``True``. For example:

@task(cache_version='1', interruptible=True)
def add_one_and_print(value_to_print: int) -> int:
    return value_to_print + 1

# %%
# By setting this value, Flyte will schedule your task on an auto-scaling group (ASG) with only spot instances. 
# 
# .. note::
# If your task gets preempted, Flyte will retry your task on a non-spot (regular) instance. This retry will not count towards a retry that a user sets.
#
# Which Tasks Should Be Set To Interruptible?
# ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
# Most Flyte workloads should be good candidates for spot instances. 
# If your task does NOT exhibit the following properties, you can set ``interruptible`` to true.
#
# * Time-sensitive: It needs to run now and can not have any unexpected delays.
# * Side Effects: The task is not idempotent, and retrying will cause issues.
# * Long-Running Tasks: The task takes > 2 hours. Having an interruption during this time frame could potentially waste a lot of computation.
#
# In a nutshell, you should use spot/preemptible instances when you want to reduce the total cost of running jobs at the expense of potential delays in execution due to restarts.

# %%
# How to Recover From Interruptions?
# ==================================
#
# .. NOTE::
#
#     Coming soon 🛠
#
