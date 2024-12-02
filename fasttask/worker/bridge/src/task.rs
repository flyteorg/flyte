use std::collections::HashMap;
use std::sync::{Arc, RwLock};
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

use crate::common::{Executor, Response, Task};
use crate::common::{TaskContext, FAILED, QUEUED, RUNNING};
use crate::pb::fasttask::TaskStatus;

use anyhow::{bail, Result};
use async_channel::{self, Receiver, Sender, TryRecvError};
use futures::sink::SinkExt;
use futures::stream::StreamExt;
use tracing::{debug, info, warn};

struct RunCommandResult {
    build_new_executor: bool,
    killed: bool,
}

pub async fn execute(
    kill_rx: Receiver<()>,
    task_contexts: Arc<RwLock<HashMap<String, TaskContext>>>,
    task_id: String,
    namespace: String,
    workflow_id: String,
    cmd: Vec<String>,
    env_vars: HashMap<String, String>,
    additional_distribution: Option<String>,
    fast_register_dir: Option<String>,
    task_status_tx: Sender<TaskStatus>,
    task_status_report_interval_seconds: u64,
    last_ack_grace_period_seconds: u64,
    backlog_tx: Sender<()>,
    backlog_rx: Receiver<()>,
    executor_tx: Sender<Executor>,
    executor_rx: Receiver<Executor>,
    build_executor_tx: Sender<()>,
) -> Result<()> {
    // check if task may be executed or backlogged
    let (mut executor, backlogged) = is_executable(&executor_rx, &backlog_tx).await?;

    info!(
        "starting task execution task_id={:?} executor={:?} backlogged={:?}",
        task_id,
        executor.is_some(),
        backlogged
    );
    if executor.is_none() && !backlogged {
        // if no executor and not backlogged then we drop the task transparently and allow grace
        // period to failover to another worker
        return Ok(());
    }

    // if backlogged we wait until we can execute
    let (mut phase, mut reason) = (QUEUED, "".to_string());
    if backlogged {
        executor = wait_in_backlog(
            task_contexts.clone(),
            &kill_rx,
            &task_id,
            &namespace,
            &workflow_id,
            &task_status_tx,
            task_status_report_interval_seconds,
            last_ack_grace_period_seconds,
            &mut phase,
            &mut reason,
            &executor_rx,
            &backlog_rx,
        )
        .await?;
    }

    // execute task by running command - the only way that executor is None is if the task is
    // previous killed during the previous `wait_in_backlog` function call
    let killed = if let Some(mut executor) = executor {
        let result = match run_command(
            task_contexts.clone(),
            &kill_rx,
            &task_id,
            &namespace,
            &workflow_id,
            cmd,
            env_vars,
            additional_distribution,
            fast_register_dir,
            &task_status_tx,
            task_status_report_interval_seconds,
            last_ack_grace_period_seconds,
            &mut phase,
            &mut reason,
            &mut executor,
        )
        .await
        {
            Ok(result) => Ok(result),
            Err(e) => Err(format!("failed to run command '{}'", e)),
        };

        // if the executor is dropped the child process is automatically killed, so we either
        // re-enqueue the existing executor or indicate that a new instance should be created
        let run_command_result = match result {
            Ok(run_command_result) => run_command_result,
            Err(e) => {
                executor_tx.send(executor).await?;
                bail!(e);
            }
        };

        if run_command_result.build_new_executor {
            build_executor_tx.send(()).await?;
        } else {
            executor_tx.send(executor).await?;
        }

        run_command_result.killed
    } else {
        true
    };

    // if not closed then report terminal status
    if !killed {
        report_terminal_status(
            task_contexts.clone(),
            &kill_rx,
            &task_id,
            &namespace,
            &workflow_id,
            &task_status_tx,
            task_status_report_interval_seconds,
            last_ack_grace_period_seconds,
            &mut phase,
            &mut reason,
        )
        .await?;
    }

    Ok(())
}

async fn is_executable<T>(
    executor_rx: &Receiver<T>,
    backlog_tx: &Sender<()>,
) -> Result<(Option<T>, bool)> {
    match executor_rx.try_recv() {
        Ok(executor) => return Ok((Some(executor), false)),
        Err(TryRecvError::Closed) => bail!("executor_rx is closed"),
        Err(TryRecvError::Empty) => {}
    }

    backlog_tx.send(()).await?;
    Ok((None, true))
}

