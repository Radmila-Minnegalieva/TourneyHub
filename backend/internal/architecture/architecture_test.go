// Package architecture проверяет правила зависимостей слоёв приложения.
package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const module = "github.com/Radmila-Minnegalieva/TourneyHub/backend/"

func productionFiles(t *testing.T, visit func(string, *token.FileSet, *ast.File)) {
	t.Helper()
	root := filepath.Join("..", "..")
	for _, directory := range []string{"internal", "cmd"} {
		err := filepath.WalkDir(filepath.Join(root, directory), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			visit(filepath.ToSlash(relative), fset, file)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestEnvironmentIsReadOnlyByConfig(t *testing.T) {
	productionFiles(t, func(path string, fset *token.FileSet, file *ast.File) {
		if strings.HasPrefix(path, "internal/config/") {
			return
		}
		packages := map[string]string{}
		for _, im := range file.Imports {
			value, _ := strconv.Unquote(im.Path.Value)
			name := filepath.Base(value)
			if im.Name != nil {
				name = im.Name.Name
			}
			packages[name] = value
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			name, ok := selector.X.(*ast.Ident)
			if !ok {
				return true
			}
			if packages[name.Name] != "os" && packages[name.Name] != "syscall" {
				return true
			}
			switch selector.Sel.Name {
			case "Getenv", "LookupEnv", "Environ", "ExpandEnv", "Setenv", "Unsetenv":
				t.Errorf("%s:%d: environment belongs to config, found %s.%s", path, fset.Position(call.Pos()).Line, name.Name, selector.Sel.Name)
			}
			return true
		})
	})
}

func TestLayerDependenciesAndDataAccess(t *testing.T) {
	sql := regexp.MustCompile(`(?is)^\s*(select\s+.+\s+from\b|insert\s+into\b|update\s+\S+\s+set\b|delete\s+from\b|with\s+\S+\s+as\s*\()`)
	productionFiles(t, func(path string, fset *token.FileSet, file *ast.File) {
		api := strings.HasPrefix(path, "internal/api/")
		business := strings.HasPrefix(path, "internal/service/")
		model := strings.HasPrefix(path, "internal/model/")
		contract := strings.HasPrefix(path, "internal/repository/") && !strings.HasPrefix(path, "internal/repository/postgres/")
		postgres := strings.HasPrefix(path, "internal/repository/postgres/")
		for _, im := range file.Imports {
			value, _ := strconv.Unquote(im.Path.Value)
			driver := strings.HasPrefix(value, "github.com/jackc/pgx") || value == "database/sql"
			transport := strings.HasPrefix(value, "github.com/labstack/echo") || strings.HasPrefix(value, module+"pkg/api/") || strings.HasPrefix(value, module+"internal/api/") || value == "net/http"
			bad := false
			switch {
			case api:
				bad = driver || strings.HasPrefix(value, module+"internal/repository/") || value == module+"internal/repository"
			case business:
				bad = driver || transport
			case model:
				bad = driver || transport || strings.HasPrefix(value, module+"internal/service") || strings.HasPrefix(value, module+"internal/repository") || strings.HasPrefix(value, module+"internal/config")
			case contract:
				bad = driver || transport || strings.HasPrefix(value, module+"internal/service")
			case postgres:
				bad = transport || strings.HasPrefix(value, module+"internal/service")
			}
			if bad {
				t.Errorf("%s: invalid layer dependency %q", path, value)
			}
		}
		if !api && !business && !model && !strings.HasPrefix(path, "cmd/bootstrap/") {
			return
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.BasicLit:
				if node.Kind == token.STRING {
					value, err := strconv.Unquote(node.Value)
					if err == nil && sql.MatchString(value) {
						t.Errorf("%s:%d: SQL belongs to repository/postgres", path, fset.Position(node.Pos()).Line)
					}
				}
			case *ast.CallExpr:
				if selector, ok := node.Fun.(*ast.SelectorExpr); ok {
					switch selector.Sel.Name {
					case "Query", "QueryRow", "Exec", "Scan", "Begin":
						t.Errorf("%s:%d: data access %s belongs to repository/postgres", path, fset.Position(node.Pos()).Line, selector.Sel.Name)
					}
				}
			}
			return true
		})
	})
}
