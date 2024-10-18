package config

import "testing"

func TestMergeSeedProjectsWithUniqueNames(t *testing.T) {
	tests := []struct {
		name                    string
		seedProjects            []string
		seedProjectsWithDetails []SeedProject
		want                    []SeedProject
	}{
		{
			name:                    "Empty inputs",
			seedProjects:            []string{},
			seedProjectsWithDetails: []SeedProject{},
			want:                    []SeedProject{},
		},
		{
			name:                    "Empty inputs",
			seedProjects:            []string{},
			seedProjectsWithDetails: nil,
			want:                    []SeedProject{},
		},
		{
			name:                    "Only seedProjects",
			seedProjects:            []string{"project1", "project2"},
			seedProjectsWithDetails: nil,
			want: []SeedProject{
				{Name: "project1", Description: "project1 description"},
				{Name: "project2", Description: "project2 description"},
			},
		},
		{
			name:         "Only seedProjectsWithDetails",
			seedProjects: []string{},
			seedProjectsWithDetails: []SeedProject{
				{Name: "project1", Description: "custom description 1"},
				{Name: "project2", Description: "custom description 2"},
			},
			want: []SeedProject{
				{Name: "project1", Description: "custom description 1"},
				{Name: "project2", Description: "custom description 2"},
			},
		},
		{
			name:         "Mixed with no overlaps",
			seedProjects: []string{"project1", "project2"},
			seedProjectsWithDetails: []SeedProject{
				{Name: "project3", Description: "custom description 3"},
				{Name: "project4", Description: "custom description 4"},
			},
			want: []SeedProject{
				{Name: "project3", Description: "custom description 3"},
				{Name: "project4", Description: "custom description 4"},
				{Name: "project1", Description: "project1 description"},
				{Name: "project2", Description: "project2 description"},
			},
		},
		{
			name:         "Mixed with overlaps",
			seedProjects: []string{"project1", "project2", "project3"},
			seedProjectsWithDetails: []SeedProject{
				{Name: "project2", Description: "custom description 2"},
				{Name: "project3", Description: "custom description 3"},
			},
			want: []SeedProject{
				{Name: "project2", Description: "custom description 2"},
				{Name: "project3", Description: "custom description 3"},
				{Name: "project1", Description: "project1 description"},
			},
		},
		{
			name:         "Duplicates in seedProjects",
			seedProjects: []string{"project1", "project1", "project2"},
			seedProjectsWithDetails: []SeedProject{
				{Name: "project3", Description: "custom description 3"},
			},
			want: []SeedProject{
				{Name: "project3", Description: "custom description 3"},
				{Name: "project1", Description: "project1 description"},
				{Name: "project2", Description: "project2 description"},
			},
		},
		{
			name:         "Duplicates in seedProjectsWithDetails",
			seedProjects: []string{"project1"},
			seedProjectsWithDetails: []SeedProject{
				{Name: "project2", Description: "custom description 2"},
				{Name: "project2", Description: "duplicate description 2"},
			},
			want: []SeedProject{
				{Name: "project2", Description: "custom description 2"},
				{Name: "project1", Description: "project1 description"},
			},
		},
		{
			name:         "All duplicates",
			seedProjects: []string{"project1", "project1", "project2"},
			seedProjectsWithDetails: []SeedProject{
				{Name: "project1", Description: "custom description 1"},
				{Name: "project2", Description: "custom description 2"},
				{Name: "project2", Description: "duplicate description 2"},
			},
			want: []SeedProject{
				{Name: "project1", Description: "custom description 1"},
				{Name: "project2", Description: "custom description 2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeSeedProjectsWithUniqueNames(tt.seedProjects, tt.seedProjectsWithDetails)

			// Check length
			if len(got) != len(tt.want) {
				t.Errorf("length mismatch: got %d projects, want %d projects", len(got), len(tt.want))
				return
			}

			gotMap := make(map[string]string)
			for _, project := range got {
				gotMap[project.Name] = project.Description
			}
			wantMap := make(map[string]string)
			for _, project := range tt.want {
				wantMap[project.Name] = project.Description
			}

			for name, wantDesc := range wantMap {
				if gotDesc, exists := gotMap[name]; !exists {
					t.Errorf("missing project %q in result", name)
				} else if gotDesc != wantDesc {
					t.Errorf("project %q description mismatch: got %q, want %q", name, gotDesc, wantDesc)
				}
			}

			for name := range gotMap {
				if _, exists := wantMap[name]; !exists {
					t.Errorf("unexpected project %q in result", name)
				}
			}
		})
	}
}
