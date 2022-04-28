// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

type ProcConf struct {
	Cmd     uint16 `json:"cmd"`
	Req     string `json:"req"`
	Resp    string `json:"resp"`
	Handler string `json:"handler"`
}

type ServiceConf struct {
	Service    string      `json:"service"`
	Mod        uint16      `json:"mod"`
	Processors []*ProcConf `json:"processors"`
}

type Config struct {
	MaxReqNum  uint16         `json:"max_req_num"`
	MaxTaskNum uint16         `json:"max_task_num"`
	Services   []*ServiceConf `json:"services"`
}

var CfgInst *Config = &Config{}
