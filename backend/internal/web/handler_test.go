package web

import "testing"

func TestPageNameFromPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "root path",
			path: "/",
			want: "home",
		},
		{
			name: "empty path",
			path: "",
			want: "home",
		},
		{
			name: "single segment",
			path: "/papers",
			want: "papers",
		},
		{
			name: "nested path",
			path: "/stacks/my",
			want: "stacks/my",
		},
		{
			name: "trailing slash",
			path: "/stacks/",
			want: "stacks",
		},
		{
			name: "no leading slash",
			path: "settings",
			want: "settings",
		},
		{
			name: "no query param",
			path: "/stacks/my?sort=desc",
			want: "stacks/my",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := pageNameFromPath(tt.path)
			if got != tt.want {
				t.Fatalf("pageNameFromPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
