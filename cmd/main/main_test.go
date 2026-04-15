package main

import (
	"testing"
	"xray-exporter/internal/pkg/app"
)

func TestParseBasicAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		user    string
		pass    string
		wantErr bool
	}{
		{name: "empty", raw: "", user: "", pass: "", wantErr: false},
		{name: "valid", raw: "alice:secret", user: "alice", pass: "secret", wantErr: false},
		{name: "first colon split", raw: "alice:sec:ret", user: "alice", pass: "sec:ret", wantErr: false},
		{name: "invalid", raw: "alice", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user, pass, err := parseBasicAuth(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseBasicAuth() err = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if user != tt.user || pass != tt.pass {
				t.Fatalf("parseBasicAuth() = (%q, %q), want (%q, %q)", user, pass, tt.user, tt.pass)
			}
		})
	}
}

func TestRootCommandPreRunParsesBasicAuth(t *testing.T) {
	t.Parallel()

	cmd := newRootCommand(app.BuildInfo{Version: "test", Commit: "test", Date: "test"})
	if err := cmd.Flags().Set("basicauth", "john:doe"); err != nil {
		t.Fatalf("Flags().Set() err = %v", err)
	}

	if err := cmd.PreRunE(cmd, nil); err != nil {
		t.Fatalf("PreRunE() err = %v", err)
	}
}

func TestRootCommandPreRunRejectsInvalidBasicAuth(t *testing.T) {
	t.Parallel()

	cmd := newRootCommand(app.BuildInfo{Version: "test", Commit: "test", Date: "test"})
	if err := cmd.Flags().Set("basicauth", "invalid"); err != nil {
		t.Fatalf("Flags().Set() err = %v", err)
	}

	if err := cmd.PreRunE(cmd, nil); err == nil {
		t.Fatal("PreRunE() expected error for invalid basic auth")
	}
}
