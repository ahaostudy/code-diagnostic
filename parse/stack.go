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
	"regexp"
	"strconv"
)

type StackTrace struct {
	Func string `json:"func"`
	File string `json:"file"`
	Line int    `json:"line"`
}

func StackTraces(stack []byte) []*StackTrace {
	regex := regexp.MustCompile(`(?m)(.*?)\(([^.]*?)\)\n\s+(.*?):(\d+)`)
	matches := regex.FindAllSubmatch(stack, -1)
	var stackTraces []*StackTrace
	for _, match := range matches {
		line, _ := strconv.Atoi(string(match[4]))
		stackTraces = append(stackTraces, &StackTrace{
			Func: string(match[1]),
			File: string(match[3]),
			Line: line,
		})
	}
	return stackTraces
}
