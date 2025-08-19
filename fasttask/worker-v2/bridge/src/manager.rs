use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use std::time::{Duration, SystemTime};

use anyhow::{anyhow, Result};
use async_channel::{Receiver, Sender, TryRecvError, TrySendError};
use bytes;
use futures::{
    stream::{SplitSink, SplitStream},
    SinkExt, StreamExt,
};
use tokio::net::{TcpListener, TcpStream};
use tokio::process::{Child, Command};

use tokio::time::interval;
use tokio_util::codec::{Framed, LengthDelimitedCodec};
use tracing::{debug, error, info, trace, warn};
use tracing_attributes::instrument;

use crate::common::{Task, ToProstDuration, FAILED, SUCCEEDED};
use crate::connection::{ConnectionBuilder, ConnectionRuntime};
use crate::pb::fasttask::{
    heartbeat_response::Operation, Capacity, ExecutionIdentifier, HeartbeatRequest,
    HeartbeatResponse, TaskStatus,
};

#[derive(Debug, Clone)]
struct TaskInfo {
    task_id: String,
    namespace: String,
    exec_id: ExecutionIdentifier,
    phase: i32,
    assigned_at: SystemTime,
    reason: Option<String>,
    env_vars: HashMap<String, String>,
    enqueue_labels: HashMap<String, String>,
    last_ack_time: Option<SystemTime>,
}

impl From<&TaskInfo> for TaskStatus {
    fn from(task_info: &TaskInfo) -> Self {
        TaskStatus {
            task_id: task_info.task_id.clone(),
            namespace: task_info.namespace.clone(),
            workflow_id: None,
            phase: task_info.phase,
            reason: task_info.reason.clone().unwrap_or_default(),
            exec_id: Some(task_info.exec_id.clone()),
            task_duration: task_info.last_ack_time.and_then(|at| {
                at.duration_since(task_info.assigned_at)
                    .map_or(None, |d| Some(d.to_prost()))
            }),
            enqueue_labels: task_info.enqueue_labels.clone(),
        }
    }
}

pub struct V2TaskManager {
    max_parallelism: i32,
    executor_registration_addr: String,
    worker_id: String,
    queue_id: String,
    heartbeat_interval_seconds: u64,

    // Task tracking
    tasks_in_progress: Arc<Mutex<HashMap<String, TaskInfo>>>,

    // Executor process and communication
    executor_process: Option<Child>,
    executor_stream_reader: Option<SplitStream<Framed<TcpStream, LengthDelimitedCodec>>>,
    executor_stream_writer:
        Option<SplitSink<Framed<TcpStream, LengthDelimitedCodec>, bytes::Bytes>>,

    // Channel for task assignment
    assignment_tx: Option<Sender<Task>>,
}

impl V2TaskManager {
    pub fn new(
        max_parallelism: i32,
        executor_registration_addr: &str,
        worker_id: String,
        queue_id: String,
        heartbeat_interval_seconds: u64,
    ) -> Self {
        Self {
            max_parallelism,
            executor_registration_addr: executor_registration_addr.to_string(),
            worker_id,
            queue_id,
            heartbeat_interval_seconds,
            tasks_in_progress: Arc::new(Mutex::new(HashMap::new())),
            executor_process: None,
            executor_stream_reader: None,
            executor_stream_writer: None,
            assignment_tx: None,
        }
    }

