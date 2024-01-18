# [RFC Template] Flytekit Notification System

**Authors:**

- [@Future-Outlier](https://github.com/Future-Outlier)
- [@Kevin Su](https://github.com/pingsutw)
- [@ByronHsu](https://github.com/ByronHsu)

## 1 Executive Summary

Flyte now can send notification in the pod by external API services (but not Flyte Admin) for better flexibility.

## 2 Motivation

Now, we can only send email when using launch plan by [flyteadmin](https://github.com/flyteorg/flyte/tree/master/flyteadmin/pkg/async/notifications).
However, this is not flexible enough, there are usecases people want to send message in the pod or the agent service.

## 3 Proposed Implementation

There are 3 main components for this feature implementation.

Take this [PR](https://github.com/flyteorg/flytekit/pull/2115/files) for example.

- base class `BaseNotifier` for any notifier. e.g.(Slack, Sendgrid ...)
- Inherit `BaseNotifier` and make your own notifier. (Take Slack Notifier for example in this PR.)
- Design your Notifier Logic in the plugin. (Take Flyin for example in thie PR.)

In this case, we implement a class `NotifierExecutor` to decide when to send notifications.
You can decide your own way to send notifications, not necessary, but recommend to create a new class to do it.

## 4 Metrics & Dashboards

## 5 Drawbacks

* Make users confused the difference between flytekit notification system and flyte admin notification system.

## 6 Alternatives

Give pod permissions to contact with flyteadmin in runtime, and send notification through flyteadmin.

## 7 Potential Impact and Dependencies

## 8 Unresolved questions

## 9 Conclusion

This feature will provide flexibility when users create flyteplugins.

**Checklist:**

- [ ]  Copy template
- [ ]  Draft RFC (think of it as a wireframe)
- [ ]  Share as WIP with folks you trust to gut-check
- [ ]  Send pull request when comfortable
- [ ]  Label accordingly
- [ ]  Assign reviewers
- [ ]  Merge PR

**Recommendations**

- Tag RFC title with [WIP] if you're still ironing out details.
- Tag RFC title with [Newbie] if you're trying out something experimental or you're not entirely convinced of what you're proposing.
- Tag RFC title with [RR] if you'd like to schedule a review request to discuss the RFC.
- If there are areas that you're not convinced on, tag people who you consider may know about this and ask for their input.
- If you have doubts, ask on [#feature-discussions](https://slack.com/app_redirect?channel=CPQ3ZFQ84&team=TN89P6GGK) for help moving something forward.