async fn report_terminal_status(
    task_contexts: Arc<RwLock<HashMap<String, TaskContext>>>,
    kill_rx: &Receiver<()>,
    task_id: &str,
    namespace: &str,
    workflow_id: &str,
    task_status_tx: &Sender<TaskStatus>,
    task_status_report_interval_seconds: u64,
    last_ack_grace_period_seconds: u64,
    phase: &mut i32,
    reason: &mut String,
) -> Result<()> {
    // send completed task status until deleted
    let mut interval =
        tokio::time::interval(Duration::from_secs(task_status_report_interval_seconds));
    loop {
        tokio::select! {
            _ = interval.tick() => {
                // if last_ack_timestamp > grace_period then kill tasks and delete from task_contexts
                let last_ack_timestamp;
                {
                    let task_contexts = task_contexts.read().unwrap();

                    // can only be `Some` because this thread is the only way to delete
                    let task_context = task_contexts.get(task_id).unwrap();
                    last_ack_timestamp = task_context.last_ack_timestamp;
                }

                if SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() - last_ack_timestamp > last_ack_grace_period_seconds {
                    warn!("task timed out task_id={:?}", task_id);
                    kill_rx.close();
                    return Ok(());
                }

                // send task status
                let error = task_status_tx.send(TaskStatus{
                    task_id: task_id.to_string(),
                    namespace: namespace.to_string(),
                    workflow_id: workflow_id.to_string(),
                    phase: *phase,
                    reason: reason.clone(),
                }).await;

                if error.is_err() {
                    warn!("failed to send task status error='{:?}'", error);
                }
            },
            _ = kill_rx.recv() => {
                kill_rx.close();
                return Ok(());
            },
        }
    }
}

async fn run_command(
    task_contexts: Arc<RwLock<HashMap<String, TaskContext>>>,
    kill_rx: &Receiver<()>,
    task_id: &str,
    namespace: &str,
    workflow_id: &str,
    cmd: Vec<String>,
    env_vars: HashMap<String, String>,
    additional_distribution: Option<String>,
    fast_register_dir: Option<String>,
    task_status_tx: &Sender<TaskStatus>,
    task_status_report_interval_seconds: u64,
    last_ack_grace_period_seconds: u64,
    phase: &mut i32,
    reason: &mut String,
    executor: &mut Executor,
) -> Result<RunCommandResult> {
    // execute command and monitor
    let task_start_ts = Instant::now();

    let buf = bincode::serialize(&Task {
        cmd,
        additional_distribution,
        fast_register_dir,
        env_vars: Some(env_vars),
    })
    .unwrap();
    executor.framed.send(buf.into()).await.unwrap();

    let mut interval =
        tokio::time::interval(Duration::from_secs(task_status_report_interval_seconds));

    loop {
        tokio::select! {
            result = executor.framed.next() => {
                // executor returned a result for the task execution
                info!("completed task_id {} in {}", task_id, task_start_ts.elapsed().as_millis());
                let buf = result.unwrap().unwrap();

                let response: Response = bincode::deserialize(&buf).unwrap();

                *phase = response.phase;
                *reason = response.reason.unwrap_or("".to_string());

                return Ok(
                    RunCommandResult{
                        build_new_executor: response.executor_corrupt,
                        killed: false,
                    })
            },
            result = executor.child.wait() => {
                // executor process completed
                match result {
                    Ok(exit_status) => {
                        *phase = FAILED;
                        *reason = format!("process completed with exit status: '{}'", exit_status);
                    },
                    Err(e) => {
                        *phase = FAILED;
                        *reason = format!("{}", e);
                    },
                }

                return Ok(
                    RunCommandResult{
                        build_new_executor: true,
                        killed: false,
                    })
            },
            _ = interval.tick() => {
                // if last_ack_timestamp > grace_period then kill tasks and delete from task_contexts
                let last_ack_timestamp;
                {
                    let task_contexts = task_contexts.read().unwrap();

                    // can only be `Some` because this thread is the only way to delete
                    let task_context = task_contexts.get(task_id).unwrap();
                    last_ack_timestamp = task_context.last_ack_timestamp;
                }

                if SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() - last_ack_timestamp > last_ack_grace_period_seconds {
                    warn!("task_id = {:?} is timed out", task_id);
                    kill_rx.close();

                    *phase = FAILED;
                    *reason = "task timed out on last ack".to_string();

                    // to abort the process we attempt to kill, if this fails it will automatically
                    // be aborted when the instance is dropped.
                    let _ = executor.child.kill().await;
                    return Ok(
                        RunCommandResult{
                            build_new_executor: true,
                            killed: false,
                        })
                }

                // send task status
                let error = task_status_tx.send(TaskStatus{
                    task_id: task_id.to_string(),
                    namespace: namespace.to_string(),
                    workflow_id: workflow_id.to_string(),
                    phase: *phase,
                    reason: reason.clone(),
                }).await;

                if error.is_err() {
                    warn!("ERROR = {:?}", error);
                }
            },
            _ = kill_rx.recv() => {
                debug!("task was killed task_id={:?}", task_id);
                kill_rx.close();

                *phase = FAILED;
                *reason = "task was killed".to_string();

                // to abort the process we attempt to kill, if this fails it will automatically
                // be aborted when the instance is dropped
                let _ = executor.child.kill().await;
                return Ok(
                    RunCommandResult{
                        build_new_executor: true,
                        killed: true,
                    })
            },
        }
    }
}

