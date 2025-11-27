use crate::actor_environment;
use crate::actor_environment::ActorEnvironment;
use futures::{SinkExt, StreamExt};
use pyo3::impl_::wrap::SomeWrap;
use pyo3::prelude::*;
use pyo3_async_runtimes::TaskLocals;
use std::collections::HashMap;
use pyo3_async_runtimes::tokio::into_future;
use std::env;
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, Ordering};
use std::time::Duration;
use tempfile::TempDir;
use tokio::net::TcpStream;
use tokio::sync::{mpsc, RwLock};
use tokio_util::codec::{Framed, LengthDelimitedCodec};
use tokio_util::sync::CancellationToken;
use tokio_util::task::TaskTracker;
use tracing::{debug, error, info, info_span, warn, Instrument};
use tracing_appender::non_blocking::WorkerGuard;
use tracing_attributes::instrument;
use unionai_actor_bridge::common::{Response, Task, ABORTED, FAILED, SUCCEEDED};
use crate::TaskError;

// Store active task cancellation tokens
type CancellationRegistry = Arc<RwLock<HashMap<String, CancellationToken>>>;

async fn watch_for_errors(controller: &Py<PyAny>) -> PyResult<()> {
    let fut = Python::with_gil(|py| {
        let ctrl = controller.bind(py);
        let coro = ctrl.call_method0("watch_for_errors")?;
        into_future(coro)
    })?;
    fut.await?;
    Ok(())
}

