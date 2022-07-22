[Back to Cookbook Menu](../../..)

# Tasks in Flytekit can be tested

## Unit test the generic_type_task

Snippet
```python
    from types.generic import generic_type_task

    in1 = {
        'a': 'hello',
        'b': 'how are you',
        'c': ['array'],
        'd': {
            'nested': 'value',
        },
    }
    output = generic_type_task.unit_test(custom=in1)
    assert output['counts'] == {
        'a': 5,
        'b': 11,
        'c': ['array'],
        'd': {
            'nested': 'value',
        },
    }
```

 - As shown above, inputs can be passed as a dictionary with each attribute as a key in the dictionary.
 - as long as the task is decorated with a flytekit decorator, a `unit_test` method is automatically defined
 - the outputs are returned as a dictionary as well
