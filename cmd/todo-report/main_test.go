package main

import (
	"strings"
	"testing"
)

func TestRunErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "no args",
			args: nil,
			want: "usage:",
		},
		{
			name: "unknown command",
			args: []string{"bogus"},
			want: "usage:",
		},
		{
			name: "age missing branch",
			args: []string{"age"},
			want: "age requires --branch",
		},
		{
			name: "drift missing branches",
			args: []string{"drift"},
			want: "drift requires --branch-a and --branch-b",
		},
		{
			name: "lint unsupported format",
			args: []string{"lint", "--branch", "main", "--format", "xml"},
			want: `unsupported format "xml"`,
		},
		{
			name: "health missing branch",
			args: []string{"health"},
			want: "health requires --branch",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := run(tc.args)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q in %q", tc.want, err.Error())
			}
		})
	}
}

func TestNormalizeFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		jsonFlag bool
		want     string
		wantErr  string
	}{
		{name: "lowercases text", format: "TEXT", want: "text"},
		{name: "trims markdown", format: " markdown ", want: "markdown"},
		{name: "json flag wins", format: "tsv", jsonFlag: true, want: "json"},
		{name: "bad format", format: "yaml", wantErr: `unsupported format "yaml"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeFormat(tc.format, tc.jsonFlag)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected %q in %q", tc.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
