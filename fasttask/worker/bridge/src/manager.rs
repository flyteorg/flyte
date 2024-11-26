use std::collections::{HashMap, HashSet};
use std::future::{Future, IntoFuture};
use std::hash::{DefaultHasher, Hash, Hasher};
use std::sync::{Arc, Mutex, RwLock};
use std::time::{SystemTime, UNIX_EPOCH};

use anyhow::{bail, Result};
use async_channel::{Receiver, Sender};
use tokio::net::TcpListener;
use tokio::process::Command;
use tokio_util::codec::{Framed, LengthDelimitedCodec};
use tracing::warn;

use crate::common::{Executor, TaskContext, SUCCEEDED};
use crate::pb::fasttask::{Capacity, TaskStatus};
use crate::task::{self};

pub trait TaskManager {
    fn ack(&self, task_id: String) -> impl Future<Output = Result<()>>;
    fn assign(
        &self,
        task_id: String,
        namespace: String,
        workflow_id: String,
        cmd: Vec<String>,
        env_vars: HashMap<String, String>,
    ) -> impl Future<Output = Result<()>>;
    fn delete(&self, task_id: String) -> impl Future<Output = Result<()>>;
    fn get_capacity(&self) -> Result<Capacity>;
    fn get_runtime<'a>(&self) -> Result<impl TaskManagerRuntime + 'a>;
}

#[trait_variant::make(TaskManagerRuntime: Send)]
pub trait LocalTaskManagerRuntime {
    async fn run(&self, task_status_tx: Sender<TaskStatus>) -> Result<()>;
}

pub struct CapacityReporter {
    capacity_trigger_tx: Sender<()>,
    capacity_report_rx: Receiver<Capacity>,
}

impl CapacityReporter {
    pub fn new(
        capacity_trigger_tx: Sender<()>,
        capacity_report_rx: Receiver<Capacity>,
    ) -> CapacityReporter {
        CapacityReporter {
            capacity_trigger_tx,
            capacity_report_rx,
        }
    }

    pub async fn get(&self) -> Result<Capacity> {
        self.capacity_trigger_tx.send(()).await?;
        let result = self.capacity_report_rx.recv().await?;
        Ok(result)
    }
}

#[derive(Clone)]
pub enum ExecutionStrategy {
    SuccessStrategy,
    TheExecutionStrategy,
}

impl ExecutionStrategy {
    async fn execute(
        &self,
        task_contexts: Arc<RwLock<HashMap<String, TaskContext>>>,
        task_assignment: TaskAssignment,
        task_status_tx: Sender<TaskStatus>,
        task_status_report_interval_seconds: u64,
        last_ack_grace_period_seconds: u64,
        backlog_tx: Sender<()>,
        backlog_rx: Receiver<()>,
        executor_tx: Sender<Executor>,
        executor_rx: Receiver<Executor>,
        build_executor_tx: Sender<()>,
    ) -> Result<()> {
        // initialize task context
        let (kill_tx, kill_rx) = async_channel::bounded(1);
        {
            let mut task_contexts = task_contexts.write().unwrap();
            task_contexts.insert(
                task_assignment.task_id.clone(),
                TaskContext {
                    kill_tx,
                    last_ack_timestamp: SystemTime::now()
                        .duration_since(UNIX_EPOCH)
                        .unwrap()
                        .as_secs(),
                },
            );
        }

        match self {
            ExecutionStrategy::SuccessStrategy => {
                task_status_tx
                    .send(TaskStatus {
                        task_id: task_assignment.task_id.clone(),
                        namespace: task_assignment.namespace,
                        workflow_id: task_assignment.workflow_id,
                        phase: SUCCEEDED,
                        reason: "reason".to_string(),
                    })
                    .await?;

                // wait for kill_rx on task_context
                kill_rx.recv().await?;
            }
            ExecutionStrategy::TheExecutionStrategy => {
                task::execute(
                    kill_rx,
                    task_contexts.clone(),
                    task_assignment.task_id.clone(),
                    task_assignment.namespace,
                    task_assignment.workflow_id,
                    task_assignment.cmd,
                    task_assignment.env_vars,
                    task_assignment.additional_distribution,
                    task_assignment.fast_register_dir,
                    task_status_tx,
                    task_status_report_interval_seconds,
                    last_ack_grace_period_seconds,
                    backlog_tx,
                    backlog_rx,
                    executor_tx,
                    executor_rx,
                    build_executor_tx,
                )
                .await?;
            }
        }

        // remove task_id from task_contexts
        {
            let mut task_contexts = task_contexts.write().unwrap();
            task_contexts.remove(&task_assignment.task_id);
        }

        Ok(())
    }
}

