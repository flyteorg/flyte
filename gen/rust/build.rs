//! Compiles Protocol Buffers and FlatBuffers schema definitions into
//! native Rust types.

use prettyplease::unparse;
use quote::quote;
use std::fs::File;
use std::io::Write;
use std::path::PathBuf;
use syn::{parse_quote, Type};

type Error = Box<dyn std::error::Error>;
type Result<T, E = Error> = std::result::Result<T, E>;

fn generate_boxed_impls() -> String {
    // Define the types we want to implement Box<T> for
    let types: Vec<Type> = vec![
        parse_quote!(crate::validate::FieldRules),
        parse_quote!(crate::validate::RepeatedRules),
        parse_quote!(crate::validate::MapRules),
        parse_quote!(crate::flyteidl::core::LiteralType),
        parse_quote!(crate::flyteidl::core::Literal),
        parse_quote!(crate::flyteidl::core::Union),
        parse_quote!(crate::google::protobuf::FeatureSet),
        parse_quote!(crate::flyteidl::core::Scalar),
        // parse_quote!(crate::pyo3::types::PyBytes),
    ];

    // Generate the code using quote
    let tokens = quote! {
        use pyo3::prelude::*;
        use pyo3::conversion::{IntoPyObject, FromPyObject};
        use std::convert::Infallible;

        #(
            impl<'py> IntoPyObject<'py> for Box<#types> {
                type Target = PyAny;
                type Output = Bound<'py, Self::Target>;
                type Error = Infallible;

                fn into_pyobject(self, py: Python<'py>) -> Result<Self::Output, Self::Error> {
                    Ok(self.as_ref().clone().into_py(py).into_bound(py))
                }
            }

            impl<'py> FromPyObject<'py> for Box<#types> {
                fn extract_bound(obj: &Bound<'py, PyAny>) -> PyResult<Self> {
                    Ok(Box::new(#types::extract_bound(obj)?))
                }
            }
        )*
    };

    // Convert the generated code to a string and format it
    let syntax = syn::parse_file(&tokens.to_string()).expect("Failed to parse generated code");
    unparse(&syntax)
}

// Map file names to module paths
fn file_to_modpath(file: &str) -> Option<&'static str> {
    match file {
        "cloudidl.common.rs" => Some("crate::cloudidl::common"),
        "cloudidl.workflow.rs" => Some("crate::cloudidl::workflow"),
        "cloudidl.logs.dataplane.rs" => Some("crate::cloudidl::logs::dataplane"),
        "flyteidl.core.rs" => Some("crate::flyteidl::core"),
        "google.rpc.rs" => Some("crate::google::rpc"),
        "validate.rs" => Some("crate::validate"),
        _ => None,
    }
}

#[derive(Debug)]
struct ProstStructInfo {
    ty: Type,
    fq_name: String,
    fields: Vec<(String, String)>,
    is_tuple: bool,
    mod_path: String, // e.g. crate::cloudidl::common
}

