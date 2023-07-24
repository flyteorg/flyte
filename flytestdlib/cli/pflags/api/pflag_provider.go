package api

import (
	"bytes"
	"fmt"
	"go/types"
	"io/ioutil"
	"os"
	"time"

	"github.com/ernesto-jimenez/gogen/imports"
	goimports "golang.org/x/tools/imports"
)

type PFlagProvider struct {
	typeName        string
	pkg             *types.Package
	fields          []FieldInfo
	pflagValueTypes []PFlagValueType
}

// Imports adds any needed imports for types not directly declared in this package.
func (p PFlagProvider) Imports() map[string]string {
	imp := imports.New(p.pkg.Name())
	for _, m := range p.fields {
		imp.AddImportsFrom(m.Typ)
	}

	return imp.Imports()
}

// WriteCodeFile evaluates the main code file template and writes the output to outputFilePath
func (p PFlagProvider) WriteCodeFile(outputFilePath string) error {
	buf := bytes.Buffer{}
	err := p.generate(GenerateCodeFile, &buf, outputFilePath)
	if err != nil {
		return fmt.Errorf("error generating code, Error: %v. Source: %v", err, buf.String())
	}

	return p.writeToFile(&buf, outputFilePath)
}

// WriteTestFile evaluates the test code file template and writes the output to outputFilePath
func (p PFlagProvider) WriteTestFile(outputFilePath string) error {
	buf := bytes.Buffer{}
	err := p.generate(GenerateTestFile, &buf, outputFilePath)
	if err != nil {
		return fmt.Errorf("error generating code, Error: %v. Source: %v", err, buf.String())
	}

	return p.writeToFile(&buf, outputFilePath)
}

func (p PFlagProvider) writeToFile(buffer *bytes.Buffer, fileName string) error {
	return ioutil.WriteFile(fileName, buffer.Bytes(), os.ModePerm)
}

// generate evaluates the generator and writes the output to buffer. targetFileName is used only to influence how imports are
// generated/optimized.
func (p PFlagProvider) generate(generator func(buffer *bytes.Buffer, info TypeInfo) error, buffer *bytes.Buffer, targetFileName string) error {
	info := TypeInfo{
		Name:            p.typeName,
		Fields:          p.fields,
		Package:         p.pkg.Name(),
		Timestamp:       time.Now(),
		Imports:         p.Imports(),
		PFlagValueTypes: p.pflagValueTypes,
	}

	if err := generator(buffer, info); err != nil {
		return err
	}

	// Update imports
	newBytes, err := goimports.Process(targetFileName, buffer.Bytes(), nil)
	if err != nil {
		return err
	}

	buffer.Reset()
	_, err = buffer.Write(newBytes)

	return err
}

func newPflagProvider(pkg *types.Package, typeName string, fields []FieldInfo, pflagValueTypes []PFlagValueType) PFlagProvider {
	return PFlagProvider{
		typeName:        typeName,
		pkg:             pkg,
		fields:          fields,
		pflagValueTypes: pflagValueTypes,
	}
}
