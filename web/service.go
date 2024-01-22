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

package web

import (
	"github.com/ahaostudy/code-diagnostic/bigmodel"
	"github.com/ahaostudy/code-diagnostic/parse"
)

func ChatService(messages []*bigmodel.Message) chan bigmodel.Result {
	var msg string
	msg += "The following error occurred in the current program: \n```\n" + config.Panic + "\n```\n\n"
	msg += "Here is its call stack: \n```\n" + config.Stack + "```\n\n"
	msg += "The source code list is as follows:\n" + parse.BuildFuncListDescription(config.LocalFunctions) + "\n"
	if config.UseChinese {
		msg += "Please reply in Chinese to help analyze the cause of the error and solve it!"
	} else {
		msg += "Please help analyze the cause of the error and solve it!"
	}
	messages = append(bigmodel.Messages(bigmodel.SystemMessage(msg)), messages...)
	return config.BigModel.Chat(messages)
}
