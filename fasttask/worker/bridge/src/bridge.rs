use std::collections::{HashMap, HashSet};
use std::future::Future;
use std::hash::{DefaultHasher, Hash, Hasher};
use std::pin::Pin;
use std::sync::{Arc, Mutex, RwLock};
use std::task::{Context, Poll, Waker};
use std::time::{Duration, SystemTime, UNIX_EPOCH};

use crate::cli::BridgeArgs;
use crate::common::TaskContext;
use crate::common::{Executor, FAILED, SUCCEEDED};
use crate::pb::fasttask::fast_task_client::FastTaskClient;
use crate::pb::fasttask::heartbeat_response::Operation;
use crate::pb::fasttask::{Capacity, HeartbeatRequest, TaskStatus};
use crate::task;

use tokio::net::TcpListener;
use tokio::process::Command;
use tokio::time::Interval;
use tokio_util::codec::{Framed, LengthDelimitedCodec};
use tonic::Request;
use tracing::{debug, error, warn};

struct AsyncBoolFuture {
    async_bool: Arc<Mutex<AsyncBool>>,
}

impl Future for AsyncBoolFuture {
    type Output = ();

    fn poll(self: Pin<&mut Self>, ctx: &mut Context<'_>) -> Poll<()> {
        let mut async_bool = self.async_bool.lock().unwrap();
        if async_bool.value {
            async_bool.value = false;
            return Poll::Ready(());
        }

        let waker = ctx.waker().clone();
        async_bool.waker = Some(waker);

        Poll::Pending
    }
}

struct AsyncBool {
    value: bool,
    waker: Option<Waker>,
}

impl AsyncBool {
    fn new() -> Self {
        Self {
            value: false,
            waker: None,
        }
    }

    fn trigger(&mut self) {
        self.value = true;
        if let Some(waker) = &self.waker {
            waker.clone().wake();
        }
    }
}

struct Heartbeater {
    interval: Interval,
    async_bool: Arc<Mutex<AsyncBool>>,
}

impl Heartbeater {
    async fn trigger(&mut self) -> () {
        let async_bool_future = AsyncBoolFuture {
            async_bool: self.async_bool.clone(),
        };

        tokio::select! {
            _ = self.interval.tick() => {},
            _ = async_bool_future => {},
        }
    }
}

