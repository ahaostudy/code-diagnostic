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

// BigModel interface
type BigModel interface {
	// Chat Receive a query for large model calls and write the output results to the Result channel in real time
	Chat(messages []*Message) chan Result
}

type Result struct {
	Type    int
	Content string
}

const (
	TypeData = iota
	TypeDone
	TypeError
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
)

func Messages(messages ...*Message) []*Message {
	return messages
}

func UserMessage(msg string) *Message {
	return &Message{
		Role:    RoleUser,
		Content: msg,
	}
}

func SystemMessage(msg string) *Message {
	return &Message{
		Role:    RoleSystem,
		Content: msg,
	}
}

func AssistantMessage(msg string) *Message {
	return &Message{
		Role:    RoleAssistant,
		Content: msg,
	}
}
