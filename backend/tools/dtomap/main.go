// Command dtomap генерирует только транспортные преобразования между независимыми моделями.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/constant"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

const modulePath = "github.com/Radmila-Minnegalieva/TourneyHub/backend"
const runtimeTypes = "github.com/oapi-codegen/runtime/types"

type source struct {
	types   map[string]ast.Expr
	imports map[string]string
	enums   map[string]map[string]bool
}

type generator struct {
	model   source
	api     source
	names   []string
	imports map[string]string
}

func main() {
	modelFile := flag.String("model", "internal/model/entities.go", "independent domain declarations")
	apiFile := flag.String("api", "pkg/api/v1/api.gen.go", "OpenAPI declarations")
	out := flag.String("out", "internal/api/v1/mapping.gen.go", "mapping output; input files are never modified")
	flag.Parse()
	if err := run(*modelFile, *apiFile, *out); err != nil {
		fmt.Fprintln(os.Stderr, "dtomap:", err)
		os.Exit(1)
	}
}

func run(modelFile, apiFile, output string) error {
	domain, err := readSource(modelFile)
	if err != nil {
		return err
	}
	api, err := readSource(apiFile)
	if err != nil {
		return err
	}
	data, err := generate(domain, api)
	if err != nil {
		return err
	}
	return os.WriteFile(output, data, 0644)
}

func readSource(filename string) (source, error) {
	file, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
	if err != nil {
		return source{}, err
	}
	result := source{types: map[string]ast.Expr{}, imports: map[string]string{}, enums: map[string]map[string]bool{}}
	for _, item := range file.Imports {
		importPath, err := strconv.Unquote(item.Path.Value)
		if err != nil {
			return result, err
		}
		name := path.Base(importPath)
		if item.Name != nil {
			name = item.Name.Name
		}
		result.imports[name] = importPath
	}
	for _, decl := range file.Decls {
		group, ok := decl.(*ast.GenDecl)
		if !ok || group.Tok != token.TYPE {
			continue
		}
		for _, item := range group.Specs {
			typ := item.(*ast.TypeSpec)
			result.types[typ.Name.Name] = typ.Type
		}
	}
	if err := result.readEnums(file); err != nil {
		return result, fmt.Errorf("%s: %w", filename, err)
	}
	return result, nil
}

// readEnums сравнивает значения, а не имена констант: переименование не меняет контракт.
func (s *source) readEnums(file *ast.File) error {
	constants := map[string]constant.Value{}
	for _, decl := range file.Decls {
		group, ok := decl.(*ast.GenDecl)
		if !ok || group.Tok != token.CONST {
			continue
		}
		var previous []ast.Expr
		var previousType ast.Expr
		for i, item := range group.Specs {
			value := item.(*ast.ValueSpec)
			expressions, typ := value.Values, value.Type
			if len(expressions) == 0 {
				expressions, typ = previous, previousType
			}
			previous, previousType = expressions, typ
			for j, name := range value.Names {
				if j >= len(expressions) {
					return fmt.Errorf("constant %s has no expression", name.Name)
				}
				computed, err := evaluate(expressions[j], constants, i)
				if err != nil {
					return fmt.Errorf("constant %s: %w", name.Name, err)
				}
				constants[name.Name] = computed
				named, ok := typ.(*ast.Ident)
				if !ok {
					continue
				}
				if _, exists := s.types[named.Name]; !exists {
					continue
				}
				if s.enums[named.Name] == nil {
					s.enums[named.Name] = map[string]bool{}
				}
				s.enums[named.Name][computed.Kind().String()+":"+computed.ExactString()] = true
			}
		}
	}
	return nil
}

func evaluate(expr ast.Expr, constants map[string]constant.Value, iotaValue int) (constant.Value, error) {
	switch value := expr.(type) {
	case *ast.BasicLit:
		return constant.MakeFromLiteral(value.Value, value.Kind, 0), nil
	case *ast.Ident:
		switch value.Name {
		case "true":
			return constant.MakeBool(true), nil
		case "false":
			return constant.MakeBool(false), nil
		case "iota":
			return constant.MakeInt64(int64(iotaValue)), nil
		}
		if result, ok := constants[value.Name]; ok {
			return result, nil
		}
	case *ast.ParenExpr:
		return evaluate(value.X, constants, iotaValue)
	case *ast.UnaryExpr:
		operand, err := evaluate(value.X, constants, iotaValue)
		if err != nil {
			return nil, err
		}
		return constant.UnaryOp(value.Op, operand, 0), nil
	}
	return nil, fmt.Errorf("unsupported constant expression %s", render(expr))
}

