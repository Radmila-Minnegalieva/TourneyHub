// Command metrics собирает воспроизводимые метрики написанного вручную Go-кода.
// Тесты, генерация OpenAPI и сам измеритель исключены из измеряемой области.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/fzipp/gocyclo"
	"github.com/uudashr/gocognit"
)

type functionMetric struct {
	Name       string `json:"name"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Lines      int    `json:"lines"`
	Cyclomatic int    `json:"cyclomatic"`
	Cognitive  int    `json:"cognitive"`
}

type fileMetric struct {
	Path         string `json:"path"`
	SHA256       string `json:"sha256"`
	SourceLines  int    `json:"source_lines"`
	CodeLines    int    `json:"code_lines"`
	CommentLines int    `json:"comment_lines"`
	BlankLines   int    `json:"blank_lines"`
}

type summary struct {
	Files               int     `json:"files"`
	Packages            int     `json:"packages"`
	Functions           int     `json:"functions"`
	SourceLines         int     `json:"source_lines"`
	CodeLines           int     `json:"code_lines"`
	CommentLines        int     `json:"comment_lines"`
	CommentDensity      float64 `json:"comment_density_percent"`
	BlankLines          int     `json:"blank_lines"`
	CyclomaticTotal     int     `json:"cyclomatic_total"`
	CyclomaticMax       int     `json:"cyclomatic_max"`
	CyclomaticMean      float64 `json:"cyclomatic_mean"`
	CyclomaticP90       int     `json:"cyclomatic_p90"`
	CognitiveMax        int     `json:"cognitive_max"`
	CognitiveMean       float64 `json:"cognitive_mean"`
	FunctionLinesMax    int     `json:"function_lines_max"`
	FunctionLinesMean   float64 `json:"function_lines_mean"`
	PackageFanOutMax    int     `json:"package_fan_out_max"`
	PackageFanOutMean   float64 `json:"package_fan_out_mean"`
	CyclomaticOver10    int     `json:"cyclomatic_over_10"`
	CognitiveOver15     int     `json:"cognitive_over_15"`
	FunctionLinesOver60 int     `json:"function_lines_over_60"`
}

type report struct {
	GoVersion     string            `json:"go_version"`
	Tools         map[string]string `json:"tools"`
	Scope         string            `json:"scope"`
	Summary       summary           `json:"summary"`
	PackageFanOut map[string]int    `json:"package_fan_out"`
	Files         []fileMetric      `json:"files"`
	Functions     []functionMetric  `json:"functions"`
}

func main() {
	root := flag.String("root", ".", "backend source directory")
	output := flag.String("out", "", "JSON destination (stdout when omitted)")
	flag.Parse()
	r, err := measure(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	data = append(data, '\n')
	if *output == "" {
		_, err = os.Stdout.Write(data)
	} else {
		err = os.WriteFile(*output, data, 0644)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func excluded(path string, d fs.DirEntry) bool {
	if d.IsDir() {
		switch d.Name() {
		case ".git", "vendor", "build", "bin", "node_modules", "tools":
			return true
		}
		return false
	}
	name := d.Name()
	return !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || strings.Contains(filepath.ToSlash(path), "cmd/metrics/")
}

func measure(root string) (report, error) {
	r := report{GoVersion: runtime.Version(), Tools: map[string]string{"gocyclo": "v0.6.0", "gocognit": "v1.2.0", "source": "Go parser/scanner"}, Scope: "Manual production Go files; excludes *_test.go, generated code, cmd/metrics, tools, vendor, build and bin", PackageFanOut: map[string]int{}, Files: []fileMetric{}, Functions: []functionMetric{}}
	imports := map[string]map[string]bool{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			if excluded(relative, d) {
				return filepath.SkipDir
			}
			return nil
		}
		if excluded(relative, d) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, data, parser.ParseComments)
		if err != nil {
			return err
		}
		if ast.IsGenerated(file) {
			return nil
		}
		metric := lineMetrics(filepath.ToSlash(relative), data)
		r.Files = append(r.Files, metric)
		pkg := filepath.ToSlash(filepath.Dir(relative))
		if imports[pkg] == nil {
			imports[pkg] = map[string]bool{}
		}
		for _, im := range file.Imports {
			value, err := strconv.Unquote(im.Path.Value)
			if err != nil {
				return err
			}
			imports[pkg][value] = true
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			r.Functions = append(r.Functions, measureFunction(pkg, metric.Path, fset, fn))
		}
		return nil
	})
	if err != nil {
		return r, err
	}
	for pkg, set := range imports {
		r.PackageFanOut[pkg] = len(set)
	}
	r.Summary = aggregate(r)
	sort.Slice(r.Functions, func(i, j int) bool {
		a, b := r.Functions[i], r.Functions[j]
		if a.Cyclomatic != b.Cyclomatic {
			return a.Cyclomatic > b.Cyclomatic
		}
		if a.File != b.File {
			return a.File < b.File
		}
		return a.Line < b.Line
	})
	return r, nil
}

func measureFunction(pkg, path string, fset *token.FileSet, fn *ast.FuncDecl) functionMetric {
	name := fn.Name.Name
	if fn.Recv != nil {
		var receiver bytes.Buffer
		_ = format.Node(&receiver, fset, fn.Recv.List[0].Type)
		name = "(" + receiver.String() + ")." + name
	}
	return functionMetric{Name: pkg + "." + name, File: path, Line: fset.Position(fn.Pos()).Line, Lines: fset.Position(fn.End()).Line - fset.Position(fn.Pos()).Line + 1, Cyclomatic: gocyclo.Complexity(fn), Cognitive: gocognit.Complexity(fn)}
}

func lineMetrics(path string, data []byte) fileMetric {
	sum := sha256.Sum256(data)
	lines := bytes.Split(bytes.TrimSuffix(data, []byte("\n")), []byte("\n"))
	m := fileMetric{Path: path, SHA256: hex.EncodeToString(sum[:]), SourceLines: len(lines)}
	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			m.BlankLines++
		}
	}
	fset := token.NewFileSet()
	f := fset.AddFile(path, -1, len(data))
	var scan scanner.Scanner
	scan.Init(f, data, nil, scanner.ScanComments)
	code, comments := map[int]bool{}, map[int]bool{}
	for {
		pos, tok, lit := scan.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.SEMICOLON && lit == "\n" {
			continue
		}
		dest := code
		if tok == token.COMMENT {
			dest = comments
		}
		start := fset.Position(pos).Line
		end := start + strings.Count(lit, "\n")
		for line := start; line <= end; line++ {
			dest[line] = true
		}
	}
	m.CodeLines = len(code)
	m.CommentLines = len(comments)
	return m
}

func aggregate(r report) summary {
	s := summary{Files: len(r.Files), Packages: len(r.PackageFanOut), Functions: len(r.Functions)}
	for _, f := range r.Files {
		s.SourceLines += f.SourceLines
		s.CodeLines += f.CodeLines
		s.CommentLines += f.CommentLines
		s.BlankLines += f.BlankLines
	}
	if s.SourceLines > 0 {
		s.CommentDensity = 100 * float64(s.CommentLines) / float64(s.SourceLines)
	}
	var cognitive, lines int
	complexities := []int{}
	for _, f := range r.Functions {
		s.CyclomaticTotal += f.Cyclomatic
		cognitive += f.Cognitive
		lines += f.Lines
		s.CyclomaticMax = max(s.CyclomaticMax, f.Cyclomatic)
		s.CognitiveMax = max(s.CognitiveMax, f.Cognitive)
		s.FunctionLinesMax = max(s.FunctionLinesMax, f.Lines)
		complexities = append(complexities, f.Cyclomatic)
		if f.Cyclomatic > 10 {
			s.CyclomaticOver10++
		}
		if f.Cognitive > 15 {
			s.CognitiveOver15++
		}
		if f.Lines > 60 {
			s.FunctionLinesOver60++
		}
	}
	if s.Functions > 0 {
		s.CyclomaticMean = float64(s.CyclomaticTotal) / float64(s.Functions)
		s.CognitiveMean = float64(cognitive) / float64(s.Functions)
		s.FunctionLinesMean = float64(lines) / float64(s.Functions)
		sort.Ints(complexities)
		s.CyclomaticP90 = complexities[(len(complexities)*9+9)/10-1]
	}
	var fanout int
	for _, count := range r.PackageFanOut {
		fanout += count
		s.PackageFanOutMax = max(s.PackageFanOutMax, count)
	}
	if s.Packages > 0 {
		s.PackageFanOutMean = float64(fanout) / float64(s.Packages)
	}
	return s
}
