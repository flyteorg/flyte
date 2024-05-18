package bubbletea

import (
	"strings"
)

var commandFlagMap = make(map[string][]string)

func InitCommandFlagMap() {
	// compile
	commandFlagMap["compile"] = []string{"--file"}
	// completion
	commandFlagMap["completion"] = []string{"--shell"}
	// config
	commandFlagMap["config|discover"] = []string{""}
	commandFlagMap["config|docs"] = []string{""}
	commandFlagMap["config|init"] = []string{""}
	commandFlagMap["config|validate"] = []string{""}
	// create
	commandFlagMap["create|execution"] = []string{"-p", "-d"}
	commandFlagMap["create|project"] = []string{"--pid"} //?
	// delete
	commandFlagMap["delete|cluster-resource-attribute"] = []string{""} //?
	commandFlagMap["delete|execution"] = []string{"-p", "-d"}
	commandFlagMap["delete|execution-cluster-label"] = []string{""}   //?
	commandFlagMap["delete|execution-queue-attribute"] = []string{""} //?
	commandFlagMap["delete|plugin-override"] = []string{""}           //?
	commandFlagMap["delete|task-resource-attribute"] = []string{""}   //?
	commandFlagMap["delete|workflow-execution-config"] = []string{""} //?
	// demo
	commandFlagMap["demo|exec"] = []string{""} //?
	commandFlagMap["demo|reload"] = []string{""}
	commandFlagMap["demo|start"] = []string{""}
	commandFlagMap["demo|status"] = []string{""}
	commandFlagMap["demo|teardown"] = []string{""}
	// get
	commandFlagMap["get|cluster-resource-attribute"] = []string{"-p", "-d"}
	commandFlagMap["get|execution"] = []string{"-p", "-d"}
	commandFlagMap["get|execution-cluster-label"] = []string{"-p", "-d"}
	commandFlagMap["get|launchplan"] = []string{"-p", "-d"}
	commandFlagMap["get|plugin-override"] = []string{"-p", "-d"}
	commandFlagMap["get|project"] = []string{}
	commandFlagMap["get|task"] = []string{"-p", "-d"}
	commandFlagMap["get|task-resource-attribute"] = []string{"-p", "-d"}
	commandFlagMap["get|workflow"] = []string{"-p", "-d"}
	commandFlagMap["get|workflow-execution-config"] = []string{}
	// register
	commandFlagMap["register|examples"] = []string{}
	commandFlagMap["register|files"] = []string{}
	// sandbox
	commandFlagMap["sandbox|exec"] = []string{} //?
	commandFlagMap["sandbox|start"] = []string{}
	commandFlagMap["sandbox|status"] = []string{}
	commandFlagMap["sandbox|teardown"] = []string{}
	// update
	commandFlagMap["update|cluster-resource-attribute"] = []string{"--attrFile"} //?
	commandFlagMap["update|execution"] = []string{"-p", "-d"}
	commandFlagMap["update|execution-cluster-label"] = []string{"--attrFile"}   //?
	commandFlagMap["update|execution-queue-attribute"] = []string{"--attrFile"} //?
	commandFlagMap["update|launchplan"] = []string{"-p", "-d"}
	commandFlagMap["update|launchplan-meta"] = []string{"-p", "-d"}
	commandFlagMap["update|plugin-override"] = []string{"--attrFile"} //?
	commandFlagMap["update|project"] = []string{"--pid"}
	commandFlagMap["update|task-meta"] = []string{"-p", "-d"}
	commandFlagMap["update|task-resource-attribute"] = []string{"--attrFile"}
	commandFlagMap["update|workflow-execution-config"] = []string{"--attrFile"}
	commandFlagMap["update|workflow-meta"] = []string{"-p", "-d"}
	// upgrade
	commandFlagMap["upgrade"] = []string{}
	// version
	commandFlagMap["version"] = []string{}
}

// sliceToString converts a slice of strings to a string representation
func sliceToString(slice []string) string {
	return strings.Join(slice, "|")
}