// Find all prost-generated types in src/
fn find_prost_types(src_dir: &PathBuf) -> Vec<ProstStructInfo> {
    use std::collections::HashSet;
    use std::path::Path;
    use syn::{Fields, Item};

    let mut structs = Vec::new();
    let mut visited = HashSet::new();

    // Helper to recursively walk modules
    fn walk_items(items: &[Item], mod_path: &mut Vec<String>, structs: &mut Vec<ProstStructInfo>) {
        for item in items {
            match item {
                Item::Mod(m) => {
                    mod_path.push(m.ident.to_string());
                    // If inline module, recurse into its content
                    if let Some((_, items)) = &m.content {
                        walk_items(items, mod_path, structs);
                    }
                    // Outline modules are handled in the main loop
                    mod_path.pop();
                }
                Item::Struct(s) => {
                    let fq = format!("{}::{}", mod_path.join("::"), s.ident);
                    let ty = match syn::parse_str::<Type>(&fq) {
                        Ok(t) => t,
                        Err(_) => continue,
                    };
                    let (fields, is_tuple) = match &s.fields {
                        Fields::Named(fields) => {
                            let fields = fields
                                .named
                                .iter()
                                .map(|f| {
                                    let name = f.ident.as_ref().unwrap().to_string();
                                    let ty = quote::ToTokens::to_token_stream(&f.ty).to_string();
                                    (name, ty)
                                })
                                .collect();
                            (fields, false)
                        }
                        Fields::Unnamed(fields) => {
                            let fields = fields
                                .unnamed
                                .iter()
                                .enumerate()
                                .map(|(i, f)| {
                                    let name = format!("field{}", i);
                                    let ty = quote::ToTokens::to_token_stream(&f.ty).to_string();
                                    (name, ty)
                                })
                                .collect();
                            (fields, true)
                        }
                        Fields::Unit => (vec![], false),
                    };
                    let mod_path = mod_path.join("::");
                    let mod_path = if !mod_path.is_empty() {
                        format!("crate::{}", mod_path)
                    } else {
                        "crate".to_string()
                    };
                    structs.push(ProstStructInfo {
                        ty,
                        fq_name: fq,
                        fields,
                        is_tuple,
                        mod_path,
                    });
                }
                _ => {}
            }
        }
    }

    // Helper to parse a file and walk its items
    fn parse_and_walk(
        path: &Path,
        mod_path: &mut Vec<String>,
        structs: &mut Vec<ProstStructInfo>,
        visited: &mut HashSet<PathBuf>,
    ) {
        if !visited.insert(path.to_path_buf()) {
            return;
        }
        let content = match std::fs::read_to_string(path) {
            Ok(c) => c,
            Err(_) => return,
        };
        let ast: syn::File = match syn::parse_file(&content) {
            Ok(f) => f,
            Err(_) => return,
        };
        walk_items(&ast.items, mod_path, structs);
        // Recursively handle outline modules
        for item in &ast.items {
            if let Item::Mod(m) = item {
                if m.content.is_none() {
                    // Outline module: look for mod.rs or <mod>.rs
                    let mod_name = m.ident.to_string();
                    let parent = path.parent().unwrap_or(Path::new(""));
                    let mod_rs = parent.join(&mod_name).join("mod.rs");
                    let mod_file = parent.join(format!("{}.rs", mod_name));
                    if mod_rs.exists() {
                        mod_path.push(mod_name.clone());
                        parse_and_walk(&mod_rs, mod_path, structs, visited);
                        mod_path.pop();
                    } else if mod_file.exists() {
                        mod_path.push(mod_name.clone());
                        parse_and_walk(&mod_file, mod_path, structs, visited);
                        mod_path.pop();
                    }
                }
            }
        }
    }

    for entry in std::fs::read_dir(src_dir).unwrap() {
        let entry = entry.unwrap();
        let path = entry.path();
        let file_name = path.file_name().unwrap().to_string_lossy();

        if !file_name.ends_with(".rs") || file_name == "serde_impl.rs" || file_name == "lib.rs" {
            continue;
        }

        let modpath = match file_to_modpath(&file_name) {
            Some(m) => m.replace("crate::", ""),
            None => continue,
        };
        let mut mod_path: Vec<String> = modpath.split("::").map(|s| s.to_string()).collect();
        parse_and_walk(&path, &mut mod_path, &mut structs, &mut visited);
    }
    structs
}

