package bubbletea

import (
	"strings"
)

var commandFlagMap = make(map[string][]string)

func InitCommandFlagMap() {
	// compile
	commandFlagMap[sliceToString([]string{"compile"})] = []string{"--file"}
	// completion
	commandFlagMap[sliceToString([]string{"completion"})] = []string{"--shell"}
	// config
	commandFlagMap[sliceToString([]string{"config", "discover"})] = []string{""}
	commandFlagMap[sliceToString([]string{"config", "docs"})] = []string{""}
	commandFlagMap[sliceToString([]string{"config", "init"})] = []string{""}
	commandFlagMap[sliceToString([]string{"config", "validate"})] = []string{""}
	// create
	commandFlagMap[sliceToString([]string{"create", "execution"})] = []string{"-p", "-d"}
	commandFlagMap[sliceToString([]string{"create", "project"})] = []string{"--pid"} //?
	// delete
	commandFlagMap[sliceToString([]string{"delete", "cluster-resource-attribute"})] = []string{""} //?
	commandFlagMap[sliceToString([]string{"delete", "execution"})] = []string{"-p", "-d"}
	commandFlagMap[sliceToString([]string{"delete", "execution-cluster-label"})] = []string{""}   //?
	commandFlagMap[sliceToString([]string{"delete", "execution-queue-attribute"})] = []string{""} //?
	commandFlagMap[sliceToString([]string{"delete", "plugin-override"})] = []string{""}           //?
	commandFlagMap[sliceToString([]string{"delete", "task-resource-attribute"})] = []string{""}   //?
	commandFlagMap[sliceToString([]string{"delete", "workflow-execution-config"})] = []string{""} //?
	// demo
	commandFlagMap[sliceToString([]string{"demo", "exec"})] = []string{""} //?
	commandFlagMap[sliceToString([]string{"demo", "reload"})] = []string{""}
	commandFlagMap[sliceToString([]string{"demo", "start"})] = []string{""}
	commandFlagMap[sliceToString([]string{"demo", "status"})] = []string{""}
	commandFlagMap[sliceToString([]string{"demo", "teardown"})] = []string{""}
	// get
	commandFlagMap[sliceToString([]string{"get", "cluster-resource-attribute"})] = []string{"-p", "-d"}
	commandFlagMap[sliceToString([]string{"get", "execution"})] = []string{"-p", "-d"}
	commandFlagMap[sliceToString([]string{"get", "execution-cluster-label"})] = []string{"-p", "-d"}
	commandFlagMap[sliceToString([]string{"get", "launchplan"})] = []string{"-p", "-d"}
	commandFlagMap[sliceToString([]string{"get", "plugin-override"})] = []string{"-p", "-d"}
	commandFlagMap[sliceToString([]string{"get", "project"})] = []string{}
	commandFlagMap[sliceToString([]string{"get", "task"})] = []string{"-p", "-d"}
	commandFlagMap[sliceToString([]string{"get", "task-resource-attribute"})] = []string{"-p", "-d"}
	commandFlagMap[sliceToString([]string{"get", "workflow"})] = []string{"-p", "-d"}
	commandFlagMap[sliceToString([]string{"get", "workflow-execution-config"})] = []string{}
	// register
	commandFlagMap[sliceToString([]string{"register", "examples"})] = []string{}
	commandFlagMap[sliceToString([]string{"register", "files"})] = []string{}
	// sandbox
	commandFlagMap[sliceToString([]string{"sandbox", "exec"})] = []string{} //?
	commandFlagMap[sliceToString([]string{"sandbox", "start"})] = []string{}
	commandFlagMap[sliceToString([]string{"sandbox", "status"})] = []string{}
	commandFlagMap[sliceToString([]string{"sandbox", "teardown"})] = []string{}
	// update
	commandFlagMap[sliceToString([]string{"update", "cluster-resource-attribute"})] = []string{"--attrFile"} //?
	commandFlagMap[sliceToString([]string{"update", "execution"})] = []string{"-p", "-d"}
	commandFlagMap[sliceToString([]string{"update", "execution-cluster-label"})] = []string{"--attrFile"}   //?
	commandFlagMap[sliceToString([]string{"update", "execution-queue-attribute"})] = []string{"--attrFile"} //?
	commandFlagMap[sliceToString([]string{"update", "launchplan"})] = []string{"-p", "-d"}
	commandFlagMap[sliceToString([]string{"update", "launchplan-meta"})] = []string{"-p", "-d"}
	commandFlagMap[sliceToString([]string{"update", "plugin-override"})] = []string{"--attrFile"} //?
	commandFlagMap[sliceToString([]string{"update", "project"})] = []string{"--pid"}
	commandFlagMap[sliceToString([]string{"update", "task-meta"})] = []string{"-p", "-d"}
	commandFlagMap[sliceToString([]string{"update", "task-resource-attribute"})] = []string{"--attrFile"}
	commandFlagMap[sliceToString([]string{"update", "workflow-execution-config"})] = []string{"--attrFile"}
	commandFlagMap[sliceToString([]string{"update", "workflow-meta"})] = []string{"-p", "-d"}
	// upgrade
	commandFlagMap[sliceToString([]string{"upgrade"})] = []string{}
	// version
	commandFlagMap[sliceToString([]string{"version"})] = []string{}
}

// sliceToString converts a slice of strings to a string representation
func sliceToString(slice []string) string {
	return strings.Join(slice, "|")
}
