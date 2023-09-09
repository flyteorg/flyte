package config

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"unsafe"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	PathFlag          = "file"
	StrictModeFlag    = "strict"
	CommandValidate   = "validate"
	CommandDiscover   = "discover"
	CommandDocs       = "docs"
	DocsSectionLength = 120
)

type AccessorProvider func(options Options) Accessor

type printer interface {
	Printf(format string, i ...interface{})
	Println(i ...interface{})
}

func NewConfigCommand(accessorProvider AccessorProvider) *cobra.Command {
	opts := Options{}
	rootCmd := &cobra.Command{
		Use:       "config",
		Short:     "Runs various config commands, look at the help of this command to get a list of available commands..",
		ValidArgs: []string{CommandValidate, CommandDiscover, CommandDocs},
	}

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validates the loaded config.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return validate(accessorProvider(opts), cmd)
		},
	}

	discoverCmd := &cobra.Command{
		Use:   "discover",
		Short: "Searches for a config in one of the default search paths.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return validate(accessorProvider(opts), cmd)
		},
	}

	docsCmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate configuration documetation in rst format",
		RunE: func(cmd *cobra.Command, args []string) error {
			sections := GetRootSection().GetSections()
			orderedSectionKeys := sets.NewString()
			for s := range sections {
				orderedSectionKeys.Insert(s)
			}
			printToc(orderedSectionKeys)
			visitedSection := map[string]bool{}
			visitedType := map[reflect.Type]bool{}
			for _, sectionKey := range orderedSectionKeys.List() {
				if canPrint(sections[sectionKey].GetConfig()) {
					printDocs(sectionKey, false, sections[sectionKey], visitedSection, visitedType)
				}
			}
			return nil
		},
	}

	// Configure Root Command
	rootCmd.PersistentFlags().StringArrayVar(&opts.SearchPaths, PathFlag, []string{}, `Passes the config file to load.
If empty, it'll first search for the config file path then, if found, will load config from there.`)

	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(discoverCmd)
	rootCmd.AddCommand(docsCmd)

	// Configure Validate Command
	validateCmd.Flags().BoolVar(&opts.StrictMode, StrictModeFlag, false, `Validates that all keys in loaded config
map to already registered sections.`)

	return rootCmd
}

// Redirects Stdout to a string buffer until context is cancelled.
func redirectStdOut() (old, new *os.File) {
	old = os.Stdout // keep backup of the real stdout
	var err error
	_, new, err = os.Pipe()
	if err != nil {
		panic(err)
	}

	os.Stdout = new

	return
}

func printDocs(title string, isSubsection bool, section Section, visitedSection map[string]bool, visitedType map[reflect.Type]bool) {
	printTitle(title, isSubsection)
	val := reflect.Indirect(reflect.ValueOf(section.GetConfig()))
	if val.Kind() == reflect.Slice {
		val = reflect.Indirect(reflect.ValueOf(val.Index(0).Interface()))
	}

	subsections := make(map[string]interface{})
	for i := 0; i < val.Type().NumField(); i++ {
		field := val.Type().Field(i)
		tagType := field.Type
		if tagType.Kind() == reflect.Ptr {
			tagType = field.Type.Elem()
		}

		fieldName := getFieldNameFromJSONTag(field)
		fieldTypeString := getFieldTypeString(tagType)
		fieldDefaultValue := getDefaultValue(fmt.Sprintf("%v", reflect.Indirect(val.Field(i))))
		fieldDescription := getFieldDescriptionFromPflag(field)

		subVal := val.Field(i)
		if tagType.Kind() == reflect.Struct {
			// In order to get value from unexported field in struct
			if subVal.Kind() == reflect.Ptr {
				subVal = reflect.NewAt(subVal.Type(), unsafe.Pointer(subVal.UnsafeAddr())).Elem()
			} else {
				subVal = reflect.NewAt(subVal.Type(), unsafe.Pointer(subVal.UnsafeAddr()))
			}
		}

		if tagType.Kind() == reflect.Map || tagType.Kind() == reflect.Slice || tagType.Kind() == reflect.Struct {
			fieldDefaultValue = getDefaultValue(subVal.Interface())
		}

		if tagType.Kind() == reflect.Struct {
			if canPrint(subVal.Interface()) {
				addSubsection(subVal.Interface(), subsections, fieldName, &fieldTypeString, tagType, visitedSection, visitedType)
			}
		}
		printSection(fieldName, fieldTypeString, fieldDefaultValue, fieldDescription, isSubsection)
	}

	if section != nil {
		sections := section.GetSections()
		orderedSectionKeys := sets.NewString()
		for s := range sections {
			orderedSectionKeys.Insert(s)
		}
		for _, sectionKey := range orderedSectionKeys.List() {
			fieldName := sectionKey
			fieldType := reflect.TypeOf(sections[sectionKey].GetConfig())
			fieldTypeString := getFieldTypeString(fieldType)
			fieldDefaultValue := getDefaultValue(sections[sectionKey].GetConfig())

			addSubsection(sections[sectionKey].GetConfig(), subsections, fieldName, &fieldTypeString, fieldType, visitedSection, visitedType)
			printSection(fieldName, fieldTypeString, fieldDefaultValue, "", isSubsection)
		}
	}
	orderedSectionKeys := sets.NewString()
	for s := range subsections {
		orderedSectionKeys.Insert(s)
	}

	for _, sectionKey := range orderedSectionKeys.List() {
		printDocs(sectionKey, true, NewSection(subsections[sectionKey], nil), visitedSection, visitedType)
	}
}

