use crate::actor_environment;
use crate::actor_environment::ActorEnvironment;
use futures::{SinkExt, StreamExt};
use pyo3::impl_::wrap::SomeWrap;
use pyo3::prelude::*;
use pyo3::pyclass::boolean_struct::False;
use pyo3_async_runtimes::TaskLocals;
use std::env;
use std::sync::Arc;
use tokio::net::TcpStream;
use tokio::sync::mpsc;
use tokio_util::codec::{Framed, LengthDelimitedCodec};
use tokio_util::task::TaskTracker;
use tracing::{debug, error, info};
use unionai_actor_bridge::common::{Response, Task, FAILED, SUCCEEDED};

pub async fn run(args: ExecutorArgs, locals: TaskLocals) -> Result<(), PyErr> {
    let stream = TcpStream::connect(args.executor_registration_addr).await?;
    let mut framed = Framed::new(stream, LengthDelimitedCodec::new());

    info!("[Actor Core] Creating a new controller...");
    let controller = match env::var("_UNION_EAGER_API_KEY") {
        Ok(value) => {
            debug!("_UNION_EAGER_API_KEY: {}", value);
            actor_environment::create_controller(None, false, value.wrap()).await?
        }
        Err(e) => {
            debug!("Couldn't read _UNION_EAGER_API_KEY: {}", e);
            actor_environment::create_controller(
                String::from("dns:///host.docker.internal:8090").wrap(),
                true,
                None,
            )
            .await?
        }
    };

    info!("[Actor core] Creating new actor environment...");
    let mut actor_environment = ActorEnvironment::new(controller, locals);

    loop {
        // retrieve next task
        let buf = match framed.next().await {
            Some(Ok(buf)) => buf,
            Some(Err(e)) => {
                error!("executor '{}' failed to read from socket: {}", args.id, e);
                break;
            }
            None => {
                error!("executor '{}' connection closed", args.id);
                break;
            }
        };

        let task: Task = bincode::deserialize(&buf).unwrap();
        debug!("executor {} received work: {:?}", args.id, task.cmd);

        let fut = actor_environment.run(task);

        let cmd_result = match fut.await {
            Ok(_) => {
                debug!("executor completed task successfully");
                Ok(())
            }
            Err(e) => {
                error!("executor failed to run task: {:?}", e);
                Err(format!("{:?}", e))
            }
        };

        // return result
        let response = match cmd_result {
            Ok(_) => Response {
                phase: SUCCEEDED,
                reason: None,
                executor_corrupt: false,
            },
            Err(e) => Response {
                phase: FAILED,
                reason: Some(e),
                executor_corrupt: false,
            },
        };

        let buf = bincode::serialize(&response).unwrap();
        if let Err(e) = framed.send(buf.into()).await {
            error!("executor '{}' failed to write to socket: {}", args.id, e);
            break;
        }
    }

    Ok(())
}

pub async fn run_worker_pool(args: ExecutorArgs, locals: TaskLocals) -> Result<(), PyErr> {
    let num_workers = args.num_workers.unwrap_or(1);

    info!("[Actor Core] Creating a new controller...");
    let controller = match env::var("_UNION_EAGER_API_KEY") {
        Ok(value) => {
            debug!("_UNION_EAGER_API_KEY: {}", value);
            actor_environment::create_controller(None, false, value.wrap()).await?
        }
        Err(e) => {
            debug!("Couldn't read _UNION_EAGER_API_KEY: {}", e);
            actor_environment::create_controller(
                String::from("dns:///host.docker.internal:8090").wrap(),
                true,
                None,
            )
            .await?
        }
    };

    info!("[Actor core] Creating new actor environment...");
    let actor_environment = Arc::new(ActorEnvironment::new(controller, locals));

    let tracker = TaskTracker::new();

    // Channels for coordinating work between TCP handler and workers
    let (work_tx, work_rx) = mpsc::channel::<Task>(100);
    let (response_tx, mut response_rx) = mpsc::channel::<Result<(), String>>(100);

    // Single TCP connection handler
    let work_rx = Arc::new(tokio::sync::Mutex::new(work_rx));
    tracker.spawn(async move {
        match TcpStream::connect(args.executor_registration_addr).await {
            Ok(stream) => {
                let mut framed = Framed::new(stream, LengthDelimitedCodec::new());

                loop {
                    tokio::select! {
                            // Read incoming data and distribute to workers
                            data = framed.next() => {
                                match data {
                                    Some(Ok(bytes)) => {
                                        let task: Task = bincode::deserialize(&bytes).unwrap();
                                        if work_tx.send(task).await.is_err() {
                                            break; // Workers are done
                                        }
                                    }
                                    Some(Err(e)) => {
                                        error!("TCP read error: {}", e);
                                        break;
                                    }
                                    None => {
                                        error!("TCP connection closed");
                                        break;
                                    }
                                }
                            }
                            // Send worker responses back over TCP
                            rx_response = response_rx.recv() => {
                            let response = match rx_response.unwrap() {
                                Ok(_) => Response {
                                    phase: SUCCEEDED,
                                    reason: None,
                                    executor_corrupt: false,
                                },
                                Err(e) => Response {
                                    phase: FAILED,
                                    reason: Some(e),
                                    executor_corrupt: false,
                                },
                            };

                            let buf = bincode::serialize(&response).unwrap();
                            if let Err(e) = framed.send(buf.into()).await {
                                error!("executor failed to write to socket: {}", e);
                                break;
                            }
                        }
                    }
                }
            }
            Err(e) => {
                error!("Failed to connect: {}", e);
            }
        }
    });

    // Spawn workers
    for i in 1..=num_workers {
        let work_rx = Arc::clone(&work_rx);
        let response_tx = response_tx.clone();
        let actor_environment = Arc::clone(&actor_environment);

        tracker.spawn(async move {
            while let Some(data) = {
                let mut rx = work_rx.lock().await;
                rx.recv().await
            } {
                let task = data;
                debug!("executor {} received work: {:?}", args.id, task.cmd);

                let fut = actor_environment.run(task);

                let response = match fut.await {
                    Ok(_) => Ok(()),
                    Err(e) => Err(format!("{:?}", e)),
                };

                // Send response
                if let Err(e) = response_tx.send(response.into()).await {
                    error!("Worker {} failed to send response: {}", i, e);
                    break;
                } else {
                    error!("Worker {} sent response", i);
                }
            }
            info!("Worker {} finished", i);
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
