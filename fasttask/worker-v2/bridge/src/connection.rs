use std::future::{Future, IntoFuture};

use anyhow::{bail, Result};
use async_channel::{Receiver, Sender};
use tonic::Request;
use tracing::{debug, warn};

use crate::pb::fasttask::fast_task_client::FastTaskClient;
use crate::pb::fasttask::{HeartbeatRequest, HeartbeatResponse};

pub trait ConnectionBuilder {
    fn get_runtime<'a>(&self) -> Result<impl ConnectionRuntime + Send + 'a>;
}

pub trait ConnectionRuntime {
    fn run(
        &self,
        heartbeat_rx: Receiver<HeartbeatRequest>,
        operation_tx: Sender<HeartbeatResponse>,
    ) -> impl Future<Output = Result<()>> + Send;
}

/*
 * GRPCConnection
 */

pub struct GRPCConnectionBuilder {
    fasttask_url: String,
}

impl GRPCConnectionBuilder {
    pub fn new(fasttask_url: String) -> GRPCConnectionBuilder {
        GRPCConnectionBuilder { fasttask_url }
    }
}

impl ConnectionBuilder for GRPCConnectionBuilder {
    fn get_runtime<'a>(&self) -> Result<impl ConnectionRuntime + Send + 'a> {
        Ok(GRPCConnectionRuntime {
            fasttask_url: self.fasttask_url.clone(),
        })
    }
}

pub struct GRPCConnectionRuntime {
    fasttask_url: String,
}

impl ConnectionRuntime for GRPCConnectionRuntime {
    fn run(
        &self,
        heartbeat_rx: Receiver<HeartbeatRequest>,
        operation_tx: Sender<HeartbeatResponse>,
    ) -> impl Future<Output = Result<()>> + Send {
        let fasttask_url = self.fasttask_url.clone();

        async move {
            let mut client = FastTaskClient::connect(fasttask_url).await?;

            // initialize request forwarding async stream
            let outbound = async_stream::stream! {
                loop {
                    let heartbeat_request = match heartbeat_rx.recv().await {
                        Ok(heartbeat_request) => heartbeat_request,
                        Err(e) => {
                            warn!("failed to recv interval heartbeat: {}", e);
                            continue
                        },
                    };

                    debug!("sending heartbeat request = {:?}", heartbeat_request);
                    yield heartbeat_request;
                }
            };

            // gRPC call to open heartbeat stream
            let response = client.heartbeat(Request::new(outbound)).await?;
            let mut inbound = response.into_inner();

            // process heartbeat responses
            loop {
                let heartbeat_response = match inbound.message().await? {
                    Some(heartbeat_response) => heartbeat_response,
                    None => bail!("failed to receive heartbeat response"),
                };

                debug!("recv heartbeat response = {:?}", heartbeat_response);
                operation_tx.send(heartbeat_response).await?;
            }
        }
        .into_future()
    }
}

/*
 * TestConnectionBuilder
 */

#[derive(Clone)]
pub struct TestConnectionBuilder {
    heartbeat_tx: Sender<HeartbeatRequest>,
    operation_rx: Receiver<HeartbeatResponse>,
}

impl TestConnectionBuilder {
    pub fn new(
        heartbeat_tx: Sender<HeartbeatRequest>,
        operation_rx: Receiver<HeartbeatResponse>,
    ) -> TestConnectionBuilder {
        TestConnectionBuilder {
            heartbeat_tx,
            operation_rx,
        }
    }
}

impl ConnectionBuilder for TestConnectionBuilder {
    fn get_runtime<'a>(&self) -> Result<impl ConnectionRuntime + Send + 'a> {
        Ok(TestConnectionRuntime {
            heartbeat_tx: self.heartbeat_tx.clone(),
            operation_rx: self.operation_rx.clone(),
        })
    }
}

pub struct TestConnectionRuntime {
    heartbeat_tx: Sender<HeartbeatRequest>,
    operation_rx: Receiver<HeartbeatResponse>,
}

impl ConnectionRuntime for TestConnectionRuntime {
    fn run(
        &self,
        heartbeat_rx: Receiver<HeartbeatRequest>,
        operation_tx: Sender<HeartbeatResponse>,
    ) -> impl Future<Output = Result<()>> + Send {
        let (heartbeat_tx, heartbeat_rx) = (self.heartbeat_tx.clone(), heartbeat_rx.clone());
        let (operation_tx, operation_rx) = (operation_tx.clone(), self.operation_rx.clone());

        async move {
            // forward heartbeats and operations down outbound channels
            loop {
                tokio::select! {
                    heartbeat_recv_result = heartbeat_rx.recv() => {
                        assert!(heartbeat_recv_result.is_ok());

                        let heartbeat = heartbeat_recv_result.unwrap();
                        let heartbeat_send_result = heartbeat_tx.send(heartbeat).await;
                        assert!(heartbeat_send_result.is_ok());
                    },
                    operation_recv_result = operation_rx.recv() => {
                        assert!(operation_recv_result.is_ok());

                        let operation = operation_recv_result.unwrap();
                        let operation_send_result = operation_tx.send(operation).await;
                        assert!(operation_send_result.is_ok());
                    }
                };
            }
        }
        .into_future()
    }
}
