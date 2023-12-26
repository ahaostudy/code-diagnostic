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
	"fmt"
	"github.com/ahaostudy/code-diagnostic/bigmodel"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
)

const (
	defaultMaxStack = 1024
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

type Diag struct {
	BigModel bigmodel.BigModel

	useChinese bool
}

func NewDiag(bm bigmodel.BigModel, opts ...Option) *Diag {
	d := &Diag{
		BigModel: bm,
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

func (diag *Diag) Diagnostic() {
	if r := recover(); r != nil {
		pnc := fmt.Sprintf("%s", r)
		stack := string(debug.Stack())
		log.Printf("diagnostic detected:\n\n\t%v\n\n\t%v",
			pnc,
			strings.ReplaceAll(stack, "\n", "\n\t"),
		)
		frames := getCallersFrames(defaultMaxStack)
		funs := GetFuncList(frames)
		diag.analyze(pnc, stack, funs)
	}
}

func (diag *Diag) BreakPoint(pnc string) {
	stack := string(debug.Stack())
	log.Printf("diagnostic detected:\n\n\t%v\n\n\t%v",
		pnc,
		strings.ReplaceAll(stack, "\n", "\n\t"),
	)
	frames := getCallersFrames(defaultMaxStack)
	funs := GetFuncList(frames)
	diag.analyze(pnc, stack, funs)
}

func (diag *Diag) analyze(pnc, stack string, funs []*Function) {
	var prompt string
	prompt += "The following error occurred in the current program: \n```\n" + pnc + "\n```\n\n"
	prompt += "Here is its call stack: \n```\n" + stack + "```\n\n"
	prompt += "The source code list is as follows:\n" + buildFuncListDescription(funs) + "\n"
	if diag.useChinese {
		prompt += "Please reply in Chinese to help analyze the cause of the error and solve it!"
	} else {
		prompt += "Please help analyze the cause of the error and solve it!"
	}

	answer := diag.BigModel.Chat(prompt)
	for finish := false; !finish; {
		ans := <-answer
		switch ans.Type {
		case bigmodel.TypeData:
			print(ans.Content)
		case bigmodel.TypeDone:
			finish = true
		case bigmodel.TypeError:
			log.Fatalln("chatgpt response error:", ans.Content)
		default:
			log.Fatalln("chatgpt response unknown type:", ans.Type)
		}
	}
	close(answer)
	println()
}

func getCallersFrames(max int) *runtime.Frames {
	pc := make([]uintptr, max)
	n := runtime.Callers(1, pc)
	pc = pc[:n]
	return runtime.CallersFrames(pc)
}

func buildFuncListDescription(funs []*Function) string {
	fset := MakeFileFuncSet(funs)

	var desc string
	for file, fs := range fset {
		desc += buildFileFunctionsDescription(file, fs)
	}
	return desc
}

func buildFileFunctionsDescription(file string, funs []*Function) string {
	desc := file + ":\n```go\n"
	for _, f := range funs {
		desc += f.Source + "\n"
	}
	desc += "```\n"
	return desc
}
