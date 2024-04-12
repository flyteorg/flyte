use crate::{ExecutorArgs, FAILED, SUCCEEDED};

use futures::sink::SinkExt;
use futures::stream::StreamExt;
use pyo3::prelude::*;
use pyo3::exceptions::PySystemExit;
use tokio::net::TcpStream;
use tokio_util::codec::{Framed, LengthDelimitedCodec};
use tracing::{debug, error, info};
use serde::{Deserialize, Serialize};

pub struct Executor {
    pub framed: Framed<TcpStream, LengthDelimitedCodec>,
    pub child: tokio::process::Child,
}

#[derive(Deserialize, Serialize)]
pub struct Task {
    pub cmd: Vec<String>,
}

#[derive(Deserialize, Serialize)]
pub struct Response {
    pub phase: i32,
    pub reason: Option<String>,
}

pub async fn run(args: ExecutorArgs, executor_registration_addr: &str) -> Result<(), Box<dyn std::error::Error>> {
    let stream = TcpStream::connect(executor_registration_addr).await?;
    let mut framed = Framed::new(stream, LengthDelimitedCodec::new());

    pyo3::prepare_freethreaded_python();
    Python::with_gil(|py| {
        // import to reduce python environment initialization time
        let _flytekit = PyModule::import(py, "flytekit").unwrap();
        let _entrypoint = PyModule::import(py, "flytekit.bin.entrypoint").unwrap();
    });

    loop {
        let buf = match framed.next().await {
            Some(Ok(buf)) => buf,
            Some(Err(e)) => {
                error!("executor '{}' failed to read from socket: {}", args.id, e);
                break;
            },
            None => {
                error!("executor '{}' connection closed", args.id);
                break;
            },
        };

        let task: Task = bincode::deserialize(&buf).unwrap();
        debug!("executor {} received work: {:?}", args.id, task.cmd);

        let result = Python::with_gil(|py| {
            let entrypoint = PyModule::import(py, "flytekit.bin.entrypoint").unwrap();

            // convert cmd to &str
            let cmd = task.cmd.iter().skip(1).map(|s| s.as_str()).collect::<Vec<&str>>();

            // execute command
            match task.cmd[0].as_str() {
                "pyflyte-fast-execute" => {
                    match entrypoint.call_method1("fast_execute_task_cmd", (cmd,),) {
                        Ok(_) => Ok(()),
                        Err(e) if e.is_instance_of::<PySystemExit>(py) => Ok(()),
                        Err(e) => Err(format!("{:?}", e)),
                    }
                },
                "pyflyte-execute" => {
                    match entrypoint.call_method1("execute_task_cmd", (cmd,),) {
                        Ok(_) => Ok(()),
                        Err(e) if e.is_instance_of::<PySystemExit>(py) => Ok(()),
                        Err(e) => Err(format!("{:?}", e)),
                    }
                },
                _ => Err(format!("unsupported task command '{}'", task.cmd[0])),
            }
        });

        let response = match result {
            Ok(_) => Response {
                phase: SUCCEEDED,
                reason: None,
            },
            Err(e) => Response {
                phase: FAILED,
                reason: Some(e),
            },
        };

        let buf = bincode::serialize(&response).unwrap();
        framed.send(buf.into()).await.unwrap();
    }

    info!("executor {} exiting", args.id);
    Ok(())
}
