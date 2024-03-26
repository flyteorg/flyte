package single

import "testing"

func TestGetConsoleFile(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"/console", "dist/index.html"},
		{"/console/", "dist/index.html"},
		{"/console/assets/xyz.png", "dist/xyz.png"},
		{"/console/assets/dir/xyz.png", "dist/dir/xyz.png"},
		{"/console/projects/flytesnacks/workflows?domain=development", "dist/index.html"},
		{"/console/select-project", "dist/index.html"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetConsoleFile(tt.name); got != tt.want {
				t.Errorf("GetConsoleFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
