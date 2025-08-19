use crate::actor_environment;
use crate::actor_environment::ActorEnvironment;
use futures::{SinkExt, StreamExt};
use pyo3::impl_::wrap::SomeWrap;
use pyo3::prelude::*;
use pyo3_async_runtimes::TaskLocals;
use std::env;
use std::sync::Arc;
use tempfile::TempDir;
use tokio::net::TcpStream;
use tokio::sync::mpsc;
use tokio_util::codec::{Framed, LengthDelimitedCodec};
use tokio_util::task::TaskTracker;
use tracing::{debug, error, info, info_span, Instrument};
use tracing_appender::non_blocking::WorkerGuard;
use tracing_attributes::instrument;
use unionai_actor_bridge::common::{Response, Task, FAILED, SUCCEEDED};

#[instrument(skip(args, locals), fields(executor.id = args.id, addr = %args.executor_registration_addr
))]
pub async fn run(args: ExecutorArgs, locals: TaskLocals) -> Result<(), PyErr> {
    info!("Starting single executor");
    let stream = TcpStream::connect(&args.executor_registration_addr).await?;
    let mut framed = Framed::new(stream, LengthDelimitedCodec::new());
    info!("Connected to bridge");

    debug!("Creating single run controller");
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

    debug!("Creating actor environment");
    let actor_environment = ActorEnvironment::new(controller, locals);

    info!("Starting task processing loop");
    loop {
        // retrieve next task
        let buf = match framed.next().await {
            Some(Ok(buf)) => buf,
            Some(Err(e)) => {
                error!(error = %e, "Failed to read from socket");
                break;
            }
            None => {
                error!("Connection closed by bridge");
                break;
            }
        };

        // We can either deserialize a Task or a task_id string.
        // If its a task_id then we need to abort the task.
        if buf.is_empty() {
            error!("Received empty task");
            continue;
        }
        let task: Task = bincode::deserialize(&buf).unwrap();
        let task_id = task.unique_task_id.clone();

        let task_execution = async {
            debug!(cmd = ?task.cmd, "Received task");
            let result = actor_environment.run(task).await;
            match &result {
                Ok(_) => debug!("Task completed successfully"),
                Err(e) => error!(error = %e, "Task failed"),
            }
            result
        }
        .instrument(info_span!("task", task.id = %task_id));

        let cmd_result = match task_execution.await {
            Ok(_) => Ok(()),
            Err(e) => Err(format!("{:?}", e)),
        };

        // return result
        let response = match cmd_result {
            Ok(_) => Response {
                phase: SUCCEEDED,
                reason: None,
                executor_corrupt: false,
                unique_task_id: task_id.clone(),
            },
            Err(e) => Response {
                phase: FAILED,
                reason: Some(e),
                executor_corrupt: false,
                unique_task_id: task_id.clone(),
            },
        };

        let buf = bincode::serialize(&response).unwrap();
        if let Err(e) = framed.send(buf.into()).await {
            error!(task.id = %task_id, error = %e, "Failed to send response to bridge");
            break;
        }
    }

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

    debug!("Creating actor environment");
    let actor_environment = Arc::new(ActorEnvironment::new(controller, locals));

    let tracker = TaskTracker::new();

    // Channels for coordinating work between TCP handler and workers
    let (work_tx, work_rx) = mpsc::channel::<Task>(args.num_workers.unwrap_or(10));
    let (response_tx, mut response_rx) = mpsc::channel::<Response>(100);

    // Single TCP connection handler
    let work_rx = Arc::new(tokio::sync::Mutex::new(work_rx));
    let tcp_addr = args.executor_registration_addr.clone();
    tracker.spawn(async move {
        match TcpStream::connect(&tcp_addr).await {
            Ok(stream) => {
                info!(addr = %tcp_addr, "Connected to bridge");
                let mut framed = Framed::new(stream, LengthDelimitedCodec::new());

                loop {
                    tokio::select! {
                        // Read incoming data and distribute to workers
                        data = framed.next() => {
                            match data {
                                Some(Ok(bytes)) => {
                                    let task: Task = bincode::deserialize(&bytes).unwrap();
                                    debug!(task.id = %task.unique_task_id, "Distributing task to worker");

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

        tracker.spawn(async move {
            info!(worker.id = i, "Worker started");
            while let Some(task) = {
                let mut rx = work_rx.lock().await;
                rx.recv().await
            } {
                let task_id = task.unique_task_id.clone();

                let task_execution = async {
                    debug!(cmd = ?task.cmd, "Processing task");
                    let result = actor_environment.run(task).await;
                    match &result {
                        Ok(_) => info!("Task completed successfully"),
                        Err(e) => error!(error = %e, "Task failed"),
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
                    Err(e) => Response {
                        phase: FAILED,
                        reason: Some(e.to_string()),
                        executor_corrupt: false,
                        unique_task_id: task_id.clone(),
                    },
                };

                debug!(task.id = %task_id, phase = response.phase, "Sending response");
                if let Err(e) = response_tx.send(response).await {
                    error!(worker.id = i, error = %e, "Failed to send response");
                    break;
                }
            }
            info!(worker.id = i, "Worker finished");
        });
    }

    tracker.close();
    tracker.wait().await;

    Ok(())
}

#[derive(Debug)]
pub struct ExecutorArgs {
    pub executor_registration_addr: String,
    pub id: usize,
    pub num_workers: Option<usize>,
}
