// This package is a central repository of all compile errors that can be reported. It contains ways to collect and format
// errors to make it easy to find and correct workflow spec problems.
package errors

import "fmt"

type ErrorCode string

// Represents a compile error for coreWorkflow.
type CompileError struct {
	code        ErrorCode
	nodeID      string
	description string
	source      string
}

// Represents a compile error with a root cause.
type CompileErrorWithCause struct {
	*CompileError
	cause error
}

// A set of Compile errors.
type CompileErrors interface {
	error
	Collect(e ...*CompileError)
	NewScope() CompileErrors
	Errors() *compileErrorSet
	HasErrors() bool
	ErrorCount() int
}

type compileErrors struct {
	errorSet          *compileErrorSet
	parent            CompileErrors
	errorCountInScope int
}

// Gets the Compile Error code
func (err CompileError) Code() ErrorCode {
	return err.code
}

// Gets a readable/formatted string explaining the compile error as well as at which node it occurred.
func (err CompileError) Error() string {
	source := ""
	if err.source != "" {
		source = fmt.Sprintf("[%v] ", err.source)
	}

	return fmt.Sprintf("%vCode: %s, Node Id: %s, Description: %s", source, err.code, err.nodeID, err.description)
}

// Gets a readable/formatted string explaining the compile error as well as at which node it occurred.
func (err CompileErrorWithCause) Error() string {
	cause := ""
	if err.cause != nil {
		cause = fmt.Sprintf(", Cause: %v", err.cause.Error())
	}

	return fmt.Sprintf("%v%v", err.CompileError.Error(), cause)
}

// Exposes the set of unique errors.
func (errs *compileErrors) Errors() *compileErrorSet {
	return errs.errorSet
}

// Appends a compile error to the set.
func (errs *compileErrors) Collect(e ...*CompileError) {
	if e != nil {
		if GetConfig().PanicOnError {
			panic(e)
		}

		if errs.parent != nil {
			errs.parent.Collect(e...)
			errs.errorCountInScope += len(e)
		} else {
			for _, err := range e {
				if err != nil {
					errs.errorSet.Put(*err)
					errs.errorCountInScope++
				}
			}
		}
	}
}

// Creates a new scope for compile errors. Parent scope will always automatically collect errors reported in any of its
// child scopes.
func (errs *compileErrors) NewScope() CompileErrors {
	return &compileErrors{parent: errs}
}

// Gets a formatted string of all compile errors collected.
func (errs *compileErrors) Error() (err string) {
	if errs.parent != nil {
		return errs.parent.Error()
	}

	err = fmt.Sprintf("Collected Errors: %v\n", len(*errs.Errors()))
	i := 0
	for _, e := range errs.Errors().List() {
		err += fmt.Sprintf("\tError %d: %s\n", i, e.Error())
		i++
	}

	return err
}

// Gets a value indicating whether there are any errors collected within current scope and all of its children.
func (errs *compileErrors) HasErrors() bool {
	return errs.errorCountInScope > 0
}

// Gets the number of errors collected within current scope and all of its children.
func (errs *compileErrors) ErrorCount() int {
	return errs.errorCountInScope
}

// Creates a new empty compile errors
func NewCompileErrors() CompileErrors {
	return &compileErrors{errorSet: &compileErrorSet{}}
}
