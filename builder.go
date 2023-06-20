// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"reflect"

	"github.com/yxlib/yx"
)

type builder struct {
	logger *yx.Logger
}

var Builder = &builder{
	logger: yx.NewLogger("server.Builder"),
}

// build server.
// @param srv, dest server.
// @param cfg, the server config.
func (b *builder) Build(srv Server, cfg *Config) {
	if cfg.IsAutoModCmd {
		b.genAutoModCmd(srv, cfg)
	}

	// cfg.MapName2Service = make(map[string]*ServiceConf)
	cfg.MapProcName2ProtoNo = make(map[string]uint32)
	for _, serviceCfg := range cfg.Services {
		b.BuildService(srv, cfg, serviceCfg)
	}

	srv.SetEnumerable(cfg.IsEnumerable, cfg.EnumProtoNo, cfg.MapProcName2ProtoNo)
}

func (b *builder) BuildService(srv Server, cfg *Config, serviceCfg *ServiceConf) {
	s, ok := ServiceBinder.GetService(serviceCfg.Service)
	if !ok {
		b.logger.W("Not support service ", serviceCfg.Service)
		return
	}

	// cfg.MapName2Service[serviceCfg.Name] = serviceCfg
	srv.AddService(s, serviceCfg.Mod)
	b.buildProcessor(s, cfg, serviceCfg)
}

func (b *builder) genAutoModCmd(srv Server, cfg *Config) {
	mod := uint32(1)
	for _, service := range cfg.Services {
		service.Mod = mod
		mod++

		cmd := uint32(1)
		for _, proc := range service.Processors {
			proc.ProtoNo = GetProtoNo(mod, cmd)
			cmd++
		}
	}
}

func (b *builder) buildProcessor(s Service, cfg *Config, serviceCfg *ServiceConf) {
	v := reflect.ValueOf(s)

	// serviceCfg.MapName2Proc = make(map[string]*ProcConf)
	for _, procCfg := range serviceCfg.Processors {
		// proto
		// if cfg.Req != "" && cfg.Resp != "" {
		err := ProtoBinder.BindProto(procCfg.ProtoNo, procCfg.Req, procCfg.Resp)
		if err != nil {
			b.logger.W("not support processor ", procCfg.Handler)
			continue
		}
		// }
		// serviceCfg.MapName2Proc[procCfg.Name] = procCfg
		if len(serviceCfg.Name) > 0 || len(procCfg.Name) > 0 {
			procFullName := fmt.Sprintf("%s.%s", serviceCfg.Name, procCfg.Name)
			cfg.MapProcName2ProtoNo[procFullName] = procCfg.ProtoNo
		}

		// processor
		m := v.MethodByName(procCfg.Handler)
		cmd := GetCmd(procCfg.ProtoNo)
		err = s.AddReflectProcessor(m, cmd)
		if err != nil {
			b.logger.E("AddReflectProcessor err: ", err)
			b.logger.W("not support processor ", procCfg.Handler)
			continue
		}
	}
}
