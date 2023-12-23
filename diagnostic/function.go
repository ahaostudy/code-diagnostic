/**
 * Copyright ahaostudy
 *
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package diagnostic

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ahaostudy/code-dignostic/utils"
)

type Function struct {
	Name   string
	File   string
	Source string
}

func NewFunction(name, file, source string) *Function {
	return &Function{
		Name:   name,
		File:   file,
		Source: source,
	}
}

// ReadFuncSource read function source code
func ReadFuncSource(frame *runtime.Frame) *Function {
	file := frame.File
	fun := filepath.Ext(frame.Function)[1:]

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		log.Fatalln("read function source code failed:", err.Error())
		return nil
	}

	for _, decl := range node.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok && fun == f.Name.Name {
			start := fset.Position(f.Pos()).Offset
			end := fset.Position(f.End()).Offset
			// read source code
			// TODO: handle index out of range
			source := utils.ReadFile(file)
			return NewFunction(fun, file, string(source[start:end]))
		}
	}

	log.Fatalln("function source code not found")
	return nil
}

func GetFuncList(frames *runtime.Frames) (funs []*Function) {
	for {
		frame, more := frames.Next()
		if strings.HasPrefix(frame.File, root) {
			fun := ReadFuncSource(&frame)
			funs = append(funs, fun)
		}
		if !more {
			break
		}
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
