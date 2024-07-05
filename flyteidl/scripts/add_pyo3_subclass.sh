sed '/pub struct /i\
#[pyo3(subclass)]\
' ./gen/pb_rust/google.protobuf.rs > temp_file && mv temp_file ./gen/pb_rust/google.protobuf.rs


sed '/pub struct /i\
#[pyo3(subclass)]\
' ./gen/pb_rust/flyteidl.core.rs > temp_file && mv temp_file ./gen/pb_rust/flyteidl.core.rs

sed '/pub struct /i\
#[pyo3(subclass)]\
' ./gen/pb_rust/flyteidl.admin.rs > temp_file && mv temp_file ./gen/pb_rust/flyteidl.admin.rs