#[instrument(skip(args, locals), fields(executor.id = args.id, num_workers = args.num_workers.unwrap_or(1)
))]
pub async fn run_worker_pool(
    args: ExecutorArgs,
    locals: TaskLocals,
    _: TempDir,
    _: WorkerGuard,
    _: WorkerGuard,
) -> Result<(), PyErr> {
    let num_workers = args.num_workers.unwrap_or(1);
    info!("Starting worker pool");

    // Create a shutdown flag that is checked by the workers
    let shutdown = Arc::new(AtomicBool::new(false));
    let worker_shutdown = Arc::clone(&shutdown);

    // Broadcast channel for system-wide shutdown (separate from user-initiated cancellations)
    let (shutdown_tx, _shutdown_rx) = tokio::sync::broadcast::channel::<String>(1);

    debug!("Creating controller for pool");
    let controller = match env::var("_UNION_EAGER_API_KEY") {
        Ok(value) => {
            debug!("Using Eager API key");
            actor_environment::create_controller(None, false, value.wrap()).await?
        }
        Err(_) => {
            debug!("Using local controller");
            actor_environment::create_controller(
                String::from("dns:///host.docker.internal:8090").wrap(),
                true,
                None,
            )
            .await?
        }
    };

    let controller_watcher = Python::with_gil(|py| {
        controller.clone_ref(py)
    });

    debug!("Creating actor environment");
    let actor_environment = Arc::new(ActorEnvironment::new(controller, locals));

    let tracker = TaskTracker::new();

    // Registry for active task cancellation tokens
    let cancellation_registry: CancellationRegistry = Arc::new(RwLock::new(HashMap::new()));

    // Channels for coordinating work between TCP handler and workers
    let (work_tx, work_rx) = mpsc::channel::<Task>(args.num_workers.unwrap_or(10));
    let (response_tx, mut response_rx) = mpsc::channel::<Response>(5 * args.num_workers.unwrap_or(50));

    // Single TCP connection handler
    let work_rx = Arc::new(tokio::sync::Mutex::new(work_rx));
    let tcp_addr = args.executor_registration_addr.clone();
    let registry_tcp = Arc::clone(&cancellation_registry);

    // Move tcp stream creation up here and split the stream into reader and writer.
    // This allows responses to be sent even if work_tx is full.

    tracker.spawn(async move {
        match TcpStream::connect(&tcp_addr).await {
            Ok(stream) => {
                info!(addr = %tcp_addr, "Connected to bridge");
                let mut framed = Framed::new(stream, LengthDelimitedCodec::new());

                // Need to fix this. This loop can end very badly.
                loop {
                    tokio::select! {
                        // Read incoming data and distribute to workers
                        data = framed.next() => {
                            match data {
                                Some(Ok(bytes)) => {
                                    let task: Task = bincode::deserialize(&bytes).unwrap();
 
                                    debug!(task.id = %task.unique_task_id, "Analyzing task for processing");

                                    // Take care of task aborts
                                    if task.cancel {
                                        let registry = registry_tcp.read().await;
                                        if let Some(token) = registry.get(&task.unique_task_id) {
                                            info!(task.id = %task.unique_task_id, "Task cancellation detected, cancelling task token");
                                            token.cancel();
                                        } else {
                                                warn!(task.id = %task.unique_task_id, "Task not found in registry for cancellation");
                                        }
                                        continue;
                                    }
 
                                    // This is a create task operation, first create a cancellation token for the task,
                                    // add it to the registry and then send the task to the worker channel
                                    let token = CancellationToken::new();
                                    // Create and add cancellation token to the registry, then drop the guard
                                    {
                                        let mut registry = registry_tcp.write().await;
                                        registry.insert(task.unique_task_id.clone(), token);
                                        debug!(task.id = %task.unique_task_id, "Registered task cancellation token");
                                    }
                                    debug!(task.id = %task.unique_task_id, "Distributing task to worker");

                                    // potentially check the atomic boolean here. don't send onto the work channel
                                    if work_tx.send(task).await.is_err() {
                                        error!("Work channel closed, TCP handler exiting");
                                        break;
                                    }
                                }
                                Some(Err(e)) => {
                                    error!(error = %e, "TCP read error");
                                    break;
                                }
                                None => {
                                    error!("TCP connection closed by bridge");
                                    break;
                                }
                            }
                        }
                        // Send worker responses back over TCP
                        rx_response_opt = response_rx.recv() => {
                            let rx_response = match rx_response_opt {
                                Some(response) => response,
                                None => {
                                    error!("Response channel closed, exiting TCP handler");
                                    break;
                                }
                            };
                            debug!(task.id = %rx_response.unique_task_id, phase = rx_response.phase, "Sending response to bridge");
                            let buf = bincode::serialize(&rx_response).unwrap();
                            if let Err(e) = framed.send(buf.into()).await {
                                error!(task.id = %rx_response.unique_task_id, error = %e, "Failed to send response to bridge");
                                break;
                            }
                        }
                    }
                }
            }
            Err(e) => {
                error!(addr = %tcp_addr, error = %e, "Failed to connect to bridge");
            }
        }
    });

    // Spawn workers
    info!(num_workers, "Spawning worker pool");
    for i in 1..=num_workers {
        let work_rx = Arc::clone(&work_rx);
        let response_tx = response_tx.clone();
        let actor_environment = Arc::clone(&actor_environment);
        let registry_worker_pool = Arc::clone(&cancellation_registry);
        let worker_shutdown_clone = Arc::clone(&worker_shutdown);
        let worker_shutdown_rx = shutdown_tx.subscribe();

        tracker.spawn(async move {
            info!(worker.id = i, "Worker started");
            while let Some(task) = {
                let mut rx = work_rx.lock().await;
                rx.recv().await
            } {
                // Before doing anything make sure we're not in a shutdown state
                if worker_shutdown_clone.load(Ordering::Acquire) {
                    warn!(worker.id = i, task.id = %task.unique_task_id, "Worker shutting down because shutdown flag set...");
                    // If in shutdown mode, shouldn't need to send back a response. The bridge should fail all non-terminal tasks.
                    break;
                }

                let task_id = task.unique_task_id.clone();

                let token = {
                    let registry = registry_worker_pool.read().await;
                    if let Some(token) = registry.get(&task_id) {
                        debug!(task.id = %task_id, "Found cancellation token for task");
                        // If the task is cancelled, we can skip processing
                        if token.is_cancelled() {
                            info!(task.id = %task_id, "Task was cancelled, skipping execution");
                            continue;
                        }
                        token.clone()
                    } else {
                        // todo: This should send back an aborted status.
                        warn!(task.id = %task_id, "No cancellation token found for task");
                        continue;
                    }
                };

                let task_execution = async {
                    debug!(cmd = ?task.cmd, "Processing task");
                    let result = actor_environment.run(task, token, worker_shutdown_rx.resubscribe()).await;
                    match &result {
                        Ok(_) => info!("Task completed successfully"),
                        Err(TaskError::Cancelled(msg)) => error!("Task aborted in run worker pool {}", msg),
                        Err(TaskError::CleanUpTimeoutPoisonPill(msg)) => error!("Task poison pill timeout detected {}", msg),
                        Err(TaskError::Python(e)) => error!(error = %e, "Task failed in run worker pool"),
                    }
                    result
                }
                .instrument(info_span!("worker_task", task.id = %task_id, worker.id = i));

                let response = match task_execution.await {
                    Ok(_) => Response {
                        phase: SUCCEEDED,
                        reason: None,
                        executor_corrupt: false,
                        unique_task_id: task_id.clone(),
                    },
                    Err(TaskError::Cancelled(msg)) => Response {
                        phase: ABORTED,
                        reason: Some(msg),
                        executor_corrupt: false,
                        unique_task_id: task_id.clone(),
                    },
                    Err(TaskError::Python(e)) => Response {
                        phase: FAILED,
                        reason: Some(e.to_string()),
                        executor_corrupt: false,
                        unique_task_id: task_id.clone(),
                    },
                    Err(TaskError::CleanUpTimeoutPoisonPill(msg)) => {
                        Response {
                            phase: FAILED,
                            reason: Some(msg),
                            executor_corrupt: true,
                            unique_task_id: task_id.clone(),
                        }
                    }
                };

                debug!(task.id = %task_id, phase = response.phase, "Sending response");
                if let Err(e) = response_tx.send(response).await {
                    error!(worker.id = i, error = %e, "Failed to send response");
                    break;
                }
            }
            warn!(worker.id = i, "Worker finished");
        });
    }

    tracker.close();

    tokio::select! {
        result = tracker.wait() => {
            info!(result = ?result, "Worker pool ended ");
        }
        watch_result = watch_for_errors(&controller_watcher) => {
            match watch_result {
                Ok(_) => {
                    error!("Watch for errors terminated but with no error, shutting down executor");
                }
                Err(e) => {
                    error!(error = %e, "Watch for errors terminated error");
                    // Once we return an error from here, the top level asyncio.run shuts down, which causes a CancelledError
                    // to be sent to all other running asyncio tasks (across any number of Runs)
                    // To prevent that, broadcast a shutdown message intentionally, which will trigger an explicit
                    // cancel from actor_environment.rs, which allows us to add a message to the cancel() call also.

                    // Set the shutdown flag to prevent new tasks from starting
                    shutdown.store(true, Ordering::Release);

                    // Broadcast shutdown to all running tasks
                    let shutdown_msg = format!("Controller error: {}", e);
                    let _ = shutdown_tx.send(shutdown_msg);

                    // Wait briefly for tasks to handle shutdown gracefully
                    tokio::time::sleep(Duration::from_millis(100)).await;

                    return Err(e);
                }
            }
        }
    }

    Ok(())
}

#[derive(Debug)]
pub struct ExecutorArgs {
    pub executor_registration_addr: String,
    pub id: usize,
    pub num_workers: Option<usize>,
}
