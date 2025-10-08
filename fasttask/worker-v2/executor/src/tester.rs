use futures::{SinkExt, StreamExt};
use pyo3::exceptions::PyRuntimeError;
use pyo3::PyErr;
use pyo3_async_runtimes::TaskLocals;
use std::fs::File;
use std::io::{self, Write};
use std::println;
use tempfile::TempDir;
use tokio::net::{TcpListener, TcpStream};
use tokio_util::codec::{Framed, LengthDelimitedCodec};
use tracing::error;
use tracing_appender::non_blocking;
use tracing_appender::non_blocking::WorkerGuard;
use tracing_subscriber::util::SubscriberInitExt;
use tracing_subscriber::{fmt, layer::SubscriberExt};
use unionai_actor_bridge::common::Task;

struct NoOpWriter;

impl Write for NoOpWriter {
    fn write(&mut self, buf: &[u8]) -> io::Result<usize> {
        Ok(buf.len()) // Pretend we wrote everything
    }

    fn flush(&mut self) -> io::Result<()> {
        Ok(())
    }
}

pub async fn run_server_test(addr: String, task: Task) -> Result<(), PyErr> {
    let listener = TcpListener::bind(&addr).await.map_err(|e| {
        PyRuntimeError::new_err(format!("Failed to bind to address {}: {}", addr, e))
    })?;
    println!("Server listening on {}", &addr);

    while let Ok((stream, addr)) = listener.accept().await {
        println!("Client connected from: {}", addr);
        tokio::spawn(handle_client(stream, task.clone()));
    }

    Ok(())
}

async fn handle_client(stream: TcpStream, task: Task) {
    let mut framed = Framed::new(stream, LengthDelimitedCodec::new());

    let task_bytes = bincode::serialize(&task).expect("Failed to serialize task");
    if let Err(e) = framed.send(task_bytes.into()).await {
        error!("Failed to send task to client: {}", e);
        return;
    }

    println!("Sent job");

    // Wait for response from worker
    match framed.next().await {
        Some(Ok(response)) => {
            println!(
                "Received response: {:?}",
                String::from_utf8_lossy(&response)
            );
        }
        Some(Err(e)) => {
            println!("Error receiving response: {}", e);
        }
        None => {
            println!("Client disconnected");
        }
    }

    println!("Client session ended");
}

fn make_test_guards() -> (TempDir, WorkerGuard, WorkerGuard) {
    // Create a temporary directory that will be automatically cleaned up
    let temp_dir = TempDir::new().expect("failed to create temp dir");

    let (non_blocking_writer, guard1) = non_blocking(NoOpWriter);
    let (non_blocking_writer, guard2) = non_blocking(NoOpWriter);

    // You can now use nb1 and nb2 as `MakeWriter`s in tracing_subscriber
    (temp_dir, guard1, guard2)
}

pub async fn run_task(task: Task, run_pool: bool, locals: TaskLocals) -> Result<(), PyErr> {
    tracing_subscriber::registry()
        .with(
            fmt::layer()
                .with_timer(fmt::time::UtcTime::rfc_3339())
                .with_span_events(fmt::format::FmtSpan::FULL)
                .with_ansi(false)
                .with_thread_names(true),
        )
        .init();

    let executor_args = crate::executor::ExecutorArgs {
        executor_registration_addr: String::from("127.0.0.1:8999"),
        id: 1,
        num_workers: Some(1),
    };

    let j = tokio::spawn(run_server_test(
        executor_args.executor_registration_addr.clone(),
        task,
    ));

    tokio::time::sleep(std::time::Duration::from_secs(1)).await;
    if run_pool {
        println!(
            "[Actor Core] Running worker pool with args: {:?}",
            executor_args
        );
        let (_temp_dir, guard1, guard2) = make_test_guards();
        crate::executor::run_worker_pool(executor_args, locals, _temp_dir, guard1, guard2).await?;
    }
    // Create a new Python GIL token inside the async block

    j.await
        .map_err(|e| PyRuntimeError::new_err(e.to_string()))?
}