    #[instrument(skip(self), fields(worker.id = %self.worker_id, queue.id = %self.queue_id))]
    pub async fn start(&mut self) -> Result<()> {
        // Start TCP listener for executor connection
        let listener = TcpListener::bind(&self.executor_registration_addr).await?;
        info!(addr = %self.executor_registration_addr, "TCP listener started in bridge");

        // Start unionai-actor-executor process
        let child = Command::new("unionai-actor-executor")
            .arg("--executor-registration-addr")
            .arg(&self.executor_registration_addr)
            .arg("--id")
            .arg("0")
            .arg("--num-workers")
            .arg(self.max_parallelism.to_string())
            .spawn()?;

        let pid = child.id();
        info!(executor.pid = ?pid, num_workers = self.max_parallelism, "Started executor process");
        self.executor_process = Some(child);

        // Wait for executor to connect
        // ai: remove the first log but keep track of time.
        info!("Waiting for executor connection");
        let (stream, addr) = listener.accept().await?;
        info!(executor.addr = %addr, "Executor connected");

        // We split the communication with the executor into a reader and writer. The writer is used
        // to send tasks to it to run, but it may block for some reason. But even if the executor blocks,
        // it may still be sending updates back to us, so we need a separate coroutine to be reading.
        let framed = Framed::new(stream, LengthDelimitedCodec::new());
        let (framed_writer, framed_reader) = framed.split();
        self.executor_stream_reader = Some(framed_reader);
        self.executor_stream_writer = Some(framed_writer);

        Ok(())
    }

    #[instrument(skip(stream_writer, assignment_rx))]
    async fn read_and_send_assignments(
        mut stream_writer: SplitSink<Framed<TcpStream, LengthDelimitedCodec>, bytes::Bytes>,
        assignment_rx: Receiver<Task>,
    ) -> Result<()> {
        info!("Starting task assignment handler");
        loop {
            match assignment_rx.recv().await {
                Ok(task) => {
                    let task_id = &task.unique_task_id;
                    let serialized_task = bincode::serialize(&task)
                        .map_err(|e| anyhow!("Failed to serialize task {}: {}", task_id, e))?;
                    match stream_writer.send(serialized_task.into()).await {
                        Ok(()) => trace!(task.id = %task_id, "Sent task to executor"),
                        Err(e) => {
                            error!(task.id = %task_id, error = %e, "Failed to send task to executor");
                        }
                    }
                }
                Err(e) => {
                    error!(error = ?e, "Assignment channel closed");
                    return Err(anyhow!("Assignment channel closed: {}", e));
                }
            }
        }
    }

