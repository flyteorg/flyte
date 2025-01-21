use std::collections::HashMap;
use std::future::Future;
use std::pin::Pin;
use std::sync::{Arc, Mutex};
use std::task::{Context, Poll, Waker};

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

pub struct AsyncBoolFuture {
    async_bool: Arc<Mutex<AsyncBool>>,
}

impl AsyncBoolFuture {
    pub fn new(async_bool: Arc<Mutex<AsyncBool>>) -> Self {
        Self { async_bool }
    }
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

pub struct AsyncBool {
    value: bool,
    waker: Option<Waker>,
}

impl AsyncBool {
    pub fn new() -> Self {
        Self {
            value: false,
            waker: None,
        }
    }

    pub fn trigger(&mut self) {
        self.value = true;
        if let Some(waker) = &self.waker {
            waker.clone().wake();
        }
    }
}
