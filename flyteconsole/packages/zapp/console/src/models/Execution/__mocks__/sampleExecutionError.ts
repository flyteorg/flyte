export const sampleError = `task failed, TaskFailedWithError: task failed, USER:Unknown: user-error. Message: [Traceback (most recent call last):

    File "/myflyteproject/test.py", line 167, in _entry_point
      return _system_func(*args, **kwargs)
    File "/myflyteproject/output.py", line 46, in set
      sdk_value = self.from_python(value)
    File "/myflyteproject/schema.py", line 108, in from_python_std
      schema = t_value.cast_to(cls.schema_type)
    File "/myflyteproject/schema.py", line 857, in cast_to

Message:

  Type error!  Received: Datetime, Expected: Datetime. Cannot cast because the column type for column 'time_period_start' does not match.

User error.]`;
