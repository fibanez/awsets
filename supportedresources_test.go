package awsets

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/trek10inc/awsets/resource"
)

// Tests to make sure the documentation is up to date with the current list of supported resource types
func Test_SupportResources(t *testing.T) {
	// Read documented list of resource types
	b, err := ioutil.ReadFile("supported_resources.txt")
	if err != nil {
		t.Fatalf("failed to load file: %v", err)
	}
	// Build map of types
	supported := make(map[resource.ResourceType]struct{})
	for _, l := range strings.Split(string(b), "\n") {
		supported[resource.ResourceType(l)] = struct{}{}
	}

	// Iterate through resource types supported in code. For each. check if it is in the documentation
	// If it is, remove it
	// If it isn't, append it to a list of resources that need added
	needsAdded := make([]resource.ResourceType, 0)
	for _, at := range Types(nil, nil) {
		_, exists := supported[at]
		if exists {
			delete(supported, at)
		} else {
			needsAdded = append(needsAdded, at)
		}
	}

	// If resources are missing from documentation, fail test & print them
	if len(needsAdded) > 0 {
		t.Fatalf("the following resource types needed added to supported_types.txt: %v\n", needsAdded)
	}
	// If resources are in documentation that are NOT supported in code, fail test and print them
	if len(supported) > 0 {
		t.Fatalf("the following resource types need removed from supported_types.txt: %v\n", supported)
	}
}
