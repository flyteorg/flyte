package single

import "testing"

func TestGetConsoleFile(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"/console", "dist/index.html"},
		{"/console/", "dist/index.html"},
		{"/console/main.js", "dist/main.js"},
		{"/console/assets/xyz.png", "dist/assets/xyz.png"},
		{"/console/assets/dir/xyz.png", "dist/assets/dir/xyz.png"},
		{"console/projects/flytesnacks/workflows?domain=development", "dist/index.html"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetConsoleFile(tt.name); got != tt.want {
				t.Errorf("GetConsoleFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
