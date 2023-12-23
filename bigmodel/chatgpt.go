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

package bigmodel

import (
	"encoding/json"
	"io"
	"log"
	"strings"

	"github.com/ahaostudy/code-diagnostic/utils"
)

// ChatGPT big model
type ChatGPT struct {
	url string

	model   string
	baseURL string
	apiKey  string
}

func NewChatGPT(apiKey string, opts ...Option) BigModel {
	gpt := &ChatGPT{
		model:   "gpt-3.5-turbo",
		baseURL: "https://api.openai.com",
		apiKey:  apiKey,
	}
	for _, opt := range opts {
		opt(gpt)
	}
	gpt.url = strings.TrimSuffix(gpt.baseURL, "/") + "/v1/chat/completions"
	return gpt
}

type Option func(*ChatGPT)

// WithSpecifyModel specify ChatGPT model
func WithSpecifyModel(model string) Option {
	return func(gpt *ChatGPT) {
		gpt.model = model
	}
}

// WithSpecifyBaseURL specify OPENAI API base url
func WithSpecifyBaseURL(url string) Option {
	return func(gpt *ChatGPT) {
		gpt.baseURL = url
	}
}

type chunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func (gpt *ChatGPT) Chat(query string) chan Result {
	out := make(chan Result)

	go func() {
		req := utils.NewRequest(gpt.url)
		req.SetData(map[string]interface{}{
			"model":  gpt.model,
			"stream": true,
			"messages": []map[string]string{
				{
					"role":    "user",
					"content": query,
				},
			},
		})
		req.SetHeader("Authorization", "Bearer "+gpt.apiKey)
		req.SetHeader("Content-Type", "application/json")

		// response
		resp, err := req.POST()
		if err != nil {
			log.Fatalln("openai request failed:", err)
			return
		}
		defer resp.Body.Close()

		// TODO: response status exception

		// read response stream data
		buf := make([]byte, 1024)
		var chunks, tmp string
		for {
			n, err := resp.Body.Read(buf)
			if err == io.EOF {
				out <- Result{Type: TypeDone}
				return
			}
			if err != nil {
				out <- Result{Type: TypeError, Content: err.Error()}
				return
			}

			chunks, tmp = tmp+string(buf[:n]), ""
			for _, chk := range strings.Split(chunks, "\n\n") {
				chk = strings.TrimPrefix(chk, "data: ")

				// done
				if chk == "[DONE]" {
					out <- Result{Type: TypeDone}
					return
				}
				data := new(chunk)
				err := json.Unmarshal([]byte(chk), data)
				if err != nil {
					tmp += chk
					continue
				}
				if len(data.Choices) == 0 {
					out <- Result{Type: TypeError, Content: "response data error"}
					return
				}
				out <- Result{Type: TypeData, Content: data.Choices[0].Delta.Content}
			}
		}
	}()

	return out
}
