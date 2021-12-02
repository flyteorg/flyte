# [RFC] [WIP] Caching of offloaded types

**Authors:**

- @eapolinario

## 1 Executive Summary

*A short paragraph or bullet list that quickly explains what you're trying to do.*




## 2 Motivation

*What motivates this proposal, and why is it important?*

*Here, we aim to get comfortable articulating the value of our actions.*

## 3 Proposed Implementation

*This is the core of your proposal, and its purpose is to help you think through the problem because [writing is thinking](https://medium.learningbyshipping.com/writing-is-thinking-an-annotated-twitter-thread-2a75fe07fade).*

*Consider:*

- *using diagrams to help illustrate your ideas.*
- *including code examples if you're proposing an interface or system contract.*
- *linking to project briefs or wireframes that are relevant.*


Proposal: expose a hash in the Literal object and develop machinery in the SDKs to fill in that value. For example, in flytekit we plan to use [external annotations](https://github.com/python/typing/issues/600) to annotate the offloaded objects that should be cached.

But first let's take a detour to explain how values are cached both in local and remote executions.


```python
@task
def foo(a: int, b: str) -> CachedPandasDataframe:
    df = pd.Dataframe(...)
    ...
    return df

@task(cached=True, version="1.0")
def bar(df: pd.Dataframe) -> int:
    ...
    
@workflow
def wf(a: int, b: str):
    df = foo(a=a, b=b)
    # Note that the task `Bar` is marked as cached. The intent here is to make sure that
    # we are able to 
    # In calls to `bar` we now consider the value of `df` in the generation
    v = bar(df=df) 
```

## 4 Metrics & Dashboards

*What are the main metrics we should be measuring? For example, when interacting with an external system, it might be the external system latency. When adding a new table, how fast would it fill up?*

N/A?

TODO: what's the observability provided by data catalog? Do we know about cache hits and misses?

## 5 Drawbacks

*Are there any reasons why we should not do this? Here we aim to evaluate risk and check ourselves.*

## 6 Alternatives

*What are other ways of achieving the same outcome?*

## 7 Potential Impact and Dependencies

*Here, we aim to be mindful of our environment and generate empathy towards others who may be impacted by our decisions.*

- *What other systems or teams are affected by this proposal?*
- *How could this be exploited by malicious attackers?*

## 8 Unresolved questions

*What parts of the proposal are still being defined or not covered by this proposal?*

## 9 Conclusion

*Here, we briefly outline why this is the right decision to make at this time and move forward!*

## 10 RFC Process Guide, remove this section when done

*By writing an RFC, you're giving insight to your team on the direction you're taking. There may not be a right or better decision in many cases, but we will likely learn from it. By authoring, you're making a decision on where you want us to go and are looking for feedback on this direction from your team members, but ultimately the decision is yours.*

This document is a:

- thinking exercise, prototype with words.
- historical record, its value may decrease over time.
- way to broadcast information.
- mechanism to build trust.
- tool to empower.
- communication channel.

This document is not:

- a request for permission.
- the most up to date representation of any process or system

**Checklist:**

- [x]  Copy template
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
