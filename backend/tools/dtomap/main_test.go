package main

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T, model, api string) (string, string, string) {
	t.Helper()
	directory := t.TempDir()
	domainFile, apiFile, output := filepath.Join(directory, "entities.go"), filepath.Join(directory, "api.go"), filepath.Join(directory, "mapping.go")
	if err := os.WriteFile(domainFile, []byte(model), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(apiFile, []byte(api), 0600); err != nil {
		t.Fatal(err)
	}
	return domainFile, apiFile, output
}

func TestGenerateTypedMappingDeterministicallyWithoutChangingInputs(t *testing.T) {
	const domain = `package model
import "github.com/google/uuid"
type State string
const Active State="active"
type Child struct {
	ID uuid.UUID
	Email string
}
type Input struct {
	Child Child
	Children []Child
	State *State
	Note *string
}
`
	const api = `package api
import openapi_types "github.com/oapi-codegen/runtime/types"
type State string
const Enabled State="active"
type Child struct {
	ID openapi_types.UUID
	Email openapi_types.Email
}
type Input struct {
	Child Child
	Children []Child
	State *State
	Note *string
}
`
	domainFile, apiFile, output := fixture(t, domain, api)
	if err := run(domainFile, apiFile, output); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if err := run(domainFile, apiFile, output); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("generation is not deterministic")
	}
	if _, err := parser.ParseFile(token.NewFileSet(), output, first, 0); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"func toModelInput(v contract.Input) model.Input", "return model.Input{", "ptr.Map(", "slice.Map(", "string(v.Email)", "types.Email(v.Email)", "model.State(v)", "contract.State(v)"} {
		if !strings.Contains(string(first), expected) {
			t.Errorf("missing typed conversion %q", expected)
		}
	}
	for filename, want := range map[string]string{domainFile: domain, apiFile: api} {
		got, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("generator changed input %s", filename)
		}
	}
}

func TestRejectContractDriftBeforeWritingOutput(t *testing.T) {
	for _, tc := range []struct {
		name   string
		domain string
		api    string
		want   string
	}{
		{"optional API field", `package model
type Input struct {
	Name string
}`, `package api
type Input struct {
	Name string
	Added *string
}`, "Input.Added: API field is absent from domain"},
		{"domain field", `package model
type Input struct {
	Name string
	Added *string
}`, `package api
type Input struct {
	Name string
}`, "Input.Added: domain field is absent from API"},
		{"optional shape", `package model
type Input struct {
	Name string
}`, `package api
type Input struct {
	Name *string
}`, "Input.Name: type shape differs"},
		{"enum values", `package model; type State string; const Active State="active"`, `package api; type State string; const Active State="inactive"`, "State: enum values differ"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			domainFile, apiFile, output := fixture(t, tc.domain, tc.api)
			const previous = "previous mapping must be preserved"
			if err := os.WriteFile(output, []byte(previous), 0600); err != nil {
				t.Fatal(err)
			}
			err := run(domainFile, apiFile, output)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, want %q", err, tc.want)
			}
			got, readErr := os.ReadFile(output)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if string(got) != previous {
				t.Fatal("failed generation changed existing mapping")
			}
		})
	}
}
