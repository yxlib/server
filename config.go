// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

type ProcConf struct {
	Name    string `json:"name"`
	Cmd     uint16 `json:"cmd"`
	Req     string `json:"req"`
	Resp    string `json:"resp"`
	Handler string `json:"handler"`
}

type ServiceConf struct {
	Name       string      `json:"name"`
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

func (c *Config) GetProcName2Proto() map[string]uint16 {
	procName2Proto := make(map[string]uint16)
	for _, service := range c.Services {
		for _, proc := range service.Processors {
			protoNo := GetProtoNo(service.Mod, proc.Cmd)
			procName2Proto[proc.Name] = protoNo
		}
	}

	return procName2Proto
}
