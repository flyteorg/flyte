extern crate proc_macro;
use proc_macro::TokenStream;
use quote::quote;
use syn::{parse_macro_input, Data, DeriveInput, Fields, Type};

#[proc_macro_derive(WithNew)]
pub fn with_new(input: TokenStream) -> TokenStream {
    // Parse the input tokens into a syntax tree
    let input = parse_macro_input!(input as DeriveInput);
    let name = &input.ident;
    let generics = &input.generics;
    let generic_params: Vec<_> = generics.params.iter().collect();
    // `where_clause` for future traits bounds handling
    let where_clause = &generics.where_clause;

    let gen = match &input.data {
        Data::Struct(data) => {
            // Extract the list of field names and types
            let (field_names, field_types): (Vec<_>, Vec<_>) = match &data.fields {
                Fields::Named(ref fields_named) => fields_named
                    .named
                    .iter()
                    .map(|f| (f.ident.clone(), &f.ty))
                    .unzip(),
                _ => panic!("AutoPyClass can only be used with structs with named fields"),
            };

            // Partition fields into required and optional
            let (required_fields, optional_fields): (Vec<_>, Vec<_>) = field_names.iter()
            .zip(field_types.iter())
            .partition(|(_, ty)| !matches!(ty, Type::Path(type_path) if type_path.path.segments.iter().any(|segment| segment.ident == "Option")));

            let required_field_names: Vec<_> =
                required_fields.iter().map(|(name, _)| name).collect();
            let required_field_types: Vec<_> = required_fields.iter().map(|(_, ty)| ty).collect();

            let optional_field_names: Vec<_> =
                optional_fields.iter().map(|(name, _)| name).collect();
            let optional_field_types: Vec<_> = optional_fields.iter().map(|(_, ty)| ty).collect();

            // Generate the struct's field initialization
            // Use default value for required fields and omit the optional fields
            let all_values = if required_field_names.is_empty() && optional_field_names.is_empty() {
                quote! {}
            } else if required_field_names.is_empty() && !optional_field_names.is_empty() {
                quote! {
                    #(#optional_field_names: #optional_field_names),*
                }
            } else if !required_field_names.is_empty() && optional_field_names.is_empty() {
                quote! {
                    #(#required_field_names: #required_field_names.unwrap_or_default()),*
                }
            } else {
                quote! {
                    #(#required_field_names: #required_field_names.unwrap_or_default()),*,
                    #(#optional_field_names: #optional_field_names),*
                }
            };

            // Make required fields optional, so we can treat trailing Option<T> arguments as having a default of None.
            let required_part = if !required_field_names.is_empty() {
                quote! { #(#required_field_names: Option<#required_field_types>),* }
            } else {
                quote! {}
            };
            let optional_part = if !optional_field_names.is_empty() {
                quote! { #(#optional_field_names: #optional_field_types),* }
            } else {
                quote! {}
            };

            let combined_arguments =
                if required_field_names.is_empty() && optional_field_names.is_empty() {
                    quote! { kwargs: Option<&pyo3::types::PyDict> }
                } else if required_field_names.is_empty() && !optional_field_names.is_empty() {
                    quote! { #optional_part , kwargs: Option<&pyo3::types::PyDict> }
                } else if !required_field_names.is_empty() && optional_field_names.is_empty() {
                    quote! { #required_part , kwargs: Option<&pyo3::types::PyDict> }
                } else {
                    quote! { #required_part, #optional_part , kwargs: Option<&pyo3::types::PyDict> }
                };
            let combined_signatures = if required_field_names.is_empty()
                && optional_field_names.is_empty()
            {
                quote! { **kwargs }
            } else if !required_field_names.is_empty() && optional_field_names.is_empty() {
                quote! { #(#required_field_names),* , **kwargs }
            } else if required_field_names.is_empty() && !optional_field_names.is_empty() {
                quote! { #(#optional_field_names = None),* , **kwargs }
            } else {
                quote! { #(#required_field_names),*, #(#optional_field_names = None),* , **kwargs }
            };

            if generic_params.is_empty() {
                // Implement methods template of the `new()` constructor function
                quote! {
                    // Macro main entrypoint, `#name` is the name of the underlying structure.
                    use pyo3::prelude::*;
                    #[pyo3::pymethods]
                    impl #name {
                        // By default, it is not possible to create an instance of a custom class from Python code.
                        // To declare a constructor, you need to define a method and annotate it with the #[new] attribute.
                        // https://pyo3.rs/v0.21.2/class#constructor

                        // Most arguments are required by default, except for trailing Option<_> arguments, which are implicitly given a default of None.
                        // This behaviour can be configured by the #[pyo3(signature = (...))] option which allows writing a signature in Python syntax.
                        // https://pyo3.rs/v0.21.2/function/signature#trailing-optional-arguments
                        // As a convenience, functions without a #[pyo3(signature = (...))] option will treat trailing Option<T> arguments as having a default of None.
                        #[new]
                        pub fn new(#combined_arguments) -> Self {
                            Self {
                                #all_values
                            }
                        }
                        // For the needs like `load_proto_from_file()`, `write_proto_to_file()` in `flytekit/core/utils.py`
                        pub fn ParseFromString(&mut self, bytes_string: &pyo3::types::PyBytes) -> Result<#name, crate::_flyteidl_rust::MessageDecodeError>
                        {
                            use prost::Message;
                            let bytes = bytes_string.as_bytes();
                            let de = Message::decode(&bytes.to_vec()[..]);
                            Ok(de?)
                        }
                        pub fn SerializeToString(&self, py: Python) -> Result<PyObject, crate::_flyteidl_rust::MessageEncodeError>
                        {
                            // Bring `prost::Message` trait to scope here. Put it in outer impl block will leads to duplicated imports.
                            use prost::Message;
                            let mut buf = vec![];
                            buf.reserve(self.encoded_len());
                            // Unwrap is safe, since we have reserved sufficient capacity in the vector.
                            self.encode(&mut buf).unwrap();
                            Ok(pyo3::types::PyBytes::new_bound(py, &buf).into())
                        }
                    }
                }
            } else {
                // https://pyo3.rs/v0.21.2/class#no-generic-parameters
                // TODO: Implement methods template of the `new()` function for generic structs
                quote! {}
            }
        }
        Data::Enum(_data) => {
            quote! {}
        }
        Data::Union(_data) => {
            quote! {}
        }
        _ => panic!("AutoPyClass can only be used with structs enums and unions"),
    };

    TokenStream::from(gen)
}
