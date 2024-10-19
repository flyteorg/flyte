(flyte_type_system_and_data_passing)=

# Flyte type system and Data passing

```{eval-rst}
.. tags:: Basic
```

### How It Works

The Flyte type system is designed to ensure that data passed between tasks and workflows adheres to specific formats and types. By leveraging Python’s type hints and type annotations, Flyte can automatically validate inputs and outputs in tasks to ensure correctness.

Flyte supports a wide range of types, including primitive types, collections, and custom types. These are enforced through Flyte’s type engine, which ensures that tasks interact with the correct data structures.

Supported Types:

	1. Primitive Types:
	    -	int, float, bool, str: Basic data types commonly used in Python functions.

		```{rli} # Define a task to add two integers
			@task
			def add(x: int, y: int) -> int:
    			return x + y

			# Define a task to multiply two integers
			@task
			def multiply(x: int, y: int) -> int:
    		return x * y
		:caption: basics/task.py
		:lines: 1
		```

	2. Collections:
	    -	List[T]: Represents a list of items where each item is of type T.
	    -	Dict[K, V]: Represents a dictionary where keys are of type K and values are of type V.

		```{rli} 
			@task
			def process_numbers(numbers: List[int]) -> float:
    			"""Calculate the average of a list of integers."""
    			return sum(numbers) / len(numbers)

			@task
			def summarize_data(data: Dict[str, int]) -> str:
				"""Create a summary of the provided data."""
				total = sum(data.values())
				summary = ", ".join([f"{key}: {value}" for key, value in data.items()])
				return f"Summary: {summary}. Total: {total}"
		:caption: basics/task.py
		:lines: 1
		```

	3. Custom Types:
	    -	FlyteFile: For handling file inputs and outputs.
	    -	FlyteSchema: Used for handling structured data (like tables or dataframes).
	    -	StructuredDataset: Used for handling data in formats like Pandas or Spark DataFrames.

		```{rli} 
			@task
			def process_file(file: FlyteFile) -> str:
    		# Simulate processing the file and returning its name
    		return f"Processed file: {file}"
		:caption: basics/task.py
		:lines: 1
		```

### When to Use the Type System

The Flyte type system should be used whenever you define tasks and workflows, as it provides the following benefits:

	•	Input and Output Validation: By specifying types, Flyte ensures that the correct data is passed into and out of tasks. This minimizes runtime errors caused by unexpected data formats.
	•	Improved Debugging: Since Flyte raises errors when there is a type mismatch, debugging becomes easier, as you know exactly where things went wrong.
	•	Seamless Data Flow: By explicitly declaring types, Flyte can efficiently serialize and deserialize data, allowing it to move between tasks without manual intervention.

### How the Type Engine Works:

Flyte’s type engine is responsible for validating and transforming data based on the specified types. Here’s a quick look at the process:

	•	Type Inference: When you define a task, Flyte uses the type hints provided in Python function signatures to infer the expected input and output types.
	•	Type Checking: During execution, Flyte checks that the data being passed to a task matches the declared types. If there’s a mismatch, the type engine will raise an error.
	•	Serialization/Deserialization: For complex types like files or datasets, Flyte automatically serializes the data (stores it in an appropriate format) and deserializes it (retrieves it) when needed by subsequent tasks.

### Limitations of the Flyte Type System:

While Flyte’s type system is robust, there are a few limitations to be aware of:

	•	Unsupported Types: Some Python objects, such as custom classes or types that Flyte cannot serialize, might not work directly within the type system. You may need to implement custom type transformers for complex types.
	•	Dynamic Types: Flyte works best with statically typed data. If your data types are dynamically determined at runtime, you may run into issues or need to handle type enforcement manually.

## Data Passing in Workflows

Data passing refers to how Flyte handles the flow of data between tasks and workflows.

### A Clear Mental Model

At a high level, data passing between Flyte tasks and workflows can be thought of as a pipeline. Each task produces outputs, which are passed as inputs to the next task in the sequence. However, Flyte handles data passing in a way that ensures separation between the control plane (which manages workflow execution) and the data plane (where raw data is stored).

There are two primary types of data in Flyte:

	1.	Metadata: Information about the data, such as the structure of the data (e.g., data types, paths to data storage).
	2.	Raw Data: The actual data itself (e.g., a Pandas DataFrame, a file in cloud storage).

### How Data is Handled in Workflow Execution

In-Memory vs. Persistent Data

	•	In-Memory Data: For small data (like int, float, bool, etc.), Flyte can pass data directly between tasks in memory.
	•	Persistent Data: For larger data (like files, datasets, or dataframes), Flyte does not pass the actual data between tasks. Instead, the data is stored in a persistent storage location (e.g., S3), and a reference (URI) to that data is passed between tasks.

Execution Context and Storage:

	•	When a task completes, Flyte checks the outputs and serializes the data. If the data is large, it is stored in cloud storage, and a reference to the location is passed to the next task.
	•	The next task retrieves the data using the URI and deserializes it back into its original form. This ensures that large data sets are not passed through memory, which could cause performance bottlenecks.

Error Handling During Data Passing:

	•	Flyte automatically handles errors related to data passing. If the output of one task doesn’t match the expected input type of the next task, Flyte raises an error, preventing invalid data from propagating through the workflow.

