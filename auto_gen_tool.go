// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"os"
	"strings"

	"github.com/yxlib/yx"
)

func GenServiceRegisterFile(cfgPath string, regFilePath string, regPackName string) {
	yx.LoadJsonConf(CfgInst, cfgPath, nil)
	GenRegisterFileByCfg(CfgInst, regFilePath, regPackName)
}

// Generate the service register file.
// @param cfgPath, the config path.
// @param regFilePath, the output register file.
// @param regPackName, the package name of the file.
func GenRegisterFileByCfg(srvCfg *Config, regFilePath string, regPackName string) {
	f, err := os.OpenFile(regFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}

	defer f.Close()

	writePackage(f, regPackName)
	writeImport(srvCfg.Services, regPackName, f)

	f.WriteString("// Auto generate by tool.\n")
	f.WriteString("func RegisterServices() {\n")

	for _, servCfg := range srvCfg.Services {
		f.WriteString("    //===============================\n")
		f.WriteString("    //        " + servCfg.Service + "\n")
		f.WriteString("    //===============================\n")

		servStr := servCfg.Service
		idx := strings.LastIndex(servStr, "/")
		if idx >= 0 {
			servStr = servStr[idx+1:]
		}
		idx = strings.Index(servStr, ".")
		packName := servStr[:idx]
		servStr = servStr[idx+1:]
		f.WriteString("    server.ServiceBinder.BindService(" + packName + ".New" + servStr + "())\n")
		for _, cfg := range servCfg.Processors {
			if cfg.Req == "" || cfg.Resp == "" {
				continue
			}

			f.WriteString("    // " + cfg.Handler + "\n")

			reqStr := cfg.Req
			idx := strings.LastIndex(reqStr, "/")
			if idx >= 0 {
				reqStr = reqStr[idx+1:]
			}
			f.WriteString("    server.ProtoBinder.RegisterProto(&" + reqStr + "{})\n")

			respStr := cfg.Resp
			idx = strings.LastIndex(respStr, "/")
			if idx >= 0 {
				respStr = respStr[idx+1:]
			}
			f.WriteString("    server.ProtoBinder.RegisterProto(&" + respStr + "{})\n")
		}

		f.WriteString("\n")
	}

	f.WriteString("}")

}

func writePackage(f *os.File, regPackName string) {
	f.WriteString("// This File auto generate by tool.\n")
	f.WriteString("// Please do not modify.\n")
	f.WriteString("// See server.GenRegisterFileByCfg().\n\n")
	f.WriteString("package " + regPackName + "\n\n")
}

func writeImport(servConfs []*ServiceConf, regPackName string, f *os.File) {
	f.WriteString("import (\n")

	packSet := yx.NewSet(yx.SET_TYPE_OBJ)
	packSet.Add("github.com/yxlib/server")

	for _, servCfg := range servConfs {
		fullPackName := yx.GetFullPackageName(servCfg.Service)
		if fullPackName != "" {
			packSet.Add(fullPackName)
		}

		for _, cfg := range servCfg.Processors {
			addProtoPackage(cfg.Req, regPackName, packSet)
			addProtoPackage(cfg.Resp, regPackName, packSet)
		}
	}

	elements := packSet.GetElements()
	for _, packName := range elements {
		f.WriteString("    \"" + packName.(string) + "\"\n")
	}

	f.WriteString(")\n\n")
}

func addProtoPackage(protoCfg string, regPackName string, packSet *yx.Set) {
	if protoCfg == "" {
		return
	}

	fullPackName := yx.GetFullPackageName(protoCfg)
	filePackName := yx.GetFilePackageName(fullPackName)
	if filePackName != regPackName {
		packSet.Add(fullPackName)
	}
}
