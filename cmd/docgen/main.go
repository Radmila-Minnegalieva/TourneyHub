// Command docgen renders the small laboratory example using Go's parser and go/doc.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/doc"
	"go/doc/comment"
	"go/parser"
	"go/token"
	"html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf16"
)

func run() error {
	out := flag.String("out", "build/docs", "output directory")
	serve := flag.String("serve", "", "serve repository on this address after generation")
	flag.Parse()
	packages, err := parser.ParseDir(token.NewFileSet(), "internal/model", func(info os.FileInfo) bool { return !strings.HasSuffix(info.Name(), "_test.go") }, parser.ParseComments)
	if err != nil {
		return err
	}
	pkg := doc.New(packages["model"], "github.com/Radmila-Minnegalieva/TourneyHub/internal/model", doc.AllDecls|doc.PreserveAST)
	var page, plain bytes.Buffer
	page.WriteString("<!doctype html><html lang=\"en\"><meta charset=\"utf-8\"><title>TourneyHub — GoDoc example</title><style>body{max-width:960px;margin:40px auto;font:16px/1.6 system-ui;padding:0 24px}pre{white-space:pre-wrap;background:#f4f5f7;padding:16px}h2{border-bottom:1px solid #ddd}a{color:#185abc}</style><h1>TourneyHub — GoDoc example</h1>")
	var printer comment.Printer
	render := func(title, text string) {
		fmt.Fprintf(&page, "<h2>%s</h2>", html.EscapeString(title))
		page.Write(printer.HTML(pkg.Parser().Parse(text)))
		fmt.Fprintf(&plain, "%s\n%s\n\n", title, text)
	}
	render("Project / package model", pkg.Doc)
	page.WriteString("<p>Source file: <a href=\"../../internal/model/rating.go\">rating.go</a>. <a href=\"example.rtf\">Download RTF</a></p>")
	plain.WriteString("Source: internal/model/rating.go\n\n")
	filenames := make([]string, 0, len(packages["model"].Files))
	for filename := range packages["model"].Files {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)
	for _, filename := range filenames {
		file := packages["model"].Files[filename]
		if file.Doc != nil {
			render("File "+filepath.Base(filename), file.Doc.Text())
		}
	}
	for _, group := range pkg.Consts {
		render("Constants: "+strings.Join(group.Names, ", "), group.Doc)
	}
	for _, typ := range pkg.Types {
		render("Type "+typ.Name, typ.Doc)
		for _, function := range typ.Funcs {
			render("Function "+function.Name, function.Doc)
		}
		for _, method := range typ.Methods {
			render("Method "+typ.Name+"."+method.Name, method.Doc)
		}
	}
	// Source includes the field descriptions, full signatures, imports and implementation.
	source, err := os.ReadFile("internal/model/rating.go")
	if err != nil {
		return err
	}
	page.WriteString("<h2>Fields, signatures and source</h2><pre>" + html.EscapeString(string(source)) + "</pre></html>")
	plain.Write(source)
	if err := os.MkdirAll(*out, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(*out, "index.html"), page.Bytes(), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(*out, "example.txt"), plain.Bytes(), 0644); err != nil {
		return err
	}
	var rtf strings.Builder
	rtf.WriteString("{\\rtf1\\ansi\\deff0{\\fonttbl{\\f0 Arial;}}\\f0\\fs22\n")
	for _, r := range plain.String() {
		switch r {
		case '\\', '{', '}':
			rtf.WriteRune('\\')
			rtf.WriteRune(r)
		case '\n':
			rtf.WriteString("\\par\n")
		default:
			if r < 128 {
				rtf.WriteRune(r)
			} else {
				for _, u := range utf16.Encode([]rune{r}) {
					fmt.Fprintf(&rtf, "\\u%d?", int16(u))
				}
			}
		}
	}
	rtf.WriteString("}")
	if err := os.WriteFile(filepath.Join(*out, "example.rtf"), []byte(rtf.String()), 0644); err != nil {
		return err
	}
	fmt.Println("Documentation: " + filepath.Join(*out, "index.html"))
	if *serve != "" {
		fmt.Printf("Open http://%s/%s/index.html\n", *serve, filepath.ToSlash(*out))
		mux := http.NewServeMux()
		prefix := "/" + filepath.ToSlash(*out) + "/"
		mux.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(*out))))
		mux.HandleFunc("GET /internal/model/rating.go", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			if _, err := w.Write(source); err != nil {
				log.Printf("write source: %v", err)
			}
		})
		return http.ListenAndServe(*serve, mux)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
