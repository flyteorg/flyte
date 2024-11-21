use std::collections::HashMap;

use async_channel::Sender;
use serde::{Deserialize, Serialize};
use tokio::net::TcpStream;
use tokio_util::codec::{Framed, LengthDelimitedCodec};

pub const FAILED: i32 = 7; // 7=retryable 8=permanent
pub const QUEUED: i32 = 3;
pub const RUNNING: i32 = 5;
pub const SUCCEEDED: i32 = 6;

pub struct TaskContext {
    pub kill_tx: Sender<()>,
    pub last_ack_timestamp: u64,
}

pub struct Executor {
    pub framed: Framed<TcpStream, LengthDelimitedCodec>,
    pub child: tokio::process::Child,
}

#[derive(Deserialize, Serialize)]
pub struct Task {
    pub cmd: Vec<String>,
    pub additional_distribution: Option<String>,
    pub fast_register_dir: Option<String>,
    pub env_vars: Option<HashMap<String, String>>,
}

#[derive(Deserialize, Serialize)]
pub struct Response {
    pub phase: i32,
    pub reason: Option<String>,

    pub executor_corrupt: bool,
}
