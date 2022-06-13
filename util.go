// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

const CMD_PER_MOD uint16 = 100

func GetMod(protoNo uint16) uint16 {
	return uint16(protoNo / CMD_PER_MOD)
}

func GetCmd(protoNo uint16) uint16 {
	return uint16(protoNo % CMD_PER_MOD)
}

func GetProtoNo(mod uint16, cmd uint16) uint16 {
	return uint16(mod*CMD_PER_MOD + cmd)
}