func generate(model, api source) ([]byte, error) {
	g := generator{model: model, api: api, imports: map[string]string{modulePath + "/internal/model": "model", modulePath + "/pkg/api/v1": "contract"}}
	for name := range model.types {
		if _, ok := api.types[name]; ok {
			g.names = append(g.names, name)
		}
	}
	sort.Strings(g.names)
	if len(g.names) == 0 {
		return nil, fmt.Errorf("no shared types in domain and API declarations")
	}
	var functions bytes.Buffer
	for _, name := range g.names {
		if err := g.validate(name, model.types[name], api.types[name]); err != nil {
			return nil, err
		}
		if !equalSet(model.enums[name], api.enums[name]) {
			return nil, fmt.Errorf("%s: enum values differ between domain and API", name)
		}
		for _, toModel := range []bool{true, false} {
			if err := g.function(&functions, name, toModel); err != nil {
				return nil, err
			}
		}
	}
	var output bytes.Buffer
	output.WriteString("// Code generated by go run ./tools/dtomap. DO NOT EDIT.\n// Независимые доменные модели не изменяются; различия полей и типов проверены генератором.\npackage api\n\nimport (\n")
	var paths []string
	for importPath := range g.imports {
		paths = append(paths, importPath)
	}
	sort.Strings(paths)
	for _, importPath := range paths {
		fmt.Fprintf(&output, "%s %q\n", g.imports[importPath], importPath)
	}
	output.WriteString(")\n\n")
	output.Write(functions.Bytes())
	return format.Source(output.Bytes())
}

func equalSet(left, right map[string]bool) bool {
	if len(left) != len(right) {
		return false
	}
	for value := range left {
		if !right[value] {
			return false
		}
	}
	return true
}

func fields(typ *ast.StructType) (map[string]ast.Expr, []string, error) {
	result := map[string]ast.Expr{}
	var names []string
	for _, field := range typ.Fields.List {
		if len(field.Names) == 0 {
			return nil, nil, fmt.Errorf("embedded fields require an explicit mapping")
		}
		for _, name := range field.Names {
			result[name.Name] = field.Type
			names = append(names, name.Name)
		}
	}
	return result, names, nil
}

func (g *generator) validate(location string, domain, api ast.Expr) error {
	if canonical(domain, g.model) == canonical(api, g.api) {
		return nil
	}
	// UUID в oapi-codegen является alias google/uuid; Email намеренно остаётся строкой в домене.
	if isSelector(api, g.api, runtimeTypes, "UUID") && isSelector(domain, g.model, "github.com/google/uuid", "UUID") {
		return nil
	}
	if isSelector(api, g.api, runtimeTypes, "Email") && render(domain) == "string" {
		return nil
	}
	switch left := domain.(type) {
	case *ast.StructType:
		right, ok := api.(*ast.StructType)
		if !ok {
			break
		}
		leftFields, names, err := fields(left)
		if err != nil {
			return fmt.Errorf("%s: %w", location, err)
		}
		rightFields, _, err := fields(right)
		if err != nil {
			return fmt.Errorf("%s: %w", location, err)
		}
		for _, name := range names {
			other, exists := rightFields[name]
			if !exists {
				return fmt.Errorf("%s.%s: domain field is absent from API", location, name)
			}
			if err := g.validate(location+"."+name, leftFields[name], other); err != nil {
				return err
			}
		}
		var extra []string
		for name := range rightFields {
			if _, exists := leftFields[name]; !exists {
				extra = append(extra, name)
			}
		}
		sort.Strings(extra)
		if len(extra) > 0 {
			return fmt.Errorf("%s.%s: API field is absent from domain (including optional fields)", location, extra[0])
		}
		return nil
	case *ast.StarExpr:
		if right, ok := api.(*ast.StarExpr); ok {
			return g.validate(location+"*", left.X, right.X)
		}
	case *ast.ArrayType:
		if right, ok := api.(*ast.ArrayType); ok && render(left.Len) == render(right.Len) {
			return g.validate(location+"[]", left.Elt, right.Elt)
		}
	case *ast.MapType:
		if right, ok := api.(*ast.MapType); ok {
			if err := g.validate(location+" key", left.Key, right.Key); err != nil {
				return err
			}
			return g.validate(location+" value", left.Value, right.Value)
		}
	}
	return fmt.Errorf("%s: type shape differs: domain %s, API %s", location, render(domain), render(api))
}

func canonical(expr ast.Expr, file source) string {
	switch typ := expr.(type) {
	case *ast.Ident:
		if typ.Name == "any" {
			return "interface{}"
		}
		return typ.Name
	case *ast.SelectorExpr:
		if qualifier, ok := typ.X.(*ast.Ident); ok {
			return file.imports[qualifier.Name] + "." + typ.Sel.Name
		}
	case *ast.StarExpr:
		return "*" + canonical(typ.X, file)
	case *ast.ArrayType:
		return "[" + render(typ.Len) + "]" + canonical(typ.Elt, file)
	case *ast.MapType:
		return "map[" + canonical(typ.Key, file) + "]" + canonical(typ.Value, file)
	}
	return render(expr)
}

