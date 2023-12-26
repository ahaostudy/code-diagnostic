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

package main

import (
	"github.com/ahaostudy/code-diagnostic/bigmodel"
	"github.com/ahaostudy/code-diagnostic/diagnostic"
	"github.com/ahaostudy/code-diagnostic/example/math"
	"os"

	"github.com/joho/godotenv"
)

var (
	baseURL string
	apiKey  string
)

func init() {
	_ = godotenv.Load("dev.env", ".env")

	// set the base_url and api_key of ChatGPT
	baseURL = os.Getenv("BASE_URL")
	apiKey = os.Getenv("API_KEY")
}

func main() {
	// initialize a diagnostic tool
	diag := diagnostic.NewDiag(
		// use the ChatGPT model
		bigmodel.NewChatGPT(apiKey, bigmodel.WithSpecifyBaseURL(baseURL)),
		// use chinese
		diagnostic.WithUseChinese(),
	)

	// automatically capture and analyze program exceptions
	defer diag.Diagnostic()

	a, b := 10, 0
	//math.Div(a, b)

	_, err := math.DivError(a, b)
	if err != nil {
		// custom breakpoint analysis
		diag.BreakPoint(err.Error())
	}
}
