# Why adding this auxiliary scripts?
# In prost builder configuration (the setting of protobuf compiler we used in `gen.rs`)
# We can only add `subclass` attributes to all Rust types, however enum is a subset of types.
# So, we can't specify PyO3's `subclass` attributes only for Rust structures.
# It's non-trivial to distinguish between these while adding attributes.
# Remember enum is not subclass-able in PyO3. Finally we're using this scipt to add `subclass` attributes only for public Rust structures.

sed '/pub struct /i\
#[pyo3(subclass)]\
' ./gen/pb_rust/google.protobuf.rs > temp_file && mv temp_file ./gen/pb_rust/google.protobuf.rs


sed '/pub struct /i\
#[pyo3(subclass)]\
' ./gen/pb_rust/flyteidl.core.rs > temp_file && mv temp_file ./gen/pb_rust/flyteidl.core.rs

sed '/pub struct /i\
#[pyo3(subclass)]\
' ./gen/pb_rust/flyteidl.admin.rs > temp_file && mv temp_file ./gen/pb_rust/flyteidl.admin.rs