/*
 * TheManager
 */

struct TaskAssignment {
    task_id: String,
    namespace: String,
    workflow_id: String,
    cmd: Vec<String>,
    env_vars: HashMap<String, String>,
    additional_distribution: Option<String>,
    fast_register_dir: Option<String>,
}

pub struct MultiProcessManager {
    backlog_length: i32,
    execution_strategy: ExecutionStrategy,
    executor_registration_addr: String,
    fast_register_dir_override: String,
    last_ack_grace_period_seconds: u64,
    parallelism: i32,
    task_status_report_interval_seconds: u64,

    backlog_tx: Sender<()>,
    backlog_rx: Receiver<()>,
    executor_tx: Sender<Executor>,
    executor_rx: Receiver<Executor>,
    fast_register_ids: Arc<Mutex<HashSet<String>>>,
    task_assignment_tx: Sender<TaskAssignment>,
    task_assignment_rx: Receiver<TaskAssignment>,
    task_contexts: Arc<RwLock<HashMap<String, TaskContext>>>,
}

impl MultiProcessManager {
    pub fn new(
        backlog_length: i32,
        execution_strategy: ExecutionStrategy,
        executor_registration_addr: String,
        fast_register_dir_override: String,
        last_ack_grace_period_seconds: u64,
        parallelism: i32,
        task_status_report_interval_seconds: u64,
    ) -> MultiProcessManager {
        let (backlog_tx, backlog_rx) = async_channel::unbounded(); // TODO - this should be a bounded channel to re-enable backlog
        let (executor_tx, executor_rx) = async_channel::unbounded();
        let (task_assignment_tx, task_assignment_rx) = async_channel::unbounded();

        MultiProcessManager {
            backlog_length,
            execution_strategy,
            executor_registration_addr,
            fast_register_dir_override,
            last_ack_grace_period_seconds,
            parallelism,
            task_status_report_interval_seconds,
            backlog_tx,
            backlog_rx,
            executor_tx,
            executor_rx,
            fast_register_ids: Arc::new(Mutex::new(HashSet::new())),
            task_assignment_tx,
            task_assignment_rx,
            task_contexts: Arc::new(RwLock::new(HashMap::<String, TaskContext>::new())),
        }
    }
}

impl TaskManager for MultiProcessManager {
    fn ack(&self, task_id: String) -> impl Future<Output = Result<()>> {
        async move {
            let mut task_contexts = self.task_contexts.write().unwrap();

            // update last ack timestamp
            if let Some(ref mut task_context) = task_contexts.get_mut(&task_id) {
                task_context.last_ack_timestamp = SystemTime::now()
                    .duration_since(UNIX_EPOCH)
                    .unwrap()
                    .as_secs();
            }

            Ok(())
        }
        .into_future()
    }

    fn assign(
        &self,
        task_id: String,
        namespace: String,
        workflow_id: String,
        cmd: Vec<String>,
        env_vars: HashMap<String, String>,
    ) -> impl Future<Output = Result<()>> {
        let fast_register_ids = self.fast_register_ids.clone();
        let task_assignment_tx = self.task_assignment_tx.clone();

        async move {
            let (additional_distribution, actor_cmd, fast_register_dir) = transform_cmd(
                cmd,
                self.fast_register_dir_override.clone(),
                fast_register_ids,
            )?;
            let task_assignment = TaskAssignment {
                task_id,
                namespace,
                workflow_id,
                cmd: actor_cmd,
                env_vars,
                additional_distribution,
                fast_register_dir,
            };

            task_assignment_tx.send(task_assignment).await?;
            Ok(())
        }
        .into_future()
    }

