package gen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"golang.org/x/tools/imports"
)

// File represents a generation file.
type File struct {
	Name string

	IncludeFiles IncludeFiles

	Functions []interface{}
	Enums     []*Enum
	Structs   []*Struct
}

// NewFile creates a new blank file.
func NewFile(name string) *File {
	return &File{
		Name:         name,
		IncludeFiles: NewIncludeFiles(),
	}
}

var templateGenerateFile = template.Must(template.New("go-clang-generate-file").Parse(`package clang

{{range $h, $dunno := $.IncludeFiles}}// #include "{{$h}}"
{{end}}// #include "go-clang.h"
import "C"

{{range $i, $f := $.Functions}}
{{$f}}
{{end}}

{{range $i, $e := $.Enums}}
{{$e.Comment}}
type {{$e.Name}} {{$e.UnderlyingType}}

const (
{{range $i, $ei := .Items}}	{{if $ei.Comment}}{{$ei.Comment}}
	{{end}}{{$ei.Name}}{{if eq $i 0}} {{$e.Name}}{{end}} = C.{{$ei.CName}}
{{end}}
)

{{range $i, $m := $e.Methods}}
{{$m}}
{{end}}
{{end}}

{{range $i, $s := $.Structs}}
{{$s.Comment}}
type {{$s.Name}} struct {
	c {{if $s.IsPointerComposition}}*{{end}}C.{{if not $s.CNameIsTypeDef}}struct_{{end}}{{$s.CName}}
}
{{range $i, $m := $s.Methods}}
{{$m}}
{{end}}
{{end}}
`))

type GenerateWaiter struct {
	sync.WaitGroup
	Errors []error
	Mu     sync.Mutex
}

var GenWaiter = &GenerateWaiter{}

func (g *GenerateWaiter) AddError(err error) {
	g.Mu.Lock()
	fmt.Println(err)
	g.Errors = append(g.Errors, err)
	g.Mu.Unlock()
}

func (g *GenerateWaiter) GetErrors() error {
	if len(g.Errors) == 0 {
		return nil
	}

	return fmt.Errorf("errors: %v", g.Errors)
}

// Generate generates file.
func (f *File) Generate() error {
	for _, e := range f.Enums {
		f.IncludeFiles.unifyIncludeFiles(e.IncludeFiles)

		for _, fu := range e.Methods {
			switch fu := fu.(type) {
			case *Function:
				f.IncludeFiles.unifyIncludeFiles(fu.IncludeFiles)
			}
		}
	}

	for _, s := range f.Structs {
		f.IncludeFiles.unifyIncludeFiles(s.IncludeFiles)

		for _, fu := range s.Methods {
			switch fu := fu.(type) {
			case *Function:
				f.IncludeFiles.unifyIncludeFiles(fu.IncludeFiles)
			}
		}
	}

	for _, fn := range f.Functions {
		switch fu := fn.(type) {
		case *Function:
			f.IncludeFiles.unifyIncludeFiles(fu.IncludeFiles)
		}
	}

	var b bytes.Buffer
	if err := templateGenerateFile.Execute(&b, f); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	filename := filepath.Join(cwd, "clang", f.Name+"_gen.go")

	bo := b.Bytes()
	bo = bytes.ReplaceAll(bo, []byte(`#include "./clang/`), []byte(`#include "./`))
	GenWaiter.Add(1)
	go func() {
		defer GenWaiter.Done()
		fmt.Println("try import ", filename)
		out, err := imports.Process(filename, bo, nil)
		if err != nil {
			fmt.Println("failed to import ", filename, err)
			// Write the file anyway so we can look at the problem
			if err := os.WriteFile(filename, bo, 0600); err != nil {
				GenWaiter.AddError(err)
				return
			}

			GenWaiter.AddError(err)
			return
		}

		fmt.Printf("file %s is successfully imported\n", filename)

		if err := os.WriteFile(filename, out, 0600); err != nil {
			GenWaiter.AddError(err)
		}
	}()

	return nil
}
