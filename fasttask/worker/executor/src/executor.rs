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
def _get_loaded_user_modules():
    # Returns modules that are user defined
    import os
    import site
    import sys

    builtin_names = set(sys.builtin_module_names)

    non_user_directories = site.getsitepackages() + [
        site.getusersitepackages(),
        os.path.join(sys.prefix, "lib"),
        os.path.join(sys.base_prefix, "lib"),
        os.path.dirname(sys.executable),
    ]
    outputs = []

    all_modules = list(sys.modules)

    for name in all_modules:
        if name in builtin_names:
            continue

        try:
            mod = sys.modules[name]
        except KeyError:
            continue

        try:
            mod_file = mod.__file__
        except Exception:
            continue

        if not isinstance(mod_file, str):
            continue

        try:
            is_non_user = any(
                os.path.commonpath([mod_file, non_user_directory]) == non_user_directory
                for non_user_directory in non_user_directories
            )
            if is_non_user:
                continue

        except ValueError:
            # This means that the files are not in the same drive, which means the
            # mod_file are not in any of the directories
            pass

        outputs.append(name)

    return outputs

def reset_env():
    # Unload modules that are user defined and resets Launchplan cache
    import sys
    from contextlib import suppress
    user_modules = _get_loaded_user_modules()
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

        // python env setup
        if let Err(e) = setup_python_env(
            py,
            _fast_registration.clone(),
            _os.clone(),
            &task.fast_register_dir,
            &task.additional_distribution,
        )
        .await
        {
            error!("executor '{}' failed to setup python env: {}", args.id, e);
            break;
        }

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

    Ok(())
}

async fn cleanup_python_env<'a>(
    py: Python<'a>,
    _fast_task: Bound<'a, PyModule>,
    _os: Bound<'a, PyModule>,
    fast_register_dir: &Option<String>,
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

    // reset to environment
    _fast_task.call_method0("reset_env")?;

    Ok(())
}