pub async fn run(args: BridgeArgs) -> Result<(), Box<dyn std::error::Error>> {
    let executor_registration_addr = args.executor_registration_addr;
    let worker_id = args.worker_id;
    let (task_status_tx, task_status_rx) = async_channel::unbounded();
    let task_statuses: Arc<RwLock<Vec<TaskStatus>>> = Arc::new(RwLock::new(vec![]));
    let heartbeat_bool = Arc::new(Mutex::new(AsyncBool::new()));

    let (backlog_tx, backlog_rx) = async_channel::unbounded();

    // build executors
    let (build_executor_tx, build_executor_rx) = async_channel::unbounded();
    let (executor_tx, executor_rx) = async_channel::unbounded();

    let executor_tx_clone = executor_tx.clone();
    let listener = match TcpListener::bind(&executor_registration_addr).await {
        Ok(listener) => listener,
        Err(e) => {
            error!("failed to bind to port '{}'", e);
            std::process::exit(1);
        }
    };

    tokio::spawn(async move {
        let mut index = 0;
        loop {
            build_executor_rx.recv().await;

            // start child process
            let child = Command::new("unionai-actor-executor")
                .arg("--executor-registration-addr")
                .arg(executor_registration_addr.clone())
                .arg("--id")
                .arg(index.to_string())
                .spawn()
                .unwrap();

            let stream = listener.accept().await.unwrap().0;
            let framed = Framed::new(stream, LengthDelimitedCodec::new());

            let executor = Executor { framed, child };
            executor_tx_clone.send(executor).await;

            index += 1;
        }
    });

    for _ in 0..args.parallelism {
        build_executor_tx.send(()).await;
    }

    // spawn task status aggregator
    let (heartbeat_bool_clone, task_statuses_clone) =
        (heartbeat_bool.clone(), task_statuses.clone());
    tokio::spawn(async move {
        loop {
            let task_status_result = task_status_rx.recv().await;

            // append task_status to task_statuses
            let task_status: TaskStatus = task_status_result.unwrap();
            if task_status.phase == SUCCEEDED || task_status.phase == FAILED {
                // if task phase is terminal then trigger heartbeat immediately
                let mut heartbeat_bool = heartbeat_bool_clone.lock().unwrap();
                heartbeat_bool.trigger();
            }
            let mut task_statuses = task_statuses_clone.write().unwrap();
            task_statuses.push(task_status);
        }
    });

    let mut fast_register_ids = HashSet::new();
    loop {
        // initialize grpc client
        let mut client = match FastTaskClient::connect(args.fasttask_url.clone()).await {
            Ok(client) => client,
            Err(e) => {
                error!(
                    "failed to connect to grpc service '{}' '{:?}'",
                    args.fasttask_url, e
                );
                tokio::time::sleep(Duration::from_secs(1)).await;
                continue;
            }
        };

        // start heartbeater
        let (
            worker_id_clone,
            queue_id_clone,
            task_statuses_clone,
            heartbeat_bool_clone,
            heartbeat_interval_seconds,
            fast_register_dir_override,
        ) = (
            worker_id.clone(),
            args.queue_id.clone(),
            task_statuses.clone(),
            heartbeat_bool.clone(),
            args.heartbeat_interval_seconds,
            args.fast_register_dir_override.clone(),
        );
        let (executor_rx_clone, parallelism_clone, backlog_rx_clone, backlog_length_clone) = (
            executor_rx.clone(),
            args.parallelism as i32,
            backlog_rx.clone(),
            args.backlog_length as i32,
        );
        let outbound = async_stream::stream! {
            let mut heartbeater = Heartbeater {
                interval: tokio::time::interval(Duration::from_secs(heartbeat_interval_seconds)),
                async_bool: heartbeat_bool_clone.clone(),
            };

            loop {
                // periodically send heartbeat
                let _ = heartbeater.trigger().await;

                let mut heartbeat_request = HeartbeatRequest {
                    worker_id: worker_id_clone.clone(),
                    queue_id: queue_id_clone.clone(),
                    capacity: Some(Capacity {
                        execution_count: parallelism_clone - (executor_rx_clone.len() as i32),
                        execution_limit: parallelism_clone,
                        backlog_count: backlog_rx_clone.len() as i32,
                        backlog_limit: backlog_length_clone,
                    }),
                    task_statuses: vec!(),
                };

                {
                    let mut task_statuses = task_statuses_clone.write().unwrap();
                    heartbeat_request.task_statuses = task_statuses.clone();
                    task_statuses.clear();
                }

                debug!("sending heartbeat request '{:?}'", heartbeat_request);
                yield heartbeat_request;
            }
        };

        // handle heartbeat responses
        let response = match client.heartbeat(Request::new(outbound)).await {
            Ok(response) => response,
            Err(e) => {
                warn!("failed to send heartbeat '{:?}'", e);
                continue;
            }
        };

        let mut inbound = response.into_inner();

        let task_contexts = Arc::new(RwLock::new(HashMap::<String, TaskContext>::new()));
        //while let Some(heartbeat_response) = inbound.message().await? {
        loop {
            let heartbeat_response_result = inbound.message().await;
            if let Err(e) = heartbeat_response_result {
                warn!("failed to retrieve heartbeat response '{:?}'", e);
                break;
            }
            let heartbeat_response_option = heartbeat_response_result.unwrap();
            if let None = heartbeat_response_option {
                break;
            }
            let heartbeat_response = heartbeat_response_option.unwrap();
            debug!("sending heartbeat response = {:?}", heartbeat_response);

            match Operation::from_i32(heartbeat_response.operation) {
                Some(Operation::Assign) => {
                    // parse and update command
                    let mut cmd_str = heartbeat_response.cmd.clone();
                    let (
                        mut pyflyte_execute_index,
                        mut additional_distribution,
                        mut fast_register_dir,
                    ) = (None, None, None);
                    if cmd_str[0].eq("pyflyte-fast-execute") {
                        let mut dest_dir_index = None;
                        for i in 0..cmd_str.len() {
                            match cmd_str[i] {
                                ref x if x.eq("--dest-dir") => dest_dir_index = Some(i + 1),
                                ref x if x.eq("--additional-distribution") => {
                                    additional_distribution = Some(cmd_str[i + 1].clone())
                                }
                                ref x if x.eq("pyflyte-execute") => pyflyte_execute_index = Some(i),
                                ref x if x.eq("pyflyte-map-execute") => {
                                    pyflyte_execute_index = Some(i)
                                }
                                _ => (),
                            }
                        }

                        // hash `additional_distribution` (fast register unique id) to identify a
                        // unique subdirectory to decompress the fast register file
                        let mut h = DefaultHasher::new();
                        additional_distribution.hash(&mut h);
                        let dir = format!("{}/{}", fast_register_dir_override.clone(), h.finish());

                        if let Err(e) = std::fs::create_dir_all(dir.clone()) {
                            warn!("failed to create fast register subdir '{}'", dir);
                            continue;
                        }

                        match dest_dir_index {
                            Some(i) => cmd_str[i] = dir.clone(),
                            None => {} // TODO @hamersaw - inject `--dest-dir fast_register_dir` into cmd_str
                        }

                        fast_register_dir = Some(dir);
                    }

                    // if the fast register file has already been processed we skip the download
                    if let Some(ref additional_distribution_str) = additional_distribution {
                        if fast_register_ids.contains(additional_distribution_str) {
                            additional_distribution = None;
                        } else {
                            fast_register_ids.insert(additional_distribution_str.clone());
                        }
                    }

                    // strip `pyflyte-fast-execute` command if exists
                    let cmd = if let Some(pyflyte_execute_index) = pyflyte_execute_index {
                        cmd_str[pyflyte_execute_index..].to_vec()
                    } else {
                        cmd_str.clone()
                    };

                    // execute command
                    let (task_id, namespace, workflow_id, env_vars) = (
                        heartbeat_response.task_id.clone(),
                        heartbeat_response.namespace.clone(),
                        heartbeat_response.workflow_id.clone(),
                        heartbeat_response.env_vars.clone(),
                    );
                    let (
                        task_contexts_clone,
                        task_status_tx_clone,
                        task_status_report_interval_seconds,
                        last_ack_grace_period_seconds,
                    ) = (
                        task_contexts.clone(),
                        task_status_tx.clone(),
                        args.task_status_report_interval_seconds,
                        args.last_ack_grace_period_seconds,
                    );
                    let (
                        backlog_tx_clone,
                        backlog_rx_clone,
                        executor_tx_clone,
                        executor_rx_clone,
                        build_executor_tx_clone,
                    ) = (
                        backlog_tx.clone(),
                        backlog_rx.clone(),
                        executor_tx.clone(),
                        executor_rx.clone(),
                        build_executor_tx.clone(),
                    );
                    tokio::spawn(async move {
                        if let Err(e) = task::execute(
                            task_contexts_clone,
                            task_id,
                            namespace,
                            workflow_id,
                            cmd,
                            env_vars,
                            additional_distribution,
                            fast_register_dir,
                            task_status_tx_clone,
                            task_status_report_interval_seconds,
                            last_ack_grace_period_seconds,
                            backlog_tx_clone,
                            backlog_rx_clone,
                            executor_tx_clone,
                            executor_rx_clone,
                            build_executor_tx_clone,
                        )
                        .await
                        {
                            warn!("failed to execute task '{:?}'", e);
                        }
                    });
                }
                Some(Operation::Ack) => {
                    let mut task_contexts = task_contexts.write().unwrap();

                    // update last ack timestamp
                    if let Some(ref mut task_context) =
                        task_contexts.get_mut(&heartbeat_response.task_id)
                    {
                        task_context.last_ack_timestamp = SystemTime::now()
                            .duration_since(UNIX_EPOCH)
                            .unwrap()
                            .as_secs();
                    }
                }
                Some(Operation::Delete) => {
                    let mut task_contexts = task_contexts.write().unwrap();

                    // send kill signal
                    if let Some(ref mut task_context) =
                        task_contexts.get_mut(&heartbeat_response.task_id)
                    {
                        if let Err(e) = task_context.kill_tx.send(()).await {
                            warn!("failed to kill task '{:?}'", e);
                        }
                    }
                }
                None => warn!(
                    "unsupported heartbeat request operation '{:?}'",
                    heartbeat_response.operation
                ),
            }
        }
    }
}