    fn delete(&self, task_id: String) -> impl Future<Output = Result<()>> {
        async move {
            let mut task_contexts = self.task_contexts.write().unwrap();

            // send kill signal
            if let Some(ref mut task_context) = task_contexts.get_mut(&task_id) {
                task_context.kill_tx.send(()).await?;
            }

            Ok(())
        }
        .into_future()
    }

    fn get_capacity(&self) -> Result<Capacity> {
        Ok(Capacity {
            execution_count: self.parallelism - (self.executor_rx.len() as i32),
            execution_limit: self.parallelism,
            backlog_count: self.backlog_rx.len() as i32,
            backlog_limit: self.backlog_length,
        })
    }

    fn get_runtime<'a>(&self) -> Result<impl TaskManagerRuntime + 'a> {
        Ok(MultiProcessRuntime {
            backlog_tx: self.backlog_tx.clone(),
            backlog_rx: self.backlog_rx.clone(),
            execution_strategy: self.execution_strategy.clone(),
            executor_tx: self.executor_tx.clone(),
            executor_rx: self.executor_rx.clone(),
            executor_registration_addr: self.executor_registration_addr.clone(),
            last_ack_grace_period_seconds: self.last_ack_grace_period_seconds,
            parallelism: self.parallelism,
            task_assignment_rx: self.task_assignment_rx.clone(),
            task_contexts: self.task_contexts.clone(),
            task_status_report_interval_seconds: self.task_status_report_interval_seconds,
        })
    }
}

pub struct MultiProcessRuntime {
    backlog_tx: Sender<()>,
    backlog_rx: Receiver<()>,
    execution_strategy: ExecutionStrategy,
    executor_tx: Sender<Executor>,
    executor_rx: Receiver<Executor>,
    executor_registration_addr: String,
    last_ack_grace_period_seconds: u64,
    parallelism: i32,
    task_assignment_rx: Receiver<TaskAssignment>,
    task_contexts: Arc<RwLock<HashMap<String, TaskContext>>>,
    task_status_report_interval_seconds: u64,
}

impl TaskManagerRuntime for MultiProcessRuntime {
    async fn run(&self, task_status_tx: Sender<TaskStatus>) -> Result<()> {
        let (backlog_tx, backlog_rx) = (self.backlog_tx.clone(), self.backlog_rx.clone());
        let (executor_tx, executor_rx) = (self.executor_tx.clone(), self.executor_rx.clone());
        let (task_assignment_rx, task_contexts) =
            (self.task_assignment_rx.clone(), self.task_contexts.clone());

        let listener = TcpListener::bind(&self.executor_registration_addr).await?;

        // send `parallelism` count down build_executor_tx to initialize
        let (build_executor_tx, build_executor_rx) = async_channel::unbounded();
        for _ in 0..self.parallelism {
            let _ = build_executor_tx.send(()).await?;
        }

        let mut index = 0;
        loop {
            tokio::select! {
                _ = build_executor_rx.recv() => {
                    // start child process
                    let child = Command::new("unionai-actor-executor")
                        .arg("--executor-registration-addr")
                        .arg(self.executor_registration_addr.clone())
                        .arg("--id")
                        .arg(index.to_string())
                        .spawn()?;

                    let stream = listener.accept().await?.0;
                    let framed = Framed::new(stream, LengthDelimitedCodec::new());

                    let executor = Executor { framed, child };
                    self.executor_tx.send(executor).await?;

                    index += 1;
                },
                task_assignment_result = task_assignment_rx.recv() => {
                    let task_assignment= task_assignment_result?;

                    let (backlog_tx_clone, backlog_rx_clone) = (backlog_tx.clone(), backlog_rx.clone());
                    let build_executor_tx_clone = build_executor_tx.clone();
                    let (executor_tx_clone, executor_rx_clone) = (executor_tx.clone(), executor_rx.clone());
                    let (task_status_tx_clone, task_contexts_clone) = (task_status_tx.clone(), task_contexts.clone());

                    let (task_status_report_interval_seconds, last_ack_grace_period_seconds) =
                        (self.task_status_report_interval_seconds, self.last_ack_grace_period_seconds);

                    let execution_strategy = self.execution_strategy.clone();
                    tokio::spawn(async move {
                        if let Err(e) = execution_strategy.execute(
                            task_contexts_clone,
                            task_assignment,
                            task_status_tx_clone,
                            task_status_report_interval_seconds,
                            last_ack_grace_period_seconds,
                            backlog_tx_clone,
                            backlog_rx_clone,
                            executor_tx_clone,
                            executor_rx_clone,
                            build_executor_tx_clone,
                        ).await
                        {
                            warn!("failed to execute task '{:?}'", e);
                        }
                    });
                },
            }
        }
    }
}