    #[instrument(skip(self, connection_builder), fields(worker.id = %self.worker_id, queue.id = %self.queue_id
    ))]
    pub async fn run<T: ConnectionBuilder>(
        &mut self,
        connection_builder: T,
        cancellation_token: tokio_util::sync::CancellationToken,
    ) -> Result<()> {
        info!("Starting main event loop");

        let tasks_in_progress = Arc::clone(&self.tasks_in_progress);

        // This is an unbounded channel that the heartbeat timer, and the handler for the messages coming from
        // the executor, will add messages to. Basically both periodically, and when this bridge receives any
        // task status changes from the executor, we should fire off a heartbeat request.
        // But this needs to be debounced (which is done below).
        // todo: this should probably be a Notify.
        let (trigger_heartbeater_tx, trigger_heartbeater_rx) = async_channel::unbounded::<()>();

        // Spawn task to handle executor responses
        let mut executor_handler = {
            let mut executor_stream_reader = self
                .executor_stream_reader
                .take()
                .ok_or_else(|| anyhow!("Executor stream not available"))?;
            let tasks_in_progress = Arc::clone(&tasks_in_progress);
            let trigger_heartbeat_from_executor_tx = trigger_heartbeater_tx.clone();

            tokio::spawn(async move {
                Self::handle_executor_responses(
                    &mut executor_stream_reader,
                    tasks_in_progress,
                    trigger_heartbeat_from_executor_tx,
                )
                .await
            })
        };

        // Start heartbeat and work assignment loop
        let mut heartbeat_timer = interval(Duration::from_secs(self.heartbeat_interval_seconds));
        let connection_runtime = connection_builder.get_runtime()?;

        // These two channels related to heartbeating.
        // This channel will be used to connect the timer to the sending of the heartbeat request.
        let (heartbeat_tx, heartbeat_rx) = async_channel::bounded(5);
        // This channel is used to handle heartbeat responses.
        let (operation_tx, operation_rx) = async_channel::bounded(self.max_parallelism as usize);

        // Spawn connection handler
        let connection_cancellation_token = cancellation_token.clone();
        let mut connection_handler = tokio::spawn(async move {
            connection_runtime
                .run(heartbeat_rx, operation_tx, connection_cancellation_token)
                .await
        });

        // This channel is used when an assign operation is received, and is used to actually send the Task down the
        // TCP stream. This channel exists because the sending of the Task to the TCP stream can be blocking.
        let (assignment_tx, assignment_rx) = async_channel::bounded(self.max_parallelism as usize);
        // Spawn a task to pull from the assignment channel and actually pass into the TCP stream
        self.assignment_tx = Some(assignment_tx);
        let stream_writer = self
            .executor_stream_writer
            .take()
            .ok_or_else(|| anyhow!("Executor stream writer not available"))?;
        let mut assignment_handler = tokio::spawn(async move {
            Self::read_and_send_assignments(stream_writer, assignment_rx).await
        });

        let trigger_heartbeat_from_timer_tx = trigger_heartbeater_tx.clone();
        // Main event loop
        let loop_result = loop {
            tokio::select! {
                _ = heartbeat_timer.tick() => {
                    // On timer... don't actually send the heartbeat, just notify
                    trigger_heartbeat_from_timer_tx.send(()).await.unwrap_or_else(|e| {
                        error!(error = %e, "Failed to send heartbeat tick");
                    });
                }
                _ = trigger_heartbeater_rx.recv() => {
                    // consume until empty as a crude debounce mechanism
                    loop {
                        match trigger_heartbeater_rx.try_recv() {
                            Ok(_) => {
                                // Process the message (in this case just a unit type)
                                // Do whatever you need to do with each message
                            }
                            Err(TryRecvError::Empty) => {
                                // Channel is empty, break out of loop
                                trace!("Heartbeat channel is empty, breaking out of inner loop");
                                break;
                            }
                            Err(TryRecvError::Closed) => {
                                // Channel is closed, break out of loop
                                debug!("Heartbeat channel is empty, breaking out of inner loop");
                                break;
                            }
                        }
                    }
                    let heartbeat_request = self.create_heartbeat_request();
                    debug!(
                        running_tasks = heartbeat_request.capacity.as_ref().map(|c| c.execution_count).unwrap_or(0),
                        "Sending heartbeat"
                    );
                    if let Err(e) = heartbeat_tx.send(heartbeat_request).await {
                        error!(error = %e, "Failed to send heartbeat");
                    }
                }

                // Handle HeartbeatResponse messages from the fast task service
                Ok(operation) = operation_rx.recv() => {
                    if let Err(e) = self.handle_operation(operation).await {
                        warn!(error = %e, "Failed to handle operation");
                    }
                }

                // Handle executor process exit
                result = &mut executor_handler => {
                    match result {
                        Ok(Ok(())) => {
                            warn!("Executor handler completed normally");
                        }
                        Ok(Err(e)) => {
                            error!("Executor handler failed: {}", e);
                        }
                        Err(e) => {
                            error!("Executor handler panicked: {}", e);
                        }
                    }
                    break Err(anyhow!("Executor process failed, terminating"));
                }

                // Handle connection failures
                result = &mut connection_handler => {
                    match result {
                        Ok(Ok(())) => {
                            warn!("Connection handler completed normally");
                        }
                        Ok(Err(e)) => {
                            error!("Connection handler failed: {}", e);
                        }
                        Err(e) => {
                            error!("Connection handler panicked: {}", e);
                        }
                    }
                    // Connection failures should be retried, but for simplicity we'll exit
                    break Err(anyhow!("Connection failed, terminating"));
                }

                // Handle cancellation signal
                _ = cancellation_token.cancelled() => {
                    info!("Received cancellation signal, shutting down gracefully");
                    break Ok(());
                }

                // Handle assignment worker failures
                result = &mut assignment_handler => {
                    match result {
                        Ok(Ok(())) => {
                            // Assignment handler should never exit as long as the process is running
                            error!("Assignment handler exited but without error");
                        }
                        Ok(Err(e)) => {
                            error!("Assignment handler failed: {}", e);
                        }
                        Err(e) => {
                            error!("Assignment handler panicked: {}", e);
                        }
                    }
                    // Exit for now, can revisit if we move to sidecar pattern.
                    break Err(anyhow!("Assignment handler failed, terminating"));
                }

                // Handle child executor process errors
                child_result = self.executor_process.as_mut().unwrap().wait() => {
                    match child_result {
                        Ok(status) if status.success() => {
                            error!("Executor process exited but is not supposed to.");
                        }
                        Ok(status) => {
                            error!("Executor process exited with status: {}", status);
                        }
                        Err(e) => {
                            error!("Failed to wait for executor process: {}", e);
                        }
                    }
                    break Err(anyhow!("Executor process terminated, exiting"));
                }
            }
        };

        // If the loop exits with an error,
        // abort connection handler and assignment handler and executor handler.
        let handles = [
            &mut executor_handler,
            &mut connection_handler,
            &mut assignment_handler,
        ];
        for h in handles {
            h.abort();
        }

        if let Err(e) = self.shutdown_executor().await {
            error!(error = %e, "Failed to shutdown executor gracefully");
        }

        match loop_result {
            Ok(()) => {
                info!("Task manager run exiting without error");
                Ok(())
            }
            Err(e) => {
                error!(error = %e, "Task manager run failed");
                // Mark unfinished tasks as failed with the error as the reason
                let reason = e.to_string();
                self.mark_unfinished_tasks_failed(&reason).await;
                Err(e)
            }
        }
    }

    #[instrument(skip(executor_stream, tasks_in_progress))]
    async fn handle_executor_responses(
        executor_stream: &mut SplitStream<Framed<TcpStream, LengthDelimitedCodec>>,
        tasks_in_progress: Arc<Mutex<HashMap<String, TaskInfo>>>,
        trigger_heartbeat_tx: Sender<()>,
    ) -> Result<()> {
        info!("Starting executor response handler");
        while let Some(response) = executor_stream.next().await {
            match response {
                Ok(bytes) => {
                    // Try to deserialize as Response first (task completion)
                    if let Ok(response) = bincode::deserialize::<crate::common::Response>(&bytes) {
                        let task_updated = {
                            let mut tasks = tasks_in_progress
                                .lock()
                                .map_err(|e| anyhow!("Failed to acquire lock: {}", e))?;
                            // if the task is in the map, then update its status
                            if let Some(task_info) = tasks.get_mut(&response.unique_task_id) {
                                // Update task status
                                trace!(
                                    task.id = %response.unique_task_id,
                                    old_phase = task_info.phase,
                                    new_phase = response.phase,
                                    "Updating task phase"
                                );
                                task_info.phase = response.phase;
                                task_info.reason = response.reason.clone();
                                if response.phase == SUCCEEDED {
                                    info!(task.id = %response.unique_task_id, "Task completed successfully");
                                } else if response.phase == FAILED || response.phase == 8 {
                                    warn!(
                                        task.id = %response.unique_task_id,
                                        phase = response.phase,
                                        reason = ?response.reason,
                                        "Task failed"
                                    );
                                }
                                true
                            } else {
                                warn!(task.id = %response.unique_task_id, "Received response for unknown task");
                                false
                            }
                        };

                        if task_updated {
                            trigger_heartbeat_tx.send(()).await.unwrap_or_else(|e| {
                                error!(error = %e, "Failed to trigger heartbeat after task update");
                            });
                            warn!(task.id = %response.unique_task_id, "removeme - sent trigger heartbeat tx");
                        }
                    } else {
                        error!(
                            bytes_len = bytes.len(),
                            "Failed to deserialize executor response"
                        );
                    }
                }
                Err(e) => {
                    error!(error = %e, "Error reading from executor");
                    return Err(e.into());
                }
            }
        }
        Ok(())
    }

    fn create_heartbeat_request(&self) -> HeartbeatRequest {
        let tasks = self.tasks_in_progress.lock().unwrap();
        let running_count = tasks
            .values()
            .filter(|task_info| {
                task_info.phase != FAILED && task_info.phase != 8 && task_info.phase != SUCCEEDED
            })
            .count();
        let task_statuses: Vec<TaskStatus> =
            tasks.values().map(|task_info| task_info.into()).collect();

        HeartbeatRequest {
            worker_id: self.worker_id.clone(),
            queue_id: self.queue_id.clone(),
            capacity: Some(Capacity {
                execution_count: running_count as i32,
                execution_limit: self.max_parallelism,
                backlog_count: 0,
                backlog_limit: 0,
            }),
            task_statuses,
        }
    }

    #[instrument(skip(self, hb_response), fields(operation = hb_response.operation, task.id = %hb_response.task_id
    ))]
    async fn handle_operation(&mut self, hb_response: HeartbeatResponse) -> Result<()> {
        match Operation::from_i32(hb_response.operation) {
            Some(Operation::Assign) => {
                trace!("Received task assignment");
                self.try_assign_task(hb_response).await
            }
            Some(Operation::Ack) => {
                trace!("Received task ACK");
                self.update_ack_time(hb_response).await
            }
            Some(Operation::Delete) => {
                trace!("Received task deletion request");
                self.delete_task(hb_response.task_id).await
            }
            None => {
                warn!(operation_code = hb_response.operation, "Unknown operation");
                Ok(())
            }
        }
    }

    async fn update_ack_time(&self, operation: HeartbeatResponse) -> Result<()> {
        let mut tasks = self
            .tasks_in_progress
            .lock()
            .map_err(|e| anyhow!("Failed to acquire lock: {}", e))?;
        if let Some(task_info) = tasks.get_mut(&operation.task_id) {
            task_info.last_ack_time = Some(SystemTime::now());
            trace!(task.id = %operation.task_id, "Updated ACK time for task");
        } else {
            warn!(task.id = %operation.task_id, "ACK received for unknown task");
        }
        Ok(())
    }

    async fn try_assign_task(&mut self, operation: HeartbeatResponse) -> Result<()> {
        let mut tasks = self
            .tasks_in_progress
            .lock()
            .map_err(|e| anyhow!("Failed to acquire lock: {}", e))?;
        let tx = self
            .assignment_tx
            .as_ref()
            .ok_or_else(|| anyhow!("Assignment channel not available"))?;

        let running_count = tasks
            .values()
            .filter(|task_info| {
                task_info.phase != FAILED && task_info.phase != 8 && task_info.phase != SUCCEEDED
            })
            .count();
        if running_count >= self.max_parallelism as usize {
            warn!(
                running_count,
                max_parallelism = self.max_parallelism,
                "Max parallelism reached"
            );
            return Err(anyhow!("Max parallelism reached"));
        }
        let exec_id = operation
            .exec_id
            .clone()
            .ok_or_else(|| anyhow!("Execution ID is required for task assignment"))?;

        // Create the initial task info to track
        let task_info = TaskInfo {
            task_id: operation.task_id.clone(),
            phase: 0,
            namespace: operation.namespace.clone(),
            exec_id: exec_id.clone(),
            assigned_at: SystemTime::now(),
            reason: None,
            env_vars: operation.env_vars.clone(),
            enqueue_labels: operation.enqueue_labels.clone(),
            last_ack_time: None,
        };

        // Create the initial task to send to the executor
        let _task = Task {
            cmd: operation.cmd,
            additional_distribution: None, // ketan: confirm that these are no longer used
            fast_register_dir: None,
            env_vars: Some(operation.env_vars),
            unique_task_id: operation.task_id.clone(),
        };

        debug!(exec.name = %exec_id.name, "Submitting task to executor");
        // try_send because we're holding the lock, if the channel is full, fail immediately
        match tx.try_send(_task) {
            Ok(()) => {
                tasks.insert(operation.task_id.clone(), task_info);
                debug!(
                    running_count = running_count + 1,
                    "Task assigned to executor"
                );
            }
            Err(TrySendError::Full(_)) => {
                error!("Assignment channel is full");
                return Err(anyhow!("Failed to send task assignment - channel full"));
            }
            Err(e) => {
                error!(error = %e, "Failed to send task assignment");
                return Err(anyhow!("Failed to send task assignment"));
            }
        }

        Ok(())
    }

    #[instrument(skip(self), fields(task.id = %task_id))]
    async fn delete_task(&mut self, task_id: String) -> Result<()> {
        let mut tasks = self.tasks_in_progress.lock().unwrap();

        // todo: check that the task is done before deleting it. if not done, we'll need to
        //       send a cancelled error to the executor in the future.
        if let Some(task_info) = tasks.remove(&task_id) {
            info!(phase = task_info.phase, "Task deleted");
        } else {
            warn!("Task not found for deletion");
        }

        Ok(())
    }

    async fn mark_unfinished_tasks_failed(&mut self, reason: &str) {
        let mut tasks = self.tasks_in_progress.lock().unwrap();
        let task_count = tasks.len();
        info!(task_count, reason, "Marking all tasks as failed");

        for (k, t_info) in tasks.iter_mut() {
            t_info.last_ack_time = Some(SystemTime::now());
            if is_terminal(t_info.phase) {
                trace!(task.id = %k, phase = t_info.phase, "Task is already terminal, skipping");
                continue; // Skip already terminal tasks
            }
            trace!(task.id = %k, old_phase = t_info.phase, "Marking task as failed");
            t_info.phase = FAILED;
        }
    }

    /// Send a final heartbeat with a new gRPC connection
    pub async fn send_final_heartbeat(&mut self, fasttask_url: String) {
        use crate::pb::fasttask::fast_task_client::FastTaskClient;
        use tonic::Request;

        info!("Sending final heartbeat before shutdown");

        // Create final heartbeat request with current task states
        let heartbeat_request = self.create_heartbeat_request();

        // Attempt to create a new gRPC connection for the final heartbeat
        match FastTaskClient::connect(fasttask_url.clone()).await {
            Ok(mut client) => {
                // Create a single heartbeat stream
                let outbound = async_stream::stream! {
                    yield heartbeat_request;
                };

                match client.heartbeat(Request::new(outbound)).await {
                    Ok(_response) => {
                        info!("Final heartbeat sent successfully");
                        // We don't need to process the response since we're shutting down
                    }
                    Err(e) => {
                        error!(error = %e, "Failed to send final heartbeat");
                    }
                }
            }
            Err(e) => {
                error!(error = %e, fasttask.url = %fasttask_url, "Failed to connect to FastTask service for final heartbeat");
            }
        }
    }

    pub async fn shutdown_executor(&mut self) -> Result<()> {
        if let Some(mut child) = self.executor_process.take() {
            match child.start_kill() {
                Ok(()) => {
                    info!(executor.pid = ?child.id(), "Initiated kill of executor process");
                    Ok(())
                }
                Err(e) => {
                    error!(error = %e, executor.pid = ?child.id(), "Failed to kill executor process");
                    Err(anyhow!("Killing of executor failed: {}", e))
                }
            }
        } else {
            debug!("Shutdown of executor received but process already gone");
            Ok(())
        }
    }
}