func isSelector(expr ast.Expr, file source, importPath, name string) bool {
	return canonical(expr, file) == importPath+"."+name
}

func (g *generator) function(output *bytes.Buffer, name string, toModel bool) error {
	source, target, prefix := "contract", "model", "toModel"
	if !toModel {
		source, target, prefix = "model", "contract", "fromModel"
	}
	fmt.Fprintf(output, "func %s%s(v %s.%s) %s.%s {\n", prefix, name, source, name, target, name)
	if structure, ok := g.model.types[name].(*ast.StructType); ok {
		modelFields, names, _ := fields(structure)
		apiFields, _, _ := fields(g.api.types[name].(*ast.StructType))
		fmt.Fprintf(output, "return %s.%s{\n", target, name)
		for _, field := range names {
			converted, err := g.convert(modelFields[field], apiFields[field], "v."+field, toModel)
			if err != nil {
				return fmt.Errorf("%s.%s: %w", name, field, err)
			}
			fmt.Fprintf(output, "%s: %s,\n", field, converted)
		}
		output.WriteString("}\n")
	} else {
		fmt.Fprintf(output, "return %s.%s(v)\n", target, name)
	}
	output.WriteString("}\n\n")
	return nil
}

func (g *generator) convert(domain, api ast.Expr, value string, toModel bool) (string, error) {
	prefix := "toModel"
	if !toModel {
		prefix = "fromModel"
	}
	if isSelector(api, g.api, runtimeTypes, "Email") {
		if toModel {
			return "string(" + value + ")", nil
		}
		return g.typeName(api, false) + "(" + value + ")", nil
	}
	switch typ := domain.(type) {
	case *ast.Ident:
		if _, ok := g.model.types[typ.Name]; ok {
			if _, shared := g.api.types[typ.Name]; !shared {
				return "", fmt.Errorf("domain type %s has no API mapping", typ.Name)
			}
			return prefix + typ.Name + "(" + value + ")", nil
		}
	case *ast.StarExpr:
		right := api.(*ast.StarExpr)
		inner, err := g.convert(typ.X, right.X, "x", toModel)
		if err != nil {
			return "", err
		}
		if inner == "x" {
			return value, nil
		}
		g.imports[modulePath+"/internal/lib/ptr"] = "ptr"
		source, target := g.typeName(right.X, false), g.typeName(typ.X, true)
		if !toModel {
			source, target = target, source
		}
		return fmt.Sprintf("ptr.Map(%s,func(x %s)%s{return %s})", value, source, target, inner), nil
	case *ast.ArrayType:
		right := api.(*ast.ArrayType)
		inner, err := g.convert(typ.Elt, right.Elt, "x", toModel)
		if err != nil {
			return "", err
		}
		if inner == "x" {
			return value, nil
		}
		if typ.Len != nil {
			return "", fmt.Errorf("fixed arrays with mapped elements require an explicit mapping")
		}
		g.imports[modulePath+"/internal/lib/slice"] = "slice"
		source, target := g.typeName(right.Elt, false), g.typeName(typ.Elt, true)
		if !toModel {
			source, target = target, source
		}
		return fmt.Sprintf("slice.Map(%s,func(x %s)%s{return %s})", value, source, target, inner), nil
	case *ast.MapType:
		right := api.(*ast.MapType)
		key, err := g.convert(typ.Key, right.Key, "x", toModel)
		if err != nil {
			return "", err
		}
		element, err := g.convert(typ.Value, right.Value, "x", toModel)
		if err != nil {
			return "", err
		}
		if key != "x" || element != "x" {
			return "", fmt.Errorf("maps with mapped keys or values require an explicit mapping")
		}
	}
	return value, nil
}

func (g *generator) typeName(expr ast.Expr, domain bool) string {
	file, prefix := g.api, "contract"
	if domain {
		file, prefix = g.model, "model"
	}
	switch typ := expr.(type) {
	case *ast.Ident:
		if _, ok := file.types[typ.Name]; ok {
			return prefix + "." + typ.Name
		}
		return typ.Name
	case *ast.SelectorExpr:
		qualifier := typ.X.(*ast.Ident).Name
		importPath := file.imports[qualifier]
		alias := path.Base(importPath)
		g.imports[importPath] = alias
		return alias + "." + typ.Sel.Name
	case *ast.StarExpr:
		return "*" + g.typeName(typ.X, domain)
	case *ast.ArrayType:
		return "[" + render(typ.Len) + "]" + g.typeName(typ.Elt, domain)
	case *ast.MapType:
		return "map[" + g.typeName(typ.Key, domain) + "]" + g.typeName(typ.Value, domain)
	}
	return render(expr)
}

func render(node any) string {
	if node == nil {
		return ""
	}
	var buffer bytes.Buffer
	if err := format.Node(&buffer, token.NewFileSet(), node); err != nil {
		return fmt.Sprint(node)
	}
	return strings.TrimSpace(buffer.String())
}
