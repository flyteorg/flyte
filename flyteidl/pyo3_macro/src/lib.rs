extern crate proc_macro;
use proc_macro::TokenStream;
use quote::quote;
use syn::{
    parse_macro_input, token::Brace, AngleBracketedGenericArguments, Data, DeriveInput, Fields,
    FnArg, GenericArgument, ImplItem, ImplItemMethod, Item, ItemEnum, ItemImpl, ItemMod,
    ItemStruct, PathArguments, ReturnType, Type, TypePath, Visibility,
};

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

            let all_values = quote! { #(#field_names),*};

            let required_part = if !required_field_names.is_empty() {
                quote! { #(#required_field_names: #required_field_types),* }
            } else {
                quote! {}
            };
            let optional_part = if !optional_field_names.is_empty() {
                quote! { #(#optional_field_names: #optional_field_types),* }
            } else {
                quote! {}
            };

            let combined_arguments =
                if !required_field_names.is_empty() && !optional_field_names.is_empty() {
                    quote! { #required_part, #optional_part }
                } else {
                    quote! { #required_part #optional_part }
                };
            let combined_signatures =
                if !required_field_names.is_empty() && !optional_field_names.is_empty() {
                    quote! { #(#required_field_names),*, #(#optional_field_names = None),* }
                } else if !required_field_names.is_empty() {
                    quote! { #(#required_field_names),* }
                } else {
                    quote! { #(#optional_field_names = None),* }
                };

            if generic_params.is_empty() {
                // Implement methods template of the `new()` function
                quote! {

                    use pyo3::prelude::*;
                    #[pyo3::pymethods]
                    impl #name {
                        // By default, it is not possible to create an instance of a custom class from Python code.
                        // To declare a constructor, you need to define a method and annotate it with the #[new] attribute.
                        // https://pyo3.rs/v0.21.2/class#constructor

                        // Most arguments are required by default, except for trailing Option<_> arguments, which are implicitly given a default of None.
                        // This behaviour can be configured by the #[pyo3(signature = (...))] option which allows writing a signature in Python syntax.
                        // https://pyo3.rs/v0.21.2/function/signature#trailing-optional-arguments
                        #[new]
                        #[pyo3(signature = ( #combined_signatures ) )]
                        pub fn new(#combined_arguments) -> Self {
                            Self {
                                #all_values
                            }
                        }

                        // use prost::Message;
                        // use pyo3::types::PyBytes;
                        pub fn ParseFromString(&mut self, bytes_string: &pyo3::types::PyBytes) -> Result<#name, crate::_flyteidl_rust::MessageDecodeError> {
                            let bt = bytes_string.as_bytes();
                            let de = prost::Message::decode(&bt.to_vec()[..]);
                            Ok(de?)
                        }
                        // TODO:
                        // pub fn SerializeToString(proto_obj: #name) -> Result<Vec<u8>, crate::flyteidl::MessageEncodeError> {
                        //     let mut buf = vec![];
                        //     proto_obj.encode(&mut buf)?;
                        //     Ok(buf)
                        // }
                    }

                    // https://github.com/hyperium/tonic/blob/c7836521dd417434d625bd653fcf00fb7f7ae25e/tonic/src/request.rs#L28
                }
            } else {
                // https://pyo3.rs/v0.21.2/class#no-generic-parameters
                // TODO: Implement methods template of the `new()` function for generic structs
                quote! {}
            }
        }
        // Data::Struct(data) => {
        //     // Extract the list of field names and types
        //     let (field_names, field_types): (Vec<_>, Vec<_>) = match &data.fields {
        //         Fields::Named(ref fields_named) => fields_named
        //             .named
        //             .iter()
        //             .map(|f| (f.ident.clone(), &f.ty))
        //             .unzip(),
        //         _ => panic!("AutoPyClass can only be used with structs with named fields"),
        //     };

        //     let all_values = quote! { #(#field_names),*};

        //     // Make all fields optional
        //     let optional_field_names: Vec<_> = field_names.iter().collect();
        //     let optional_field_types: Vec<_> = field_types.iter().map(|ty| quote! { #ty }).collect();

        //     let optional_part = if !optional_field_names.is_empty() {
        //         quote! { #(#optional_field_names: #optional_field_types),* }
        //     } else {
        //         quote! {}
        //     };

        //     let combined_signatures = if !optional_field_names.is_empty() {
        //         quote! { #(#optional_field_names = None),* }
        //     } else {
        //         quote! {}
        //     };

        //     if generic_params.is_empty() {
        //         // Implement methods template of the `new()` function
        //         quote! {
        //             #[::pyo3::pymethods]
        //             impl #name {
        //                 #[new]
        //                 #[pyo3(signature = ( #combined_signatures ) )]
        //                 pub fn new(#optional_part) -> Self {
        //                     Self {
        //                         #all_values
        //                     }
        //                 }
        //             }
        //         }
        //     } else {
        //         // TODO: Implement methods template of the `new()` function for generic structs
        //         quote! {}
        //     }
        // }
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