fn generate_encode_decode(infos: &[ProstStructInfo]) -> String {
    // Helper to qualify types
    fn qualify_type(ty: &str, current_mod: &str) -> String {
        let primitives = [
            "String", "u8", "u16", "u32", "u64", "i8", "i16", "i32", "i64", "bool", "f32", "f64",
            "usize", "isize", "char", "str",
        ];
        let ty = ty.trim();
        let mut ty_clean = ty
            .replace(" :: ", "::")
            .replace("crate ::", "crate::")
            .replace("super ::", "super::")
            .replace("self ::", "self::");
        while ty_clean.contains("::::") {
            ty_clean = ty_clean.replace("::::", "::");
        }
        // Handle generics recursively (e.g., Option<Sort>, Vec<Filter>)
        if let Some(start) = ty_clean.find('<') {
            if let Some(end) = ty_clean.rfind('>') {
                let outer = &ty_clean[..start];
                let inner = &ty_clean[start + 1..end];
                let qualified_inner = inner
                    .split(',')
                    .map(|s| qualify_type(s, current_mod))
                    .collect::<Vec<_>>()
                    .join(", ");
                return format!("{}<{}>", outer, qualified_inner);
            }
        }
        // Handle super:: prefix by resolving up the module path
        if ty_clean.starts_with("super::") {
            let mut mod_parts: Vec<&str> = current_mod.split("::").collect();
            let mut ty_remainder = ty_clean.to_string();
            while ty_remainder.starts_with("super::") {
                ty_remainder = ty_remainder[7..].to_string();
                if mod_parts.len() > 1 {
                    mod_parts.pop();
                }
            }
            let abs_path = if ty_remainder.is_empty() {
                mod_parts.join("::")
            } else {
                format!("{}::{}", mod_parts.join("::"), ty_remainder)
            };
            return abs_path;
        }
        if ty_clean.starts_with("self::") {
            // self:: just means current_mod
            let ty_remainder = &ty_clean[6..];
            return format!("{}::{}", current_mod, ty_remainder);
        }
        if ty_clean.starts_with("crate::")
            || ty_clean.starts_with("::")
            || ty_clean.starts_with("prost::")
            || ty_clean.starts_with("core::")
            || ty_clean.starts_with("std::")
            || ty_clean.starts_with("alloc::")
        {
            ty_clean
        } else if primitives.iter().any(|&p| ty_clean == p)
            || ty_clean.starts_with("Option")
            || ty_clean.starts_with("Vec")
            || ty_clean.starts_with("HashMap")
        {
            ty_clean
        } else if ty_clean.contains("::") {
            // Prepend current_mod for relative paths like enriched_identity::Principal
            format!("{}::{}", current_mod, ty_clean)
        } else {
            format!("{}::{}", current_mod, ty_clean.replace(' ', ""))
        }
    }
    let mut py_methods = String::new();
    for info in infos {
        let fq = info
            .fq_name
            .replace("crate :: ", "crate::")
            .replace(" :: ", "::");
        let fields = &info.fields;
        let is_tuple = info.is_tuple;
        let mod_path = &info.mod_path;
        let py_new = if fields.is_empty() {
            format!(
                "        #[new]\n        pub fn py_new() -> Self {{\n            Self::default()\n        }}\n"
            )
        } else if is_tuple {
            let args = fields
                .iter()
                .map(|(n, t)| format!("{}: {}", n, qualify_type(t, mod_path)))
                .collect::<Vec<_>>()
                .join(", ");
            let vals = fields
                .iter()
                .map(|(n, _)| n.clone())
                .collect::<Vec<_>>()
                .join(", ");
            format!(
                "        #[new]\n        pub fn py_new({}) -> Self {{\n            Self({})\n        }}\n",
                args, vals
            )
        } else {
            let args = fields
                .iter()
                .map(|(n, t)| format!("{}: {}", n, qualify_type(t, mod_path)))
                .collect::<Vec<_>>()
                .join(", ");
            let vals = fields
                .iter()
                .map(|(n, _)| format!("{}", n))
                .collect::<Vec<_>>()
                .join(", ");
            format!(
                "        #[new]\n        pub fn py_new({}) -> Self {{\n            Self {{ {} }}\n        }}\n",
                args, vals
            )
        };
        py_methods += &format!(
            r#"
        #[::pyo3::pymethods]
        impl {fq} {{
{py_new}
            fn __repr__(&self) -> String {{
                format!("{{:?}}", self)
            }}
            fn __str__(&self) -> String {{
                format!("{{:?}}", self)
            }}
        }}
"#,
            fq = fq,
            py_new = py_new
        );
    }

    let syntax = syn::parse_file(&py_methods.to_string()).expect("Failed to parse generated code");
    unparse(&syntax)
}

#[derive(Default)]
struct ModuleNode {
    children: std::collections::HashMap<String, ModuleNode>,
    types: Vec<String>, // store fully qualified type paths
}

