/**
 * copyright ahaostudy
 *
 * licensed to the apache software foundation (asf) under one or more
 * contributor license agreements.  see the notice file distributed with
 * this work for additional information regarding copyright ownership.
 * the asf licenses this file to you under the apache license, version 2.0
 * (the "license"); you may not use this file except in compliance with
 * the license.  you may obtain a copy of the license at
 *
 *     http://www.apache.org/licenses/license-2.0
 *
 * unless required by applicable law or agreed to in writing, software
 * distributed under the license is distributed on an "as is" basis,
 * without warranties or conditions of any kind, either express or implied.
 * see the license for the specific language governing permissions and
 * limitations under the license.
 */

package parse

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ahaostudy/code-diagnostic/utils"
)

var root string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("get working path failed:", err.Error())
		return
	}
	root = wd
}

type Function struct {
	Name    string   `json:"name"`
	Params  []*Field `json:"params"`
	Results []*Field `json:"results"`
	File    string   `json:"file"`
	Type    string   `json:"type"`
	Source  string   `json:"source"`
	Line    int      `json:"line"`
}

func NewFunction(name string, params, results []*Field, file, source string) *Function {
	typ := "function"
	if strings.Contains(name, ".") {
		typ = "method"
	}
	return &Function{
		Name:    name,
		Params:  params,
		Results: results,
		File:    file,
		Type:    typ,
		Source:  source,
	}
}

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ReadFuncSource read parse source code
func ReadFuncSource(file, fun string, strict bool) (*Function, error) {
	for i := len(fun) - 1; i >= 0; i-- {
		if fun[i] == '/' {
			fun = fun[i+1:]
			break
		}
	}
	for i := 0; i < len(fun); i++ {
		if fun[i] == '.' {
			fun = fun[i+1:]
			break
		}
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		log.Fatalln("read parse source code failed:", err.Error())
		return nil, err
	}

	for _, decl := range node.Decls {
		var f *ast.FuncDecl
		var ok bool
		if f, ok = decl.(*ast.FuncDecl); !ok {
			continue
		}
		funcName := f.Name.Name
		if f.Recv != nil && len(f.Recv.List) > 0 {
			switch typ := f.Recv.List[0].Type.(type) {
			case *ast.Ident:
				funcName = typ.Name + "." + funcName
			case *ast.StarExpr:
				if ident, ok := typ.X.(*ast.Ident); ok {
					funcName = "(*" + ident.Name + ")." + funcName
				}
			}
		}
		if fun == funcName {
			start := fset.Position(f.Pos()).Offset
			end := fset.Position(f.End()).Offset
			// read source code
			// TODO: handle index out of range
			source := utils.ReadFile(file)

			var params []*Field
			for _, param := range f.Type.Params.List {
				for _, name := range param.Names {
					p := &Field{Name: name.Name}
					if decl, ok := name.Obj.Decl.(*ast.Field); ok {
						p.Type = GetTypeStr(decl.Type)
					}
					params = append(params, p)
				}
			}
			var results []*Field
			if f.Type.Results != nil {
				for _, result := range f.Type.Results.List {
					r := &Field{Type: GetTypeStr(result.Type)}
					results = append(results, r)
				}
			}
			return NewFunction(fun, params, results, file, string(source[start:end])), nil
		}
	}
	if strings.Contains(fun, ".") && !strict {
		return ReadFuncSource(file, strings.TrimRight(fun, filepath.Ext(fun)), strict)
	}

	return nil, fmt.Errorf("the source code of parse %s cannot be found in %s", fun, file)
}

func GetTypeStr(t ast.Expr) string {
	switch typ := t.(type) {
	case *ast.Ident:
		return typ.Name
	case *ast.StarExpr:
		if ident, ok := typ.X.(*ast.Ident); ok {
			return "*" + ident.Name
		}
	case *ast.ArrayType:
		if ident, ok := typ.Elt.(*ast.Ident); ok {
			return "[]" + ident.Name
		}
	case *ast.MapType:
		if key, ok := typ.Key.(*ast.Ident); ok {
			if value, ok := typ.Value.(*ast.Ident); ok {
				return "map[" + key.Name + "]" + value.Name
			}
		}
	default:
		log.Fatalln("unsupported type:", typ)
	}
	return ""
}

func GetFuncList(frames *runtime.Frames) (funs []*Function) {
	set := map[string]struct{}{}
	for {
		frame, more := frames.Next()
		if strings.HasPrefix(frame.File, root) {
			if _, ok := set[frame.Function]; !ok {
				fun, err := ReadFuncSource(frame.File, frame.Function, true)
				if err != nil {
					log.Println(err)
					continue
				}
				if fun != nil {
					fun.Line = frame.Line
					funs = append(funs, fun)
					set[frame.Function] = struct{}{}
				}
			}
		}
		if !more {
			break
		}
	}
	return
}

func GetFuncListWithStackTraces(stackTraces []*StackTrace) (funs []*Function) {
	for _, trace := range stackTraces {
		name := filepath.Base(trace.Func)
		if name == "panic" {
			funs = funs[:0]
			continue
		}
		fun, err := ReadFuncSource(trace.File, name, false)
		if err != nil {
			fun = NewFunction(trace.Func, nil, nil, trace.File, "")
		}
		if !strings.HasSuffix(name, fun.Name) {
			fun.Params = nil
			fun.Results = nil
		}
		fun.Name = name
		fun.Line = trace.Line
		funs = append(funs, fun)
	}
	return
}

func MakeFileFuncSet(funs []*Function) map[string][]*Function {
	fset := make(map[string][]*Function)
	for _, f := range funs {
		fset[f.File] = append(fset[f.File], f)
	}
	return fset
}

func BuildFuncListDescription(funs []*Function) string {
	fset := MakeFileFuncSet(funs)

	var desc string
	for file, fs := range fset {
		desc += BuildFileFunctionsDescription(file, fs)
	}
	return desc
}

func BuildFileFunctionsDescription(file string, funs []*Function) string {
	desc := file + ":\n```go\n"
	for _, f := range funs {
		desc += f.Source + "\n"
	}
	desc += "```\n"
	return desc
}