fn is_terminal(phase: i32) -> bool {
    // 8 permanent failure
    phase == FAILED || phase == 8 || phase == SUCCEEDED
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::common::RUNNING;
    use crate::pb::fasttask::{heartbeat_response, ExecutionIdentifier};
    use std::time::Duration;

    fn create_test_task_info() -> TaskInfo {
        TaskInfo {
            task_id: "test-task-123".to_string(),
            namespace: "test-namespace".to_string(),
            exec_id: ExecutionIdentifier {
                org: "test-org".to_string(),
                domain: "test-domain".to_string(),
                name: "test-execution".to_string(),
                project: "test-project".to_string(),
            },
            phase: RUNNING,
            assigned_at: SystemTime::now() - Duration::from_secs(10),
            reason: None,
            env_vars: HashMap::new(),
            enqueue_labels: HashMap::new(),
            last_ack_time: None,
        }
    }

    fn create_test_heartbeat_response() -> HeartbeatResponse {
        HeartbeatResponse {
            task_id: "test-task-123".to_string(),
            namespace: "test-namespace".to_string(),
            workflow_id: None,
            cmd: vec!["python".to_string(), "test.py".to_string()],
            operation: heartbeat_response::Operation::Ack as i32,
            env_vars: HashMap::new(),
            enqueue_labels: HashMap::new(),
            exec_id: Some(ExecutionIdentifier {
                org: "test-org".to_string(),
                domain: "test-domain".to_string(),
                name: "test-execution".to_string(),
                project: "test-project".to_string(),
            }),
        }
    }

    #[tokio::test]
    async fn test_update_ack_time_existing_task() {
        let tasks = Arc::new(Mutex::new(HashMap::new()));
        let task_info = create_test_task_info();
        let task_id = task_info.task_id.clone();

        // Insert task without ack time
        assert!(task_info.last_ack_time.is_none());
        tasks.lock().unwrap().insert(task_id.clone(), task_info);

        let manager = V2TaskManager {
            max_parallelism: 10,
            executor_registration_addr: "test-addr".to_string(),
            worker_id: "test-worker".to_string(),
            queue_id: "test-queue".to_string(),
            heartbeat_interval_seconds: 30,
            tasks_in_progress: tasks.clone(),
            executor_process: None,
            executor_stream_reader: None,
            executor_stream_writer: None,
            assignment_tx: None,
        };

        let hb_response = create_test_heartbeat_response();
        let result = manager.update_ack_time(hb_response).await;

        assert!(result.is_ok());

        let tasks_guard = tasks.lock().unwrap();
        let updated_task = tasks_guard.get(&task_id).unwrap();
        assert!(updated_task.last_ack_time.is_some());
    }

    #[tokio::test]
    async fn test_update_ack_time_nonexistent_task() {
        let tasks = Arc::new(Mutex::new(HashMap::new()));
        let manager = V2TaskManager {
            max_parallelism: 10,
            executor_registration_addr: "test-addr".to_string(),
            worker_id: "test-worker".to_string(),
            queue_id: "test-queue".to_string(),
            heartbeat_interval_seconds: 30,
            tasks_in_progress: tasks.clone(),
            executor_process: None,
            executor_stream_reader: None,
            executor_stream_writer: None,
            assignment_tx: None,
        };

        let hb_response = create_test_heartbeat_response();
        let result = manager.update_ack_time(hb_response).await;

        // Should succeed even if task doesn't exist (just logs a warning)
        assert!(result.is_ok());
        assert!(tasks.lock().unwrap().is_empty());
    }

    #[test]
    fn test_taskinfo_to_taskstatus_with_ack_time() {
        let mut task_info = create_test_task_info();
        let ack_time = SystemTime::now();
        task_info.last_ack_time = Some(ack_time);

        let task_status = TaskStatus::from(&task_info);

        assert_eq!(task_status.task_id, task_info.task_id);
        assert_eq!(task_status.namespace, task_info.namespace);
        assert_eq!(task_status.phase, task_info.phase);
        assert_eq!(task_status.exec_id, Some(task_info.exec_id));
        assert!(task_status.task_duration.is_some());

        let duration = task_status.task_duration.unwrap();
        assert!(duration.seconds >= 0);
        assert!(duration.nanos >= 0);
    }
}
