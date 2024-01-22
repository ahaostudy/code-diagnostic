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
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/ahaostudy/code-diagnostic/bigmodel"
	"github.com/ahaostudy/code-diagnostic/parse"
)

type Config struct {
	Panic          string
	Stack          string
	LocalFunctions []*parse.Function
	Functions      []*parse.Function

	BigModel   bigmodel.BigModel
	UseChinese bool
}

var (
	config *Config
	root   string
)

func InitConfig(conf *Config) {
	config = conf
}

func init() {
	_, file, _, _ := runtime.Caller(0)
	root = filepath.Dir(file)
}

func Run(port int) error {
	initRouter()

	logStr := fmt.Sprintf("Diagnostic service started:\n\nhttp://localhost:%d/", port)
	ip, ok := getLocalIP()
	if ok {
		logStr += fmt.Sprintf("\nhttp://%s:%d/", ip, port)
	}
	logStr += "\n\nYou can enter the diagnostic service to view detailed error analysis."
	log.Println(logStr)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func getLocalIP() (string, bool) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", false
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp != 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				return "", false
			}
			for _, addr := range addrs {
				ipNet, ok := addr.(*net.IPNet)
				if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
					return ipNet.IP.String(), true
				}
			}
		}
	}
	return "", false
}
