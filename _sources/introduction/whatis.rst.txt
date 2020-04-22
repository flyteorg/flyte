.. _introduction-what-is:

###############
What is Flyte?
###############

Workflows are ideal for organizing complex processes. As Lyft has evolved, developers have turned to workflows to solve an increasing number of business problems. Flyte was created as an end-to-end solution for developing, executing, and monitoring workflows reliably at scale.

The system powers products all over Lyft, including Data Science, Pricing, Fraud, Driver Engagement, Locations, ETA, and Autonomous. 

Flyte handles all the overhead involved in executing complex workflows, including hardware provisioning, scheduling, data storage, and monitoring. This allows developers to focus solely on writing their workflow logic.

Development
-----------

In Flyte, workflows are expressed as a graph. To create a workflow, you must compose a graph that conforms to our strict specification. We provide developer tools that make it easy to define robust workflows that conform to this spec. The spec allows us to perform up-front validation and ensure data is passed properly from one task to the next. This practice has proven extremely useful at Lyft in time and cost savings.

In as few as 20 lines of code, a user can define a custom workflow, submit the graph to Flyte, and snooze while Flyte fulfills the execution on a schedule, or on-demand.

Execution
---------

When fulfilling a workflow execution, Flyte will dynamically provision the resources necessary to complete the work. Tasks are launched using docker containers to ensure consistency across environments. The system can distribute workloads across 100s of machines, making it useful for production model training and batch compute at Lyft.

As tasks complete, unused resources will be deprovisioned to control costs. 

Workflows come with a variety of knobs such as automatic retries, and custom notifications on success or failure.

Monitoring
----------

The Flyte system continuously monitors workflow executions throughout their lifecycle. You can visit the Flyte web interface at any time to track the status of a workflow.

The web interface provides the ability to drill down into the individual workflow steps, and get rich error messaging to triage any issues.
