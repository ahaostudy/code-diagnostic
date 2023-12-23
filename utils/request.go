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

package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Request struct {
	url    string
	header map[string]string
	data   map[string]interface{}
}

func NewRequest(url string) *Request {
	return &Request{url: url}
}

func (r *Request) SetHeader(key, value string) {
	if r.header == nil {
		r.header = make(map[string]string)
	}
	r.header[key] = value
}

func (r *Request) SetData(data map[string]interface{}) {
	r.data = data
}

func (r *Request) POST() (*http.Response, error) {
	data, err := json.Marshal(r.data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", r.url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	for k, v := range r.header {
		req.Header.Set(k, v)
	}

	return new(http.Client).Do(req)
}