fn build_module_tree(
    types: &[Type],
    all_types: &mut std::collections::BTreeSet<String>,
) -> ModuleNode {
    let mut root = ModuleNode::default();
    for ty in types {
        let fq = quote!(#ty)
            .to_string()
            .replace("crate :: ", "crate::")
            .replace(" :: ", "::");
        let parts: Vec<_> = fq.split("::").map(|s| s.trim().to_string()).collect();
        if parts.len() < 2 {
            continue;
        }
        let (mod_path, type_name) = parts.split_at(parts.len() - 1);
        let mut node = &mut root;
        for part in mod_path {
            node = node.children.entry(part.clone()).or_default();
        }
        let full_type = parts.join("::");
        node.types.push(full_type.clone());
        all_types.insert(full_type);
    }
    root
}

fn generate_pymodules_file(
    module_tree: &ModuleNode,
    all_types: &std::collections::BTreeSet<String>,
) -> String {
    let mut code = String::new();
    code += "use pyo3::prelude::*;\n\
    #[pyo3::pymodule]\n";
    //     for ty in all_types {
    //         code += &format!(
    //             "use {};
    // ",
    //             ty
    //         );
    //     }
    code += &generate_pymodules(module_tree, &[]);
    code
}

fn generate_pymodules(node: &ModuleNode, mod_path: &[String]) -> String {
    let mod_name = mod_path
        .last()
        .cloned()
        .unwrap_or_else(|| "cloud".to_string());
    let func_name = if mod_path.is_empty() {
        "init_pymodule".to_string()
    } else {
        format!("init_{}", mod_path.join("_"))
    };
    let mut code = format!(
        "pub fn {}_mod(_py: pyo3::Python, m: &Bound<'_, PyModule>) -> pyo3::PyResult<()> {{\n",
        mod_name
    );
    for ty in &node.types {
        code += &format!("    m.add_class::<crate::{}>()?;\n", ty);
    }
    for (child_name, _child_node) in &node.children {
        let submod_func = format!("{}_mod", child_name.clone());
        code += &format!(
            "    let submod = pyo3::types::PyModule::new(_py, \"{}\")?;\n",
            child_name
        );
        code += &format!("    {}(_py, &submod)?;\n", submod_func);
        code += &format!("    m.add_submodule(&submod)?;\n");
    }
    code += "    Ok(())\n}\n";
    for (child_name, child_node) in &node.children {
        let mut child_path = mod_path.to_vec();
        child_path.push(child_name.clone());
        code += &generate_pymodules(child_node, &child_path);
    }
    code
}

fn main() -> Result<()> {
    let descriptor_path: PathBuf = "descriptors.bin".into();
    println!("cargo:rerun-if-changed={}", descriptor_path.display());

    let mut config = prost_build::Config::new();
    config
        .file_descriptor_set_path(&descriptor_path)
        .compile_well_known_types()
        .disable_comments(["."])
        // .bytes(["."]) // Add prost::bytes::Bytes exclusion
        .type_attribute(".", "#[pyo3::pyclass(dict, get_all, set_all)]")
        // .type_attribute(".", "#[pyo3_prost::pyclass_for_prost_struct]")
        // .type_attribute("bytes::Bytes", "#[pyo3::pyclass(dict, get_all, set_all)]")
        // .type_attribute(
        //     "::pyo3::types::PyBytes",
        //     "#[pyo3::pyclass(dict, get_all, set_all)]",
        // )
        .skip_protoc_run();

    let empty: &[&str] = &[];
    config.compile_protos(empty, empty)?;

    let descriptor_set = std::fs::read(descriptor_path)?;
    pbjson_build::Builder::new()
        .register_descriptors(&descriptor_set)?
        .exclude([
            ".google.protobuf.Duration",
            ".google.protobuf.Timestamp",
            ".google.protobuf.Value",
            ".google.protobuf.Struct",
            ".google.protobuf.ListValue",
            ".google.protobuf.NullValue",
            ".google.protobuf.BoolValue",
            ".google.protobuf.BytesValue",
            ".google.protobuf.DoubleValue",
            ".google.protobuf.FloatValue",
            ".google.protobuf.Int32Value",
            ".google.protobuf.Int64Value",
            ".google.protobuf.StringValue",
            ".google.protobuf.UInt32Value",
            ".google.protobuf.UInt64Value",
        ])
        .build(&[".google"])?;

    // Generate Box<T> implementations
    let boxed_impls = generate_boxed_impls();
    let out_dir = std::env::var("OUT_DIR").unwrap();
    let boxed_impls_path = PathBuf::from(&out_dir).join("boxed_impls.rs");
    let mut file = File::create(boxed_impls_path)?;
    file.write_all(boxed_impls.as_bytes())?;

    // Find all prost types
    let src_dir = PathBuf::from("src");
    let prost_infos = find_prost_types(&src_dir);
    eprintln!("Found {} prost types", prost_infos.len());
    eprintln!("Generating implementations for {:?} types", prost_infos);
    // unsafe { std::intrinsics::breakpoint(); }

    // Generate encode,decode implementations for all prost types
    let serde_impls = generate_encode_decode(&prost_infos);
    let serde_impls_path = PathBuf::from(&out_dir).join("serde_impls.rs");
    let mut file = File::create(serde_impls_path)?;
    file.write_all(serde_impls.as_bytes())?;

    // --- New: Generate PyO3 module tree code ---
    let mut all_types = std::collections::BTreeSet::new();
    let module_tree = build_module_tree(
        &prost_infos.iter().map(|i| i.ty.clone()).collect::<Vec<_>>(),
        &mut all_types,
    );
    let pymodules_code = generate_pymodules_file(&module_tree, &all_types);
    let pymodules_path = PathBuf::from(&out_dir).join("pymodules.rs");
    let mut file = File::create(pymodules_path)?;
    file.write_all(pymodules_code.as_bytes())?;
    // --- End new ---

    Ok(())
}
