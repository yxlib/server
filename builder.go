// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
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
func (b *builder) Build(srv *Server, cfg *Config) {
	for _, servCfg := range cfg.Services {
		b.parsePatternCfg(srv, servCfg)
	}
}

func (b *builder) parsePatternCfg(srv *Server, servCfg *ServiceConf) {
	s, ok := ServiceBinder.GetService(servCfg.Service)
	if !ok {
		b.logger.W("Not support service ", servCfg.Service)
		return
	}

	srv.AddService(s, servCfg.Mod)
	b.parseProcCfg(s, servCfg)
}

func (b *builder) parseProcCfg(s Service, servCfg *ServiceConf) {
	v := reflect.ValueOf(s)

	for _, cfg := range servCfg.Processors {
		// proto
		if cfg.Req != "" && cfg.Resp != "" {
			err := ProtoBinder.BindProto(servCfg.Mod, cfg.Cmd, cfg.Req, cfg.Resp)
			if err != nil {
				b.logger.W("not support processor ", cfg.Handler)
				continue
			}
		}

		// processor
		m := v.MethodByName(cfg.Handler)
		err := s.AddReflectProcessor(m, cfg.Cmd)
		if err != nil {
			b.logger.E("AddReflectProcessor err: ", err)
			b.logger.W("not support processor ", cfg.Handler)
			continue
		}
	}
}
