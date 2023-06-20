// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

const CMD_PER_MOD uint32 = 100

func GetMod(protoNo uint32) uint32 {
	return protoNo / CMD_PER_MOD
}

func GetCmd(protoNo uint32) uint32 {
	return protoNo % CMD_PER_MOD
}

func GetProtoNo(mod uint32, cmd uint32) uint32 {
	return mod*CMD_PER_MOD + cmd
}