// Print Table of contents
func printToc(orderedSectionKeys sets.String) {
	for _, sectionKey := range orderedSectionKeys.List() {
		fmt.Printf("- `%s <#section-%s>`_\n\n", sectionKey, sectionKey)
	}
}

func printTitle(title string, isSubsection bool) {
	if isSubsection {
		fmt.Println(title)
		fmt.Println(strings.Repeat("^", DocsSectionLength))
	} else {
		fmt.Println("Section:", title)
		fmt.Println(strings.Repeat("=", DocsSectionLength))
	}
	fmt.Println()
}

func printSection(name string, dataType string, defaultValue string, description string, isSubsection bool) {
	c := "-"
	if isSubsection {
		c = "\""
	}

	fmt.Printf("%s ", name)
	fmt.Printf("(%s)\n", dataType)
	fmt.Println(strings.Repeat(c, DocsSectionLength))
	fmt.Println()
	if description != "" {
		fmt.Printf("%s\n\n", description)
	}
	if defaultValue != "" {
		val := strings.Replace(defaultValue, "\n", "\n  ", -1)
		val = ".. code-block:: yaml\n\n  " + val
		fmt.Printf("**Default Value**: \n\n%s\n", val)
	}
	fmt.Println()
}

func addSubsection(val interface{}, subsections map[string]interface{}, fieldName string,
	fieldTypeString *string, fieldType reflect.Type, visitedSection map[string]bool, visitedType map[reflect.Type]bool) {

	if visitedSection[*fieldTypeString] {
		if !visitedType[fieldType] {
			// Some types have the same name, but they are different type.
			// Add field name at the end to tell the difference between them.
			*fieldTypeString = fmt.Sprintf("%s (%s)", *fieldTypeString, fieldName)
			subsections[*fieldTypeString] = val
		}
	} else {
		visitedSection[*fieldTypeString] = true
		subsections[*fieldTypeString] = val
	}
	*fieldTypeString = fmt.Sprintf("`%s`_", *fieldTypeString)
	visitedType[fieldType] = true
}

func getDefaultValue(val interface{}) string {
	defaultValue, err := yaml.Marshal(val)
	if err != nil {
		return ""
	}
	DefaultValue := string(defaultValue)
	return DefaultValue
}

func getFieldTypeString(tagType reflect.Type) string {
	kind := tagType.Kind()
	if kind == reflect.Ptr {
		tagType = tagType.Elem()
		kind = tagType.Kind()
	}

	FieldTypeString := kind.String()
	if kind == reflect.Map || kind == reflect.Slice || kind == reflect.Struct {
		FieldTypeString = tagType.String()
	}
	return FieldTypeString
}

func getFieldDescriptionFromPflag(field reflect.StructField) string {
	if pFlag := field.Tag.Get("pflag"); len(pFlag) > 0 && !strings.HasPrefix(pFlag, "-") {
		var commaIdx int
		if commaIdx = strings.Index(pFlag, ","); commaIdx < 0 {
			commaIdx = -1
		}
		if strippedDescription := pFlag[commaIdx+1:]; len(strippedDescription) > 0 {
			return strings.TrimPrefix(strippedDescription, " ")
		}
	}
	return ""
}

func getFieldNameFromJSONTag(field reflect.StructField) string {
	if jsonTag := field.Tag.Get("json"); len(jsonTag) > 0 && !strings.HasPrefix(jsonTag, "-") {
		var commaIdx int
		if commaIdx = strings.Index(jsonTag, ","); commaIdx < 0 {
			commaIdx = len(jsonTag)
		}
		if strippedName := jsonTag[:commaIdx]; len(strippedName) > 0 {
			return strippedName
		}
	}
	return field.Name
}

// Print out config docs if and only if the section type is struct or slice
func canPrint(b interface{}) bool {
	val := reflect.Indirect(reflect.ValueOf(b))
	if val.Kind() == reflect.Struct || val.Kind() == reflect.Slice {
		return true
	}
	return false
}

func validate(accessor Accessor, p printer) error {
	// Redirect stdout
	old, n := redirectStdOut()
	defer func() {
		err := n.Close()
		if err != nil {
			panic(err)
		}
	}()
	defer func() { os.Stdout = old }()

	err := accessor.UpdateConfig(context.Background())

	printInfo(p, accessor)
	if err == nil {
		green := color.New(color.FgGreen).SprintFunc()
		p.Println(green("Validated config file successfully."))
	} else {
		red := color.New(color.FgRed).SprintFunc()
		p.Println(red("Failed to validate config file."))
	}

	return err
}

func printInfo(p printer, v Accessor) {
	cfgFile := v.ConfigFilesUsed()
	if len(cfgFile) != 0 {
		green := color.New(color.FgGreen).SprintFunc()

		p.Printf("Config file(s) found at: %v\n", green(strings.Join(cfgFile, "\n")))
	} else {
		red := color.New(color.FgRed).SprintFunc()
		p.Println(red("Couldn't find a config file."))
	}
}