fn transform_cmd(
    mut cmd: Vec<String>,
    fast_register_dir_override: String,
    fast_register_ids: Arc<Mutex<HashSet<String>>>,
) -> Result<(Option<String>, Vec<String>, Option<String>)> {
    let (mut pyflyte_execute_index, mut additional_distribution, mut fast_register_dir) =
        (None, None, None);

    if cmd[0].eq("pyflyte-fast-execute") {
        let mut dest_dir_index = None;
        for i in 0..cmd.len() {
            match cmd[i] {
                ref x if x.eq("--dest-dir") => dest_dir_index = Some(i + 1),
                ref x if x.eq("--additional-distribution") => {
                    additional_distribution = Some(cmd[i + 1].clone())
                }
                ref x if x.eq("pyflyte-execute") => pyflyte_execute_index = Some(i),
                ref x if x.eq("pyflyte-map-execute") => pyflyte_execute_index = Some(i),
                _ => (),
            }
        }

        // hash `additional_distribution` (fast register unique id) to identify a
        // unique subdirectory to decompress the fast register file
        let mut h = DefaultHasher::new();
        additional_distribution.hash(&mut h);
        let dir = format!("{}/{}", fast_register_dir_override, h.finish());

        if let Err(e) = std::fs::create_dir_all(dir.clone()) {
            bail!("failed to create fast register subdir '{}': {:?}", dir, e);
        }

        match dest_dir_index {
            Some(i) => cmd[i] = dir.clone(),
            None => {} // TODO - inject `--dest-dir fast_register_dir` into cmd_str
        }

        fast_register_dir = Some(dir);
    }

    // if the fast register file has already been processed we skip the download
    if let Some(ref additional_distribution_str) = additional_distribution {
        let mut fast_register_ids = fast_register_ids.lock().unwrap();
        if fast_register_ids.contains(additional_distribution_str) {
            additional_distribution = None;
        } else {
            fast_register_ids.insert(additional_distribution_str.clone());
        }
    }

    // strip `pyflyte-fast-execute` command if exists
    let cmd = if let Some(pyflyte_execute_index) = pyflyte_execute_index {
        cmd[pyflyte_execute_index..].to_vec()
    } else {
        cmd.clone()
    };

    Ok((additional_distribution, cmd, fast_register_dir))
}

/*
 * SuccessManager
 */

pub struct SuccessManager {
    task_tx: Sender<(String, String, String)>,
    task_rx: Receiver<(String, String, String)>,
}

impl SuccessManager {
    pub fn new() -> SuccessManager {
        let (task_tx, task_rx) = async_channel::unbounded();
        SuccessManager { task_tx, task_rx }
    }
}

impl TaskManager for SuccessManager {
    fn ack(&self, _task_id: String) -> impl Future<Output = Result<()>> {
        async { Ok(()) }.into_future()
    }

