use futures::sink::SinkExt;
use futures::stream::StreamExt;
use pyo3::exceptions::PySystemExit;
use pyo3::prelude::*;
use tokio::net::TcpStream;
use tokio_util::codec::{Framed, LengthDelimitedCodec};
use tracing::{debug, error, info};
use unionai_actor_bridge::common::{Response, Task, FAILED, SUCCEEDED};

#[derive(Debug)]
pub struct ExecutorArgs {
    pub executor_registration_addr: String,
    pub id: usize,
}

pub async fn run(py: Python<'_>, args: ExecutorArgs) -> Result<(), Box<dyn std::error::Error>> {
    let stream = TcpStream::connect(args.executor_registration_addr).await?;
    let mut framed = Framed::new(stream, LengthDelimitedCodec::new());

    // import to reduce python environment initialization time
    let _flytekit = PyModule::import_bound(py, "flytekit").unwrap();
    let _entrypoint = PyModule::import_bound(py, "flytekit.bin.entrypoint").unwrap();

    loop {
        let buf = match framed.next().await {
            Some(Ok(buf)) => buf,
            Some(Err(e)) => {
                error!("executor '{}' failed to read from socket: {}", args.id, e);
                break;
            }
            None => {
                error!("executor '{}' connection closed", args.id);
                break;
            }
        };

        let task: Task = bincode::deserialize(&buf).unwrap();
        debug!("executor {} received work: {:?}", args.id, task.cmd);

        let result = {
            let entrypoint = PyModule::import_bound(py, "flytekit.bin.entrypoint").unwrap();

            // convert cmd to &str
            let cmd = task
                .cmd
                .iter()
                .skip(1)
                .map(|s| s.as_str())
                .collect::<Vec<&str>>();

            // execute command
            match task.cmd[0].as_str() {
                "pyflyte-fast-execute" => {
                    match entrypoint.call_method1("fast_execute_task_cmd", (cmd,)) {
                        Ok(_) => Ok(()),
                        Err(e) if e.is_instance_of::<PySystemExit>(py) => Ok(()),
                        Err(e) => Err(format!("{:?}", e)),
                    }
                }
                "pyflyte-execute" => match entrypoint.call_method1("execute_task_cmd", (cmd,)) {
                    Ok(_) => Ok(()),
                    Err(e) if e.is_instance_of::<PySystemExit>(py) => Ok(()),
                    Err(e) => Err(format!("{:?}", e)),
                },
                _ => Err(format!("unsupported task command '{}'", task.cmd[0])),
            }
        };

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
