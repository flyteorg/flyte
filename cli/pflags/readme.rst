================
Pflags Generator
================

This tool enables you to generate code to add pflags for all fields in a struct (recursively). In conjunction with the config package, this can be useful to generate cli flags that overrides configs while maintaing type safety and not having to deal with string typos.

Getting Started
^^^^^^^^^^^^^^^
 - ``go get github.com/lyft/flytestdlib/cli/pflags``
 - call ``pflags <struct_name> --package <pkg_path>`` OR
 - add ``//go:generate pflags <struct_name>`` to the top of the file where the struct is declared.
   <struct_name> has to be a struct type (it can't be, for instance, a slice type). 
   Supported fields' types within the struct: basic types (string, int8, int16, int32, int64, bool), json-unmarshalable types and other structs that conform to the same rules or slices of these types.
 
This generates two files (struct_name_pflags.go and struct_name_pflags_test.go). If you open those, you will notice that all generated flags default to empty/zero values and no usage strings. That behavior can be customized using ``pflag`` tag.

.. code-block::

   type TestType struct {
    StringValue   string         `json:"str" pflag:"\"hello world\",\"life is short\""`
    BoolValue     bool           `json:"bl" pflag:",This is a bool value that will default to false."`
  }

``pflag`` tag is a comma-separated list. First item represents default value. Second value is usage. 
