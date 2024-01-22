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
	"log"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/ahaostudy/code-diagnostic/bigmodel"
	"github.com/ahaostudy/code-diagnostic/parse"
	"github.com/ahaostudy/code-diagnostic/web"
)

const (
	defaultMaxStack = 1024
	defaultWebPort  = 5789
)

type Diag struct {
	BigModel bigmodel.BigModel

	useChinese bool
	useWeb     bool
	webPort    int
}

func NewDiag(bm bigmodel.BigModel, opts ...Option) *Diag {
	d := &Diag{
		BigModel: bm,
	}
	for _, opt := range opts {
		opt(d)
	}
	if d.webPort == 0 {
		d.webPort = defaultWebPort
	}
	return d
}

func (diag *Diag) Diagnostic() {
	if r := recover(); r != nil {
		pnc := fmt.Sprintf("%s", r)
		stack := string(debug.Stack())
		frames := getCallersFrames(defaultMaxStack)
		diag.diagnostic(pnc, stack, frames)
	}
}

func (diag *Diag) BreakPoint(pnc string) {
	stack := string(debug.Stack())
	frames := getCallersFrames(defaultMaxStack)
	diag.diagnostic(pnc, stack, frames)
}

func (diag *Diag) diagnostic(pnc, stack string, frames *runtime.Frames) {
	log.Printf("diagnostic detected:\n\n\t%v\n\n\t%v",
		pnc,
		strings.ReplaceAll(stack, "\n", "\n\t"),
	)
	localFuns := parse.GetFuncList(frames)
	if !diag.useWeb {
		diag.analyze(pnc, stack, localFuns)
	} else {
		stackTraces := parse.StackTraces([]byte(stack))
		funs := parse.GetFuncListWithStackTraces(stackTraces)
		web.InitConfig(&web.Config{
			Panic:          pnc,
			Stack:          stack,
			LocalFunctions: localFuns,
			Functions:      funs,
			BigModel:       diag.BigModel,
			UseChinese:     diag.useChinese,
		})
		if err := web.Run(diag.webPort); err != nil {
			log.Fatalln("web run error:", err)
		}
	}
}

func (diag *Diag) analyze(pnc, stack string, funs []*parse.Function) {
	var msg string
	msg += "The following error occurred in the current program: \n```\n" + pnc + "\n```\n\n"
	msg += "Here is its call stack: \n```\n" + stack + "```\n\n"
	msg += "The source code list is as follows:\n" + parse.BuildFuncListDescription(funs) + "\n"
	if diag.useChinese {
		msg += "Please reply in Chinese to help analyze the cause of the error and solve it!"
	} else {
		msg += "Please help analyze the cause of the error and solve it!"
	}

	answer := diag.BigModel.Chat(bigmodel.Messages(bigmodel.UserMessage(msg)))
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
