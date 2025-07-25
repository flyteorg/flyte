// Struct representing the parsed arguments
#[derive(Debug, Default)]
pub struct TaskArgs {
    pub inputs: Option<String>,
    pub outputs_path: Option<String>,
    pub version: Option<String>,
    pub run_base_dir: Option<String>,
    pub raw_data_path: Option<String>,
    pub checkpoint_path: Option<String>,
    pub prev_checkpoint: Option<String>,
    pub name: Option<String>,
    pub run_name: Option<String>,
    pub project: Option<String>,
    pub domain: Option<String>,
    pub org: Option<String>,
    pub image_cache: Option<String>,
    pub tgz: Option<String>,
    pub pkl: Option<String>,
    pub dest: Option<String>,
    pub resolver: Option<String>,
    pub resolver_args: Vec<String>,
}

impl TaskArgs {
    pub fn from_command(command: &[String]) -> Result<Self, String> {
        let mut args = TaskArgs::default();
        let mut i = 0;
        let mut after_resolver = false;
        while i < command.len() {
            if after_resolver {
                // All remaining args go into resolver_args
                args.resolver_args.push(command[i].clone());
                i += 1;
                continue;
            }
            match command[i].as_str() {
                "--inputs" | "-i" => {
                    i += 1;
                    args.inputs = command.get(i).cloned();
                }
                "--outputs-path" | "-o" => {
                    i += 1;
                    args.outputs_path = command.get(i).cloned();
                }
                "--version" | "-v" => {
                    i += 1;
                    args.version = command.get(i).cloned();
                }
                "--run-base-dir" => {
                    i += 1;
                    args.run_base_dir = command.get(i).cloned();
                }
                "--raw-data-path" | "-r" => {
                    i += 1;
                    args.raw_data_path = command.get(i).cloned();
                }
                "--checkpoint-path" | "-c" => {
                    i += 1;
                    args.checkpoint_path = command.get(i).cloned();
                }
                "--prev-checkpoint" | "-p" => {
                    i += 1;
                    args.prev_checkpoint = command.get(i).cloned();
                }
                "--name" => {
                    i += 1;
                    args.name = command.get(i).cloned();
                }
                "--run-name" => {
                    i += 1;
                    args.run_name = command.get(i).cloned();
                }
                "--project" => {
                    i += 1;
                    args.project = command.get(i).cloned();
                }
                "--domain" => {
                    i += 1;
                    args.domain = command.get(i).cloned();
                }
                "--org" => {
                    i += 1;
                    args.org = command.get(i).cloned();
                }
                "--image-cache" => {
                    i += 1;
                    args.image_cache = command.get(i).cloned();
                }
                "--tgz" => {
                    i += 1;
                    args.tgz = command.get(i).cloned();
                }
                "--pkl" => {
                    i += 1;
                    args.pkl = command.get(i).cloned();
                }
                "--dest" => {
                    i += 1;
                    args.dest = command.get(i).cloned();
                }
                "--resolver" => {
                    i += 1;
                    args.resolver = command.get(i).cloned();
                    after_resolver = true;
                }
                _ => {
                    // skip positional/unknowns before resolver
                }
            }
            i += 1;
        }

        // Example: Ensure required arguments are present (customize as needed)
        if args.inputs.is_none() {
            return Err("Missing required argument --inputs".to_string());
        }
        if args.outputs_path.is_none() {
            return Err("Missing required argument --outputs-path".to_string());
        }
        if args.version.is_none() {
            return Err("Missing required argument --version".to_string());
        }

        Ok(args)
    }
}
