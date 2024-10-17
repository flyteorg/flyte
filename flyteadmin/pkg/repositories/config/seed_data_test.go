package config

import (
	"testing"

	"gopkg.in/yaml.v2"
)

func TestSeedProjectUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name        string
		yamlData    string
		expected    []SeedProject
		expectError bool
	}{
		{
			name: "StringInput",
			yamlData: `
- "ProjectA"
- "ProjectB"
`,
			expected: []SeedProject{
				{Name: "ProjectA", Description: ""},
				{Name: "ProjectB", Description: ""},
			},
			expectError: false,
		},
		{
			name: "MapInput",
			yamlData: `
- name: "ProjectC"
  description: "Description for ProjectC"
- name: "ProjectD"
  description: "Description for ProjectD"
`,
			expected: []SeedProject{
				{Name: "ProjectC", Description: "Description for ProjectC"},
				{Name: "ProjectD", Description: "Description for ProjectD"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var projects []SeedProject
			err := yaml.Unmarshal([]byte(tt.yamlData), &projects)
			if (err != nil) != tt.expectError {
				t.Fatalf("Expected error: %v, got error: %v", tt.expectError, err)
			}

			if !tt.expectError {
				if len(projects) != len(tt.expected) {
					t.Fatalf("Expected %d projects, got %d", len(tt.expected), len(projects))
				}

				for i, project := range projects {
					if project != tt.expected[i] {
						t.Errorf("Project at index %d does not match expected.\nGot: %+v\nExpected: %+v", i, project, tt.expected[i])
					}
				}
			}
		})
	}
}
