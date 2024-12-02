use std::future::Future;
use std::pin::Pin;
use std::sync::{Arc, Mutex, RwLock};
use std::task::{Context, Poll, Waker};
use std::time::Duration;

use anyhow::Result;
use async_channel::{Receiver, Sender};
use tokio::time::Interval;

use crate::common::{FAILED, SUCCEEDED};
use crate::manager::CapacityReporter;
use crate::pb::fasttask::{HeartbeatRequest, TaskStatus};

pub trait Heartbeater {
    fn get_runtime<'a>(&self) -> Result<impl HeartbeatRuntime + 'a>;
}

#[trait_variant::make(HeartbeatRuntime: Send)]
pub trait LocalHeartbeatRuntime {
    async fn run(
        &self,
        task_status_rx: Receiver<TaskStatus>,
        heartbeat_tx: Sender<HeartbeatRequest>,
        capacity_reporter: CapacityReporter,
    ) -> Result<()>;
}

/*
 * PeriodicHeartbeater
 */

pub struct PeriodicHeartbeater {
    heartbeat_interval_seconds: u64,
    queue_id: String,
    worker_id: String,
}

impl PeriodicHeartbeater {
    pub fn new(
        heartbeat_interval_seconds: u64,
        queue_id: String,
        worker_id: String,
    ) -> PeriodicHeartbeater {
        PeriodicHeartbeater {
            heartbeat_interval_seconds,
            queue_id,
            worker_id,
        }
    }
}

impl Heartbeater for PeriodicHeartbeater {
    fn get_runtime<'a>(&self) -> Result<impl HeartbeatRuntime + 'a> {
        Ok(PeriodicHeartbeatRuntime {
            heartbeat_interval_seconds: self.heartbeat_interval_seconds,
            queue_id: self.queue_id.clone(),
            worker_id: self.worker_id.clone(),
        })
    }
}

pub struct PeriodicHeartbeatRuntime {
    heartbeat_interval_seconds: u64,
    queue_id: String,
    worker_id: String,
}

impl HeartbeatRuntime for PeriodicHeartbeatRuntime {
    async fn run(
        &self,
        task_status_rx: Receiver<TaskStatus>,
        heartbeat_tx: Sender<HeartbeatRequest>,
        capacity_reporter: CapacityReporter,
    ) -> Result<()> {
        let heartbeat_bool = Arc::new(Mutex::new(AsyncBool::new()));
        let mut heartbeat_trigger = HeartbeatTrigger {
            interval: tokio::time::interval(Duration::from_secs(self.heartbeat_interval_seconds)),
            async_bool: heartbeat_bool.clone(),
        };
        let task_statuses: Arc<RwLock<Vec<TaskStatus>>> = Arc::new(RwLock::new(vec![]));

        loop {
            tokio::select! {
               task_status_result = task_status_rx.recv() => {
                   let task_status: TaskStatus = task_status_result.unwrap();
                   if task_status.phase == SUCCEEDED || task_status.phase == FAILED {
                       // if task phase is terminal then trigger heartbeat immediately
                       let mut heartbeat_bool = heartbeat_bool.lock().unwrap();
                       heartbeat_bool.trigger();
                   }
                   let mut task_statuses = task_statuses.write().unwrap();
                   task_statuses.push(task_status);
               },
               _ = heartbeat_trigger.trigger() => {
                   let capacity = capacity_reporter.get().await?;
                   let mut heartbeat_request = HeartbeatRequest {
                       worker_id: self.worker_id.clone(),
                       queue_id: self.queue_id.clone(),
                       capacity: Some(capacity),
                       task_statuses: vec!(),
                   };

                   {
                       let mut task_statuses = task_statuses.write().unwrap();
                       heartbeat_request.task_statuses = task_statuses.clone();
                       task_statuses.clear();
                   }

                   heartbeat_tx.send(heartbeat_request).await?;
               },
            }
        }
    }
}

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

struct HeartbeatTrigger {
    interval: Interval,
    async_bool: Arc<Mutex<AsyncBool>>,
}

