# FastTask Server CLI

A gRPC server implementation of the FastTask service that reads tasks from a JSON file and distributes them to connected workers via heartbeat streams.

## Usage

```bash
# Build the server
go build -o fasttask-server .

# Run the server with default settings
./fasttask-server

# Run with custom port and tasks file
./fasttask-server -port 9090 -tasks /path/to/your/tasks.json
```

## Flags

- `-port`: gRPC server port (default: 8080)
- `-tasks`: Path to tasks JSON file (default: "testdata/tasks.json")

## Task JSON Format

The tasks file should contain an array of task objects:

```json
[
  {
    "task_id": "unique-task-id",
    "namespace": "namespace-name", 
    "command": ["command", "arg1", "arg2"],
    "env_vars": {
      "ENV_VAR_NAME": "value",
      "ANOTHER_VAR": "another_value"
    }
  }
]
```

## How It Works

1. Server loads tasks from the specified JSON file on startup
2. Tasks are queued internally for distribution 
3. When workers connect via the Heartbeat gRPC stream:
   - Server receives worker capacity information
   - Server assigns available tasks based on worker capacity
   - Server sends task assignments with ASSIGN operation
   - Server tracks assigned tasks per worker

## Features

- Reads task definitions from JSON file
- Implements FastTask gRPC service with Heartbeat streaming
- Distributes tasks based on worker capacity (execution_limit - execution_count + backlog slots)
- Tracks task assignments and worker status
- Logs worker connections and task assignments
- Supports environment variable injection for tasks

## Example Output

```
2025/08/04 16:48:18 Loaded 3 tasks from testdata/tasks.json
2025/08/04 16:48:18 FastTask server starting on port 8080
2025/08/04 16:48:18 Loading tasks from: testdata/tasks.json
2025/08/04 16:48:25 Client connected for heartbeat stream
2025/08/04 16:48:25 Received heartbeat from worker worker-001, queue default-queue
2025/08/04 16:48:25 Worker capacity: 0/5 active, 0/10 backlog
2025/08/04 16:48:25 Assigned task rcf7sw8frpsbz7gcmk9v-bswd03zvzi8892d19tyoab8eq-0 to worker worker-001
```