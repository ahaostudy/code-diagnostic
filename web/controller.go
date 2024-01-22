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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/ahaostudy/code-diagnostic/bigmodel"
	"github.com/ahaostudy/code-diagnostic/parse"
)

func HTMLHandlerFunc(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(root, "html", path))
	}
}

func GetPanic(w http.ResponseWriter, r *http.Request) {
	Success(w, JSON{
		"panic":     config.Panic,
		"stack":     config.Stack,
		"functions": config.Functions,
	})
}

// GetFuncSource
// TODO: unused
func GetFuncSource(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	fun := query.Get("func")
	file := query.Get("file")
	function, err := parse.ReadFuncSource(file, fun, false)
	if err != nil {
		Error(w, err.Error())
		return
	}
	Success(w, JSON{
		"function": function,
	})
}

type ChatRequest struct {
	Messages []*bigmodel.Message `json:"messages"`
}

func Chat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		Event(w, "error", err.Error())
		return
	}
	data := new(ChatRequest)
	err = json.Unmarshal(body, &data)
	if err != nil {
		Event(w, "error", err.Error())
		return
	}

	answer := ChatService(data.Messages)
	for finish := false; !finish; {
		ans := <-answer
		switch ans.Type {
		case bigmodel.TypeData:
			Event(w, "message", ans.Content)
		case bigmodel.TypeDone:
			Event(w, "done", "")
		case bigmodel.TypeError:
			Event(w, "error", "chatgpt response error: "+ans.Content)
		default:
			Event(w, "error", "chatgpt response unknown type: "+fmt.Sprint(ans.Type))
		}
		finish = ans.Type != bigmodel.TypeData
	}
	close(answer)
}