impl HeartbeatTrigger {
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

/*
 * PassthroughHeartbeater
 */

pub struct PassthroughHeartbeater {}

impl PassthroughHeartbeater {
    pub fn new() -> PassthroughHeartbeater {
        PassthroughHeartbeater {}
    }
}

impl Heartbeater for PassthroughHeartbeater {
    fn get_runtime<'a>(&self) -> Result<impl HeartbeatRuntime + 'a> {
        Ok(PassthroughHeartbeatRuntime {})
    }
}

pub struct PassthroughHeartbeatRuntime {}

impl HeartbeatRuntime for PassthroughHeartbeatRuntime {
    async fn run(
        &self,
        task_status_rx: Receiver<TaskStatus>,
        heartbeat_tx: Sender<HeartbeatRequest>,
        capacity_reporter: CapacityReporter,
    ) -> Result<()> {
        loop {
            let task_status_result = task_status_rx.recv().await;
            assert!(task_status_result.is_ok());

            let capacity_result = capacity_reporter.get().await;
            assert!(capacity_result.is_ok());

            let heartbeat_request = HeartbeatRequest {
                worker_id: "worker_id".to_string(),
                queue_id: "queue_id".to_string(),
                capacity: Some(capacity_result.unwrap()),
                task_statuses: vec![task_status_result.unwrap()],
            };

            let heartbeat_result = heartbeat_tx.send(heartbeat_request).await;
            assert!(heartbeat_result.is_ok());
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn periodic_heartbeater() {
        use crate::pb::fasttask::Capacity;
        use std::time::Instant;

        // spawn capacity reporting mechanism
        let (capacity_trigger_tx, capacity_trigger_rx) = async_channel::unbounded();
        let (capacity_report_tx, capacity_report_rx) = async_channel::unbounded();

        let capacity_handle = tokio::spawn(async move {
            loop {
                let recv_result = capacity_trigger_rx.recv().await;
                assert!(recv_result.is_ok());

                let send_result = capacity_report_tx
                    .send(Capacity {
                        execution_count: 0,
                        execution_limit: 0,
                        backlog_count: 0,
                        backlog_limit: 0,
                    })
                    .await;
                assert!(send_result.is_ok());
            }
        });

        // start heartbeater
        let (heartbeat_tx, heartbeat_rx) = async_channel::unbounded();
        let (task_status_tx, task_status_rx) = async_channel::unbounded();
        let heartbeater =
            PeriodicHeartbeater::new(1u64, "queue_id".to_string(), "worker_id".to_string());

        let heartbeat_runtime_result = heartbeater.get_runtime();
        assert!(heartbeat_runtime_result.is_ok());
        let heartbeat_runtime = heartbeat_runtime_result.unwrap();

        let heartbeat_handle = tokio::spawn(async move {
            let capacity_reporter = CapacityReporter::new(capacity_trigger_tx, capacity_report_rx);
            super::HeartbeatRuntime::run(
                &heartbeat_runtime,
                task_status_rx,
                heartbeat_tx,
                capacity_reporter,
            )
            .await
        });

        // flush heartbeat_rx
        while !heartbeat_rx.is_empty() {
            let heartbeat_result = heartbeat_rx.recv().await;
            assert!(heartbeat_result.is_ok());
        }

        // send terminal task_status and verify immediate heartbeat
        let task_status_succeeded = TaskStatus {
            task_id: "task_id".to_string(),
            namespace: "namespace".to_string(),
            workflow_id: "workflow_id".to_string(),
            phase: SUCCEEDED,
            reason: "reason".to_string(),
        };

        let succeeded_now = Instant::now();
        let succeeded_send_result = task_status_tx.send(task_status_succeeded).await;
        assert!(succeeded_send_result.is_ok());

        let succeeded_recv_result = heartbeat_rx.recv().await;
        assert!(succeeded_recv_result.is_ok());
        let succeeded_elapsed = succeeded_now.elapsed().as_millis();

        assert!(succeeded_elapsed <= 10); // time from terminal TaskStatus report to heartbeat is <10ms

        // flush heartbeat_rx
        while !heartbeat_rx.is_empty() {
            let heartbeat_result = heartbeat_rx.recv().await;
            assert!(heartbeat_result.is_ok());
        }

        // verify periodic heartbeats by capturing at least two heartbeat timestamps over 3 seconds
        // and validating the durations between each are < 1050ms (1 second with drift).
        let mut sleep_counter = 0;
        let mut heartbeat_counter = 0;
        let mut instant: Option<Instant> = None;

        while sleep_counter < 60 && heartbeat_counter < 3 {
            tokio::time::sleep(Duration::from_millis(50)).await;
            if !heartbeat_rx.is_empty() {
                let heartbeat_result = heartbeat_rx.recv().await;
                assert!(heartbeat_result.is_ok());

                heartbeat_counter += 1;
                if instant.is_some() {
                    assert!(instant.unwrap().elapsed().as_millis() < 1050);
                }

                instant = Some(Instant::now());
            }
            sleep_counter += 1;
        }

        assert!(heartbeat_counter >= 2);

        // cleanup
        capacity_handle.abort();
        let _ = capacity_handle.await;

        heartbeat_handle.abort();
        let _ = heartbeat_handle.await;
    }
}
