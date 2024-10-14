package main

import (
	"bytes"
	_ "embed"
	"go/format"
	"io/fs"
	"os"
	"strings"
	"text/template"
)

//go:embed view.tgo
var viewTemplate string

type viewData struct {
	Views [][]string
}

// This is used to generate code for the ecs library
func main() {
	data := viewData{
		Views: [][]string{
			[]string{"A"},
			[]string{"A", "B"},
			[]string{"A", "B", "C"},
			[]string{"A", "B", "C", "D"},
			[]string{"A", "B", "C", "D", "E"},
			[]string{"A", "B", "C", "D", "E", "F"},
			[]string{"A", "B", "C", "D", "E", "F", "G"},
			[]string{"A", "B", "C", "D", "E", "F", "G", "H"},
			[]string{"A", "B", "C", "D", "E", "F", "G", "H", "I"},
			[]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"},
			[]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K"},
			[]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L"},
		},
	}
	funcs := template.FuncMap{
		"join": strings.Join,
		"lower": func(val string) string {
			return strings.ToLower(val)
		},
		"nils": func(n int) string {
			val := make([]string, 0)
			for i := 0; i < n; i++ {
				val = append(val, "nil")
			}
			return strings.Join(val, ", ")
		},
		"retlist": func(val []string) string {
			ret := make([]string, len(val))
			for i := range val {
				ret[i] = "ret" + val[i]
			}
			return strings.Join(ret, ", ")
		},
		"lambdaArgs": func(val []string) string {
			ret := make([]string, len(val))
			for i := range val {
				ret[i] = strings.ToLower(val[i]) + " *" + val[i]
			}
			return strings.Join(ret, ", ")
		},
		"sliceLambdaArgs": func(val []string) string {
			ret := make([]string, len(val))
			for i := range val {
				ret[i] = strings.ToLower(val[i]) + " []" + val[i]
			}
			return strings.Join(ret, ", ")
		},
		"parallelLambdaStructArgs": func(val []string) string {
			ret := make([]string, len(val))
			for i := range val {
				ret[i] = strings.ToLower(val[i]) + " []" + val[i]
			}
			return strings.Join(ret, "; ")
		},
		"parallelLambdaArgsFromStruct": func(val []string) string {
			ret := make([]string, len(val))
			for i := range val {
				ret[i] = "param" + val[i]
			}
			return strings.Join(ret, ", ")
		},
	}

	t := template.Must(template.New("ViewTemplate").Funcs(funcs).Parse(viewTemplate))

	buf := bytes.NewBuffer([]byte{})

	t.Execute(buf, data)

	filename := "view_gen.go"

	// Attempt to write the file as formatted, falling back to writing it normally
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		err = os.WriteFile(filename, buf.Bytes(), fs.ModePerm)
		if err != nil {
			panic(err)
		}
		panic(err)
	}

	err = os.WriteFile(filename, formatted, fs.ModePerm)
	if err != nil {
		panic(err)
	}
}
