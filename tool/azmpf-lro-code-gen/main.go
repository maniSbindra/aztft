package main

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/magodo/aztft/internal/resmap"
)

//go:embed lroResourceList.txt
var lroResourceListContent string

var lroResources []string

func init() {
	// Populate lroResources from the embedded file
	lroResources = strings.Split(strings.TrimSpace(lroResourceListContent), "\n")
}

func uniqueSlice(slice []string) []string {
	uniqueMap := make(map[string]bool)
	for _, item := range slice {
		uniqueMap[item] = true
	}

	uniqueSlice := make([]string, 0, len(uniqueMap))
	for item := range uniqueMap {
		uniqueSlice = append(uniqueSlice, item)
	}

	return uniqueSlice
}

func main() {
	resmap.Init()

	tfArmMap := resmap.TF2ARMIdMap

	lroResources = uniqueSlice(lroResources)
	var actions []string

	for _, lroResource := range lroResources {
		for tfType, tfArmItem := range tfArmMap {
			if tfArmItem.ManagementPlane != nil && !tfArmItem.IsRemoved && tfType == lroResource {
				if len(tfArmItem.ManagementPlane.ImportSpecs) == 0 {
					continue
				}
				actions = append(actions, getAction(tfArmItem.ManagementPlane.ImportSpecs[0], "/write"))
			}
		}
	}

	actions = uniqueSlice(actions)
	for _, action := range actions {
		fmt.Printf("\t\"%s\": {\n", getAction(action, "/write"))
		fmt.Printf("\t\t\"%s\",\n", getAction(action, "/operationStatuses/read"))
		fmt.Printf("\t},\n")
	}

}

func getAction(importSpec string, actionSfx string) string {
	return fmt.Sprintf("%s%s", strings.Replace(importSpec, "/subscriptions/resourceGroups/", "", 1), actionSfx)
}