    fn assign(
        &self,
        task_id: String,
        namespace: String,
        workflow_id: String,
        _cmd: Vec<String>,
        _env_vars: HashMap<String, String>,
    ) -> impl Future<Output = Result<()>> {
        let task_tx = self.task_tx.clone();
        async move {
            let send_result = task_tx.send((task_id, namespace, workflow_id)).await;
            assert!(send_result.is_ok());
            Ok(())
        }
        .into_future()
    }

    fn delete(&self, _task_id: String) -> impl Future<Output = Result<()>> {
        async { Ok(()) }.into_future()
    }

    fn get_capacity(&self) -> Result<Capacity> {
        Ok(Capacity {
            execution_count: 0,
            execution_limit: 0,
            backlog_count: 0,
            backlog_limit: 0,
        })
    }

    fn get_runtime<'a>(&self) -> Result<impl TaskManagerRuntime + 'a> {
        let task_rx = self.task_rx.clone();
        Ok(SuccessRuntime { task_rx })
    }
}

pub struct SuccessRuntime {
    task_rx: Receiver<(String, String, String)>,
}

impl TaskManagerRuntime for SuccessRuntime {
    async fn run(&self, task_status_tx: Sender<TaskStatus>) -> Result<()> {
        let task_rx = self.task_rx.clone();
        loop {
            let task_result = task_rx.recv().await;
            assert!(task_result.is_ok());

            let (task_id, namespace, workflow_id) = task_result.unwrap();
            let task_status = TaskStatus {
                task_id,
                namespace,
                workflow_id,
                phase: SUCCEEDED,
                reason: "".to_string(),
            };

            let send_result = task_status_tx.send(task_status).await;
            assert!(send_result.is_ok());
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn multi_process_manager() {
        use std::time::Duration;

        let (task_status_tx, task_status_rx) = async_channel::unbounded();
        let (task_id, namespace, workflow_id) = (
            "task_id".to_string(),
            "namespace".to_string(),
            "workflow_id".to_string(),
        );
        let cmd = vec!["".to_string()];
        let env_vars = HashMap::new();

        // create manager
        let backlog_length = 5i32;
        let parallelism = 0i32; // explicitly set to 0 so that no workers are created
        let manager = MultiProcessManager::new(
            backlog_length,
            ExecutionStrategy::SuccessStrategy,
            "127.0.0.1:15605".to_string(),
            "/tmp/fasttask".to_string(),
            30,
            parallelism,
            10,
        );

        // start manager runtime
        let manager_runtime_result = manager.get_runtime();
        assert!(manager_runtime_result.is_ok());
        let manager_runtime = manager_runtime_result.unwrap();

        let manager_handle = tokio::spawn(async move {
            super::TaskManagerRuntime::run(&manager_runtime, task_status_tx).await
        });

        // validate get capacity works
        let capacity_result = manager.get_capacity();
        assert!(capacity_result.is_ok());

        let capacity = capacity_result.unwrap();
        assert_eq!(capacity.execution_count, 0);
        assert_eq!(capacity.execution_limit, parallelism);
        assert_eq!(capacity.backlog_count, 0);
        assert_eq!(capacity.backlog_limit, backlog_length);

        // assign task
        let assign_result = manager
            .assign(
                task_id.clone(),
                namespace.clone(),
                workflow_id.clone(),
                cmd,
                env_vars,
            )
            .await;
        assert!(assign_result.is_ok());

        // recv success TaskStatus
        let mut counter = 0;
        while task_status_rx.is_empty() && counter < 5 {
            tokio::time::sleep(Duration::from_millis(10)).await;
            counter += 1;
        }
        assert!(!task_status_rx.is_empty());

        let task_status_result = task_status_rx.recv().await;
        assert!(task_status_result.is_ok());
        let task_status = task_status_result.unwrap();

        assert_eq!(task_status.task_id, task_id);
        assert_eq!(task_status.namespace, namespace);
        assert_eq!(task_status.workflow_id, workflow_id);
        assert_eq!(task_status.phase, SUCCEEDED);

        // validate task_context
        let initial_ack_timestamp = {
            let task_contexts = manager.task_contexts.read().unwrap();
            assert_eq!(task_contexts.len(), 1);
            let task_context_option = task_contexts.get(&task_id);
            assert!(task_context_option.is_some());

            let task_context = task_context_option.unwrap();
            task_context.last_ack_timestamp
        };

        // ack and validate
        tokio::time::sleep(Duration::from_secs(1)).await;
        let ack_result = manager.ack(task_id.clone()).await;
        assert!(ack_result.is_ok());

        let last_ack_timestamp = {
            let task_contexts = manager.task_contexts.read().unwrap();
            assert_eq!(task_contexts.len(), 1);
            let task_context_option = task_contexts.get(&task_id);
            assert!(task_context_option.is_some());

            let task_context = task_context_option.unwrap();
            task_context.last_ack_timestamp
        };

        assert!(last_ack_timestamp > initial_ack_timestamp);

        // delete and validate
        let delete_result = manager.delete(task_id).await;
        assert!(delete_result.is_ok());

        tokio::time::sleep(Duration::from_millis(50)).await; // sleep to allow delete to propagate

        {
            let task_contexts = manager.task_contexts.read().unwrap();
            assert_eq!(task_contexts.len(), 0);
        }

        // cleanup
        manager_handle.abort();
        let _ = manager_handle.await;
    }

    #[tokio::test]
    async fn transform_cmd_happy() {
        let fast_register_dir_override = "/tmp".to_string();
        let fast_register_ids = Arc::new(Mutex::new(HashSet::new()));

        let cmd = vec!(
            "pyflyte-fast-execute",
            "--additional-distribution",
            "s3://my-s3-bucket/flytesnacks/development/7RZ4KNYDFWV7OLEGSCOH6RJ3SM======/fast00f05aad96604fca790e82f31fccdecc.tar.gz",
            "--dest-dir",
            ".",
            "--",
            "pyflyte-execute",
            "--inputs",
            "s3://my-s3-bucket/metadata/propeller/flytesnacks-development-avwmk5z2pbzccvdq5dqm/n0/data/inputs.pb",
            "--output-prefix",
            "s3://my-s3-bucket/metadata/propeller/flytesnacks-development-avwmk5z2pbzccvdq5dqm/n0/data/0",
            "--raw-output-data-prefix",
            "s3://my-s3-bucket/test/g8/avwmk5z2pbzccvdq5dqm-n0-0",
            "--checkpoint-path",
            "s3://my-s3-bucket/test/g8/avwmk5z2pbzccvdq5dqm-n0-0/_flytecheckpoints",
            "--prev-checkpoint",
            "\"\"",
            "--resolver",
            "flytekit.core.python_auto_container.default_task_resolver",
            "--",
            "task-module",
            "hello_world",
            "task-name",
            "say_hello").iter().map(|x| x.to_string()).collect::<Vec<String>>();

        let additional_distribution = cmd[2].clone();

        // execute once to ensure correct transformation
        let first_cmd_result = transform_cmd(
            cmd.clone(),
            fast_register_dir_override.clone(),
            fast_register_ids.clone(),
        );
        assert!(first_cmd_result.is_ok());

        let (first_additional_distribution, first_cmd, first_fast_register_dir) =
            first_cmd_result.unwrap();

        assert_eq!(first_cmd.len(), 18); // stripped `pyflyte-fast-execute` command
        assert_eq!(first_cmd[0], "pyflyte-execute");
        assert_eq!(first_additional_distribution, Some(additional_distribution)); // additional_distribution set
        assert_eq!(
            first_fast_register_dir,
            Some(format!(
                "{}/2680285252280366039",
                fast_register_dir_override
            ))
        ); // fast_register_dir was overridden

        // execute again to ensure we do not inform to `download_distribution` twice
        let second_cmd_result = transform_cmd(
            cmd.clone(),
            fast_register_dir_override.clone(),
            fast_register_ids.clone(),
        );
        assert!(second_cmd_result.is_ok());

        let (second_additional_distribution, _, _) = second_cmd_result.unwrap();

        assert_eq!(second_additional_distribution, None); // additional_distribution not set
    }
}
