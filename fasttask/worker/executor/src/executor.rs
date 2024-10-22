use futures::sink::SinkExt;
use futures::stream::StreamExt;
use pyo3::exceptions::PySystemExit;
use pyo3::prelude::*;
use pyo3::types::IntoPyDict;
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
    let _fast_registration =
        PyModule::import_bound(py, "flytekit.tools.fast_registration").unwrap();
    let _fast_task = PyModule::from_code_bound(
        py,
        r#"
import sys

def set_modules(x):
    sys.modules = x
    "#,
        "fast_task.py",
        "fast_task",
    )?;
    let _os = PyModule::import_bound(py, "os").unwrap();

    let cwd = _os.call_method0("getcwd").unwrap();

    loop {
        // retrieve next task
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

        // python env setup
        let sys_modules = match setup_python_env(
            py,
            _fast_registration.clone(),
            _os.clone(),
            &task.fast_register_dir,
            &task.additional_distribution,
        )
        .await
        {
            Ok(sys_modules) => sys_modules,
            Err(e) => {
                error!("executor '{}' failed to setup python env: {}", args.id, e);
                break;
            }
        };

        // run python command
        let cmd = task
            .cmd
            .iter()
            .skip(1)
            .map(|s| s.as_str())
            .collect::<Vec<&str>>();

        let cmd_result = match _entrypoint.call_method1("execute_task_cmd", (cmd,)) {
            Ok(_) => Ok(()),
            Err(e) if e.is_instance_of::<PySystemExit>(py) => Ok(()),
            Err(e) => Err(format!("{:?}", e)),
        };

        // python env cleanup
        let mut executor_corrupt = false;
        if let Err(e) = cleanup_python_env(
            py,
            _fast_task.clone(),
            _os.clone(),
            &task.fast_register_dir,
            sys_modules,
            cwd.clone(),
        )
        .await
        {
            error!("executor '{}' failed to cleanup python env: {}", args.id, e);
            executor_corrupt = true;
        }

        // return result
        let response = match cmd_result {
            Ok(_) => Response {
                phase: SUCCEEDED,
                reason: None,
                executor_corrupt: executor_corrupt,
            },
            Err(e) => Response {
                phase: FAILED,
                reason: Some(e),
                executor_corrupt: executor_corrupt,
            },
        };

        let buf = bincode::serialize(&response).unwrap();
        if let Err(e) = framed.send(buf.into()).await {
            error!("executor '{}' failed to write to socket: {}", args.id, e);
            break;
        }

        if executor_corrupt {
            break;
        }
    }

    info!("executor {} exiting", args.id);
    Ok(())
}

async fn setup_python_env<'a>(
    py: Python<'a>,
    _fast_registration: Bound<'a, PyModule>,
    _os: Bound<'a, PyModule>,
    fast_register_dir: &Option<String>,
    additional_distribution: &Option<String>,
) -> Result<Bound<'a, PyAny>, PyErr> {
    let locals = [("sys", py.import_bound("sys")?)].into_py_dict_bound(py);

    if let Some(ref fast_register_dir) = fast_register_dir {
        // download `additional_distribution` if necessary
        if let Some(ref additional_distribution) = additional_distribution {
            _fast_registration.call_method1(
                "download_distribution",
                (additional_distribution, fast_register_dir),
            )?;
        }

        // append `fast_register_dir` to sys path
        py.eval_bound(
            &format!("sys.path.insert(0, \"{}\")", fast_register_dir),
            None,
            Some(&locals),
        )?;

        // update workdir to `fast_register_dir`;
        _os.call_method1("chdir", (fast_register_dir,))?;
    }

    // return current system modules
    py.eval_bound("sys.modules.copy()", None, Some(&locals))
}

async fn cleanup_python_env<'a>(
    py: Python<'a>,
    _fast_task: Bound<'a, PyModule>,
    _os: Bound<'a, PyModule>,
    fast_register_dir: &Option<String>,
    sys_modules: Bound<'_, PyAny>,
    cwd: Bound<'_, PyAny>,
) -> Result<(), PyErr> {
    let locals = [("sys", py.import_bound("sys")?)].into_py_dict_bound(py);

    // remote `fast_register_dir` from sys path
    if let Some(ref fast_register_dir) = fast_register_dir {
        py.eval_bound(
            &format!("sys.path.remove(\"{}\")", fast_register_dir),
            None,
            Some(&locals),
        )?;
    }

    // update workdir to original;
    _os.call_method1("chdir", (cwd,))?;

    // reset to original system modules
    _fast_task.call_method1("set_modules", (sys_modules,))?;

    Ok(())
}
