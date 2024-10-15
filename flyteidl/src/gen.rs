use glob::glob;
use std::fs;
use std::path::PathBuf;
use tonic_build;

fn main() {
    let mut protos: Vec<PathBuf> = Vec::new();

    // Iterate through directories in protos/flyteidl/
    for entry in fs::read_dir("protos/flyteidl").expect("Failed to read directory") {
        match entry {
            Ok(entry) => {
                if entry.file_type().expect("Failed to get file type").is_dir() {
                    let dir_path = entry.path().join("*.proto");
                    let pattern = dir_path.to_str().expect("Failed to convert path to string");

                    for proto in glob(pattern).expect("Failed to read glob pattern") {
                        match proto {
                            Ok(path) => protos.push(path),
                            Err(e) => eprintln!("Error while reading glob entry: {}", e),
                        }
                    }
                }
            }
            Err(e) => eprintln!("Error while reading directory entry: {}", e),
        }
    }

    tonic_build::configure()
        .out_dir("gen/pb_rust")
        .build_server(false)
        .build_client(true)
        // `type` already includes enum type
        .type_attribute(".", "#[pyo3::pyclass(dict, get_all, set_all)]")
        // `enum` cannot be `subclass`
        // .enum_attribute(".", "#[pyclass(get_all, set_all)]")
        .type_attribute(
            ".",
            "#[derive(::pyo3_macro::WithNew, serde::Serialize, serde::Deserialize)]",
        )
        .compile_well_known_types(true)
        .compile(
            &protos
                .iter()
                .map(|p| p.to_str().unwrap())
                .collect::<Vec<_>>(),
            &[
                "protos/",
                "protos/google/api/", // The path to where the googleapis protos are
                "protos/protoc-gen-openapiv2/options/", // The path to where the grpc-gateway protos are
            ],
        )
        .unwrap();
    println!("gRPC Rust client code generation completed successfully.");
}
