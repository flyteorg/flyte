use std::collections::HashMap;

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
import inspect
import sys
import os
from contextlib import suppress

def _get_fast_register_modules(fast_register_path):
    # Get modules that are defined in fast_register_path
    fast_register_modules = []

    module_names = list(sys.modules)
    for module_name in module_names:
        # Do not unload this module
        if module_name == "_union_fast_task":
            continue

        try:
            module_file_path = inspect.getfile(sys.modules[module_name])
        except (TypeError, KeyError):
            continue

        absolute_file_path = os.path.abspath(module_file_path)
        if not os.path.commonpath([fast_register_path, absolute_file_path]) == fast_register_path:
            continue

        fast_register_modules.append(module_name)

    return fast_register_modules

def reset_env(fast_register_path):
    # Unload modules that are user defined and resets Launchplan cache
    user_modules = _get_fast_register_modules(fast_register_path)
    for name in user_modules:
        with suppress(KeyError):
            del sys.modules[name]

    from flytekit import LaunchPlan

    if hasattr(LaunchPlan, "CACHE"):
        LaunchPlan.CACHE = {}
    "#,
        "_union_fast_task.py",
        "_union_fast_task",
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

        let mut original_env_vars = HashMap::new();

        // python env setup
        if let Err(e) = setup_python_env(
            py,
            _fast_registration.clone(),
            _os.clone(),
            &task.fast_register_dir,
            &task.additional_distribution,
            &task.env_vars,
            &mut original_env_vars,
        )
        .await
        {
            error!("executor '{}' failed to setup python env: {}", args.id, e);
            break;
        }

        let entrypoint_method = task.cmd.get(0).map(|s| s.as_str());
        let entrypoint_cmd = match entrypoint_method {
            Some("pyflyte-map-execute") => "map_execute_task_cmd",
            Some("pyflyte-execute") => "execute_task_cmd",
            _ => {
                return Err(
                    format!("unsupported execution method '{:?}'", entrypoint_method).into(),
                )
            }
        };

        // run python command
        let cmd = task
            .cmd
            .iter()
            .skip(1)
            .map(|s| s.as_str())
            .collect::<Vec<&str>>();

        let cmd_result = match _entrypoint.call_method1(entrypoint_cmd, (cmd,)) {
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
            &task.env_vars,
            &mut original_env_vars,
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
    env_vars: &Option<HashMap<String, String>>,
    original_env_vars: &mut HashMap<String, Option<String>>,
) -> Result<(), PyErr> {
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

    // Set environment variables if provided
    if let Some(ref env_vars) = env_vars {
        let os_environ = _os.getattr("environ")?;
        for (key, value) in env_vars {
            original_env_vars.insert(
                key.clone(),
                match os_environ.get_item(key) {
                    Ok(v) => v.extract::<String>().ok(),
                    Err(_) => None,
                },
            );

            os_environ.set_item(key, value)?;
        }
    }

    Ok(())
}

async fn cleanup_python_env<'a>(
    py: Python<'a>,
    _fast_task: Bound<'a, PyModule>,
    _os: Bound<'a, PyModule>,
    fast_register_dir: &Option<String>,
    env_vars: &Option<HashMap<String, String>>,
    original_env_vars: &mut HashMap<String, Option<String>>,
    cwd: Bound<'_, PyAny>,
) -> Result<(), PyErr> {
    let locals = [("sys", py.import_bound("sys")?)].into_py_dict_bound(py);

    // flush stdout to ensure any print statements are persisted
    py.eval_bound("sys.stdout.flush()", None, Some(&locals))?;

    // remote `fast_register_dir` from sys path
    if let Some(ref fast_register_dir) = fast_register_dir {
        py.eval_bound(
            &format!("sys.path.remove(\"{}\")", fast_register_dir),
            None,
            Some(&locals),
        )?;

        // reset to environment
        _fast_task.call_method1("reset_env", (fast_register_dir,))?;
    }

    // Unset environment variables if provided
    if let Some(ref env_vars) = env_vars {
        let os_environ = _os.getattr("environ")?;
        for key in env_vars.keys() {
            match original_env_vars.get(key).and_then(|v| v.as_deref()) {
                Some(value) => os_environ.set_item(key, value)?,
                None => os_environ.del_item(key)?,
            }
        }
    }

    // update workdir to original;
    _os.call_method1("chdir", (cwd,))?;
    Ok(())
}
