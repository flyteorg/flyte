use std::sync::{Arc, Mutex};
use std::time::Duration;

use crate::common::{AsyncBool, AsyncBoolFuture};
use crate::connection::{ConnectionBuilder, ConnectionRuntime};
use crate::heartbeater::{HeartbeatRuntime, Heartbeater};
use crate::manager::{CapacityReporter, TaskManager, TaskManagerRuntime};
use crate::pb::fasttask::heartbeat_response::Operation;

use tracing::{self, warn};

pub async fn run<T: ConnectionBuilder, U: Heartbeater + Send, V: TaskManager>(
    connection_builder: T,
    heartbeater: U,
    manager: V,
) -> Result<(), Box<dyn std::error::Error>> {
    let (capacity_trigger_tx, capacity_trigger_rx) = async_channel::unbounded();
    let (capacity_report_tx, capacity_report_rx) = async_channel::unbounded();
    let (heartbeat_tx, heartbeat_rx) = async_channel::unbounded();
    let (operation_tx, operation_rx) = async_channel::unbounded();
    let (task_status_tx, task_status_rx) = async_channel::unbounded();

    // initialize and start manager
    let manager_runtime_ready = Arc::new(Mutex::new(AsyncBool::new()));
    let manager_runtime_ready_clone = manager_runtime_ready.clone();

    let manager_runtime = manager.get_runtime()?; // TODO @hamersaw - handle error
    let _manager_handle = tokio::spawn(async move {
        // currently panicking if manager runtime fails rather than attempting to restart. this will
        // effectively force a new replica and failover tasks. a manager runtime failure should
        // only occur as a bug.
        if let Err(e) = manager_runtime
            .run(manager_runtime_ready_clone, task_status_tx)
            .await
        {
            panic!("manager failed: {}", e);
        }
    });

    let manager_runtime_future = AsyncBoolFuture::new(manager_runtime_ready);
    manager_runtime_future.await;

    // start heartbeater
    let heartbeat_runtime = heartbeater.get_runtime()?; // TODO @hamersaw - handle error
    let _heartbeat_handle = tokio::spawn(async move {
        let capacity_reporter = CapacityReporter::new(capacity_trigger_tx, capacity_report_rx);

        // same panic logic as manager_runtime execution above
        if let Err(e) = heartbeat_runtime
            .run(task_status_rx, heartbeat_tx, capacity_reporter)
            .await
        {
            panic!("heartbeater failed: {}", e);
        }
    });

    loop {
        let (heartbeat_rx_clone, operation_tx_clone) = (heartbeat_rx.clone(), operation_tx.clone());

        let connection_runtime = connection_builder.get_runtime()?; // TODO @hamersaw - handle error
        let mut connection_join_handle = tokio::spawn(async move {
            connection_runtime
                .run(heartbeat_rx_clone, operation_tx_clone)
                .await
        });

        // loop over operation channel
        loop {
            tokio::select! {
                _ = capacity_trigger_rx.recv() => {
                    let capacity = manager.get_capacity()?;
                    capacity_report_tx.send(capacity).await?;
                },
                operation_result = operation_rx.recv() => {
                    let operation = match operation_result {
                        Ok(operation) => operation,
                        Err(e) => {
                            warn!("failed to receive operation: {}", e);
                            continue
                        }
                    };

                    let result = match Operation::from_i32(operation.operation) {
                        Some(Operation::Assign) => manager.assign(operation.task_id.clone(), operation.namespace,
                            operation.workflow_id, operation.cmd, operation.env_vars).await,
                        Some(Operation::Ack) => manager.ack(operation.task_id.clone()).await,
                        Some(Operation::Delete) => manager.delete(operation.task_id.clone()).await,
                        _ => unimplemented!(),
                    };

                    if let Err(e) = result {
                        warn!("failed to evaluate operation '{}' on task '{}': {}", operation.operation, operation.task_id, e);
                    }
                },
                result = &mut connection_join_handle => {
                    if let Err(e) = result {
                        warn!("connection failed: {}", e);
                    }

                    // attempt to flush heartbeat channel and break to reset the connection
                    while !heartbeat_rx.is_empty() {
                        if let Err(e) = heartbeat_rx.recv().await {
                            warn!("failed to flush heartbeat channel: {}", e);
                            break;
                        }
                    }

                    break
                },
            }
        }

        // sleep between connection reconnects
        tokio::time::sleep(Duration::from_secs(1)).await;
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn happy() {
        use std::collections::HashMap;

        use crate::common::SUCCEEDED;
        use crate::connection::TestConnectionBuilder;
        use crate::heartbeater::PassthroughHeartbeater;
        use crate::manager::SuccessManager;
        use crate::pb::fasttask::HeartbeatResponse;

        // build bridge components
        let (heartbeat_tx, heartbeat_rx) = async_channel::unbounded();
        let (operation_tx, operation_rx) = async_channel::unbounded();

        let connection = TestConnectionBuilder::new(heartbeat_tx, operation_rx);
        let heartbeater = PassthroughHeartbeater::new();
        let manager = SuccessManager::new();

        // start bridge
        let run_handle = tokio::spawn(async move {
            let run_result = run(connection, heartbeater, manager).await;
            assert!(run_result.is_ok());
        });

        // send operation
        let heartbeat_request = HeartbeatResponse {
            task_id: "task_id".to_string(),
            namespace: "namespace".to_string(),
            workflow_id: "workflow_id".to_string(),
            cmd: vec!["c".to_string(), "m".to_string(), "d".to_string()],
            env_vars: HashMap::new(),
            operation: Operation::Assign.into(),
        };

        let operation_result = operation_tx.send(heartbeat_request.clone()).await;
        assert!(operation_result.is_ok());

        // wait for heartbeat response (50ms)
        let mut counter = 0;
        while heartbeat_rx.is_empty() && counter < 5 {
            tokio::time::sleep(Duration::from_millis(10)).await;
            counter += 1;
        }
        assert!(!heartbeat_rx.is_empty());

        let heartbeat_response_result = heartbeat_rx.recv().await;
        assert!(heartbeat_response_result.is_ok());
        let heartbeat_response = heartbeat_response_result.unwrap();

        // validate heartbeat response
        assert!(heartbeat_response.capacity.is_some());
        let capacity = heartbeat_response.capacity.unwrap();
        assert_eq!(capacity.execution_count, 0);
        assert_eq!(capacity.execution_limit, 0);
        assert_eq!(capacity.backlog_count, 0);
        assert_eq!(capacity.backlog_limit, 0);

        assert_eq!(heartbeat_response.task_statuses.len(), 1);

        let task_status = &heartbeat_response.task_statuses[0];
        assert_eq!(task_status.task_id, heartbeat_request.task_id);
        assert_eq!(task_status.namespace, heartbeat_request.namespace);
        assert_eq!(task_status.workflow_id, heartbeat_request.workflow_id);
        assert_eq!(task_status.phase, SUCCEEDED);

        // cleanup
        run_handle.abort();
        let _ = run_handle.await;
    }
}