async fn wait_in_backlog<T>(
    task_contexts: Arc<RwLock<HashMap<String, TaskContext>>>,
    kill_rx: &Receiver<()>,
    task_id: &str,
    namespace: &str,
    workflow_id: &str,
    task_status_tx: &Sender<TaskStatus>,
    task_status_report_interval_seconds: u64,
    last_ack_grace_period_seconds: u64,
    phase: &mut i32,
    reason: &mut String,
    executor_rx: &Receiver<T>,
    backlog_rx: &Receiver<()>,
) -> Result<Option<T>> {
    let mut interval =
        tokio::time::interval(Duration::from_secs(task_status_report_interval_seconds));
    loop {
        tokio::select! {
            result = executor_rx.recv() => {
                let executor = match result {
                    Ok(executor) => executor,
                    Err(e) => bail!(format!("failed to retrieve executor: {:?}", e)),
                };

                backlog_rx.recv().await?;

                *phase = RUNNING;
                return Ok(Some(executor));
            },
            _ = interval.tick() => {
                // if last_ack_timestamp > grace_period then kill tasks and delete from task_contexts
                let last_ack_timestamp;
                {
                    let task_contexts = task_contexts.read().unwrap();

                    // can only be `Some` because this thread is the only way to delete
                    let task_context = task_contexts.get(task_id).unwrap();
                    last_ack_timestamp = task_context.last_ack_timestamp;
                }

                if SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() - last_ack_timestamp > last_ack_grace_period_seconds {
                    warn!("task timed out task_id={:?}", task_id);
                    kill_rx.close();
                    return Ok(None);
                }

                // send task status
                let error = task_status_tx.send(TaskStatus{
                    task_id: task_id.to_string(),
                    namespace: namespace.to_string(),
                    workflow_id: workflow_id.to_string(),
                    phase: *phase,
                    reason: reason.clone(),
                }).await;

                if error.is_err() {
                    warn!("failed to send task status error='{:?}'", error);
                }
            },
            _ = kill_rx.recv() => {
                kill_rx.close();
                return Ok(None);
            },
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    use async_channel::unbounded;
    use tokio::task::JoinHandle;

    #[tokio::test]
    async fn test_is_executable() {
        // This test uses a i32 instead of an Executor to check is_executable.

        let (sender_executor, receiver_executor) = unbounded::<i32>();
        let (sender_backlog, receiver_backlog) = unbounded::<()>();

        // If nothing in exectuor_rx, then the backlog receives a message.
        let result = is_executable(&receiver_executor, &sender_backlog).await;
        assert_eq!(result.unwrap(), (None, true));
        let backlog_item = receiver_backlog.try_recv();
        assert!(backlog_item.is_ok());
        assert!(receiver_backlog.is_empty());

        // is_executable consumes values on from sender_executor
        let result = sender_executor.send(4).await;
        assert!(result.is_ok());

        let result = is_executable(&receiver_executor, &sender_backlog).await;
        assert_eq!(result.ok(), Some((Some(4), false)));
        assert!(sender_executor.is_empty());

        // Closed backlog_tx returns error
        assert!(sender_backlog.close());
        let result = is_executable(&receiver_executor, &sender_backlog).await;
        assert!(result.is_err());

        // Closed executor_rx returns error
        let (sender_backlog, _) = unbounded::<()>();
        assert!(sender_executor.close());
        let result = is_executable(&receiver_executor, &sender_backlog).await;
        assert!(result.is_err());
    }

    async fn wait_in_backlog_setup<T: Send + 'static>(
        last_ack_grace_period_seconds: u64,
    ) -> (
        Sender<T>,
        Sender<()>,
        Receiver<TaskStatus>,
        JoinHandle<Result<Option<T>>>,
    ) {
        // initialize variables
        let task_contexts = Arc::new(RwLock::new(HashMap::<String, TaskContext>::new()));
        let (kill_tx, kill_rx) = async_channel::unbounded();
        let (task_id, namespace, workflow_id) = (
            "task_id".to_string(),
            "namespace".to_string(),
            "workflow_id".to_string(),
        );
        let (task_status_tx, task_status_rx) = async_channel::unbounded();
        let (task_status_report_interval_seconds, last_ack_grace_period_seconds) =
            (1, last_ack_grace_period_seconds);
        let (mut phase, mut reason) = (QUEUED, "".to_string());
        let (executor_tx, executor_rx) = async_channel::unbounded();
        let (backlog_tx, backlog_rx) = async_channel::unbounded();

        {
            let mut task_contexts = task_contexts.write().unwrap();
            task_contexts.insert(
                task_id.clone(),
                TaskContext {
                    kill_tx: kill_tx.clone(),
                    last_ack_timestamp: SystemTime::now()
                        .duration_since(UNIX_EPOCH)
                        .unwrap()
                        .as_secs(),
                },
            );
        }

        let backlog_send_result = backlog_tx.send(()).await;
        assert!(backlog_send_result.is_ok());

        // execute `wait_in_backlog`
        let wait_in_backlog_handle = tokio::spawn(async move {
            wait_in_backlog(
                task_contexts.clone(),
                &kill_rx,
                &task_id,
                &namespace,
                &workflow_id,
                &task_status_tx,
                task_status_report_interval_seconds,
                last_ack_grace_period_seconds,
                &mut phase,
                &mut reason,
                &executor_rx,
                &backlog_rx,
            )
            .await
        });

        (executor_tx, kill_tx, task_status_rx, wait_in_backlog_handle)
    }

    #[tokio::test]
    async fn wait_in_backlog_happy() {
        // initialize wait_in_backlog env
        let (executor_tx, _, task_status_rx, wait_in_backlog_handle) =
            wait_in_backlog_setup(5).await;

        // verify periodic task status' by capturing at least two timestamps over 3 seconds
        // and validating the durations between each are < 1050ms (1 second with drift).
        let mut sleep_counter = 0;
        let mut task_status_counter = 0;
        let mut instant: Option<Instant> = None;

        while sleep_counter < 60 && task_status_counter < 3 {
            tokio::time::sleep(Duration::from_millis(50)).await;
            if !task_status_rx.is_empty() {
                let task_status_result = task_status_rx.recv().await;
                assert!(task_status_result.is_ok());

                task_status_counter += 1;
                if instant.is_some() {
                    assert!(instant.unwrap().elapsed().as_millis() < 1050);
                }

                instant = Some(Instant::now());
            }
            sleep_counter += 1;
        }

        assert!(task_status_counter >= 2);

        // executor available
        let executor_send_result = executor_tx.send(()).await;
        assert!(executor_send_result.is_ok());

        let mut counter = 0;
        while counter < 5 && !wait_in_backlog_handle.is_finished() {
            tokio::time::sleep(Duration::from_millis(10)).await;
            counter += 1;
        }

        assert!(wait_in_backlog_handle.is_finished());

        let result = wait_in_backlog_handle.await;
        assert!(result.is_ok());
    }

    #[tokio::test]
    async fn wait_in_backlog_kill() {
        // initialize wait_in_backlog env
        let (_, kill_tx, _, wait_in_backlog_handle) = wait_in_backlog_setup::<()>(5).await;

        // use kill_tx to abort task
        let kill_result = kill_tx.send(()).await;
        assert!(kill_result.is_ok());

        let mut counter = 0;
        while counter < 5 && !wait_in_backlog_handle.is_finished() {
            tokio::time::sleep(Duration::from_millis(10)).await;
            counter += 1;
        }

        assert!(wait_in_backlog_handle.is_finished());

        let result = wait_in_backlog_handle.await;
        assert!(result.is_ok());
    }

    #[tokio::test]
    async fn wait_in_backlog_ack_timeout() {
        // initialize wait_in_backlog env
        let (_, _, _, wait_in_backlog_handle) = wait_in_backlog_setup::<()>(1).await;

        // wait for ack timeout
        let mut counter = 0;
        while counter < 40 && !wait_in_backlog_handle.is_finished() {
            tokio::time::sleep(Duration::from_millis(50)).await;
            counter += 1;
        }

        assert!(wait_in_backlog_handle.is_finished());

        let result = wait_in_backlog_handle.await;
        assert!(result.is_ok());
    }

    fn report_terminal_status_setup(
        last_ack_grace_period_seconds: u64,
    ) -> (Sender<()>, JoinHandle<Result<()>>, Receiver<TaskStatus>) {
        // initialize variables
        let task_contexts = Arc::new(RwLock::new(HashMap::<String, TaskContext>::new()));
        let (kill_tx, kill_rx) = async_channel::unbounded();
        let (task_id, namespace, workflow_id) = (
            "task_id".to_string(),
            "namespace".to_string(),
            "workflow_id".to_string(),
        );
        let (task_status_tx, task_status_rx) = async_channel::unbounded();
        let (task_status_report_interval_seconds, last_ack_grace_period_seconds) =
            (1, last_ack_grace_period_seconds);
        let (mut phase, mut reason) = (QUEUED, "".to_string());

        {
            let mut task_contexts = task_contexts.write().unwrap();
            task_contexts.insert(
                task_id.clone(),
                TaskContext {
                    kill_tx: kill_tx.clone(),
                    last_ack_timestamp: SystemTime::now()
                        .duration_since(UNIX_EPOCH)
                        .unwrap()
                        .as_secs(),
                },
            );
        }

        // execute `report_terminal_status`
        let report_terminal_status_handle = tokio::spawn(async move {
            report_terminal_status(
                task_contexts.clone(),
                &kill_rx,
                &task_id,
                &namespace,
                &workflow_id,
                &task_status_tx,
                task_status_report_interval_seconds,
                last_ack_grace_period_seconds,
                &mut phase,
                &mut reason,
            )
            .await
        });

        (kill_tx, report_terminal_status_handle, task_status_rx)
    }

    #[tokio::test]
    async fn report_terminal_status_kill() {
        let (kill_tx, report_terminal_status_handle, task_status_rx) =
            report_terminal_status_setup(5);

        // verify periodic task status' by capturing at least two timestamps over 3 seconds
        // and validating the durations between each are < 1050ms (1 second with drift).
        let mut sleep_counter = 0;
        let mut task_status_counter = 0;
        let mut instant: Option<Instant> = None;

        while sleep_counter < 60 && task_status_counter < 3 {
            tokio::time::sleep(Duration::from_millis(50)).await;
            if !task_status_rx.is_empty() {
                let task_status_result = task_status_rx.recv().await;
                assert!(task_status_result.is_ok());

                task_status_counter += 1;
                if instant.is_some() {
                    assert!(instant.unwrap().elapsed().as_millis() < 1050);
                }

                instant = Some(Instant::now());
            }
            sleep_counter += 1;
        }

        assert!(task_status_counter >= 2);

        // kill_tx
        let kill_send_result = kill_tx.send(()).await;
        assert!(kill_send_result.is_ok());

        let mut counter = 0;
        while counter < 5 && !report_terminal_status_handle.is_finished() {
            tokio::time::sleep(Duration::from_millis(10)).await;
            counter += 1;
        }

        assert!(report_terminal_status_handle.is_finished());

        let result = report_terminal_status_handle.await;
        assert!(result.is_ok());
    }

    #[tokio::test]
    async fn report_terminal_status_ack_timeout() {
        let (_, report_terminal_status_handle, _) = report_terminal_status_setup(1);

        // wait for ack timeout
        let mut counter = 0;
        while counter < 40 && !report_terminal_status_handle.is_finished() {
            tokio::time::sleep(Duration::from_millis(50)).await;
            counter += 1;
        }

        assert!(report_terminal_status_handle.is_finished());

        let result = report_terminal_status_handle.await;
        assert!(result.is_ok());
    }
}
