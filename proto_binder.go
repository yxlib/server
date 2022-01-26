// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"errors"
	"reflect"
)

var (
	ErrProtoBindRegProtoIsNil  = errors.New("proto is nil")
	ErrProtoBindUnknownPattern = errors.New("unknown pattern")
	ErrProtoBindUnknownOpr     = errors.New("unknown operation")
	ErrProtoBindProtoExist     = errors.New("proto has exist")
	ErrProtoBindProtoNotExist  = errors.New("proto not exist")
)

const CMD_PER_MOD = 100

type protoBinder struct {
	mapName2ProtoType map[string]reflect.Type
	mapProtoNo2Req    map[uint16]reflect.Type
	mapProtoNo2Resp   map[uint16]reflect.Type
}

var ProtoBinder = &protoBinder{
	mapName2ProtoType: make(map[string]reflect.Type),
	mapProtoNo2Req:    make(map[uint16]reflect.Type),
	mapProtoNo2Resp:   make(map[uint16]reflect.Type),
}

// Register proto type.
// @param proto, the proto.
func (b *protoBinder) RegisterProto(proto interface{}) error {
	if proto == nil {
		return ErrProtoBindRegProtoIsNil
	}

	t := reflect.TypeOf(proto)
	t = t.Elem()
	path := t.PkgPath()
	name := path + "." + t.Name()
	b.mapName2ProtoType[name] = t

	return nil
}

// Get the proto type by type name.
// @param name, the proto type name.
// @return reflect.Type, the reflect type of the proto.
// @return bool, true mean success, false mean failed.
func (b *protoBinder) GetProtoType(name string) (reflect.Type, bool) {
	t, ok := b.mapName2ProtoType[name]
	return t, ok
}

// Bind protos.
// @param mod, the module of the service.
// @param cmd, the command of the service.
// @param reqProtoName, the request proto name.
// @param respProtoName, the response proto name.
// @return error, error.
func (b *protoBinder) BindProto(mod uint16, cmd uint16, reqProtoName string, respProtoName string) error {
	protoNo := b.getProtoNo(mod, cmd)

	_, ok := b.mapProtoNo2Req[protoNo]
	if ok {
		return ErrProtoBindProtoExist
	}

	_, ok = b.mapProtoNo2Resp[protoNo]
	if ok {
		return ErrProtoBindProtoExist
	}

	t, ok := b.GetProtoType(reqProtoName)
	if !ok {
		return ErrProtoBindProtoNotExist
	}
	b.mapProtoNo2Req[protoNo] = t

	t, ok = b.GetProtoType(respProtoName)
	if !ok {
		return ErrProtoBindProtoNotExist
	}
	b.mapProtoNo2Resp[protoNo] = t
	return nil
}

// Get the request reflect type.
// @param mod, the module of the service.
// @param cmd, the command of the service.
// @return reflect.Type, reflect type of the request.
// @return error, error.
func (b *protoBinder) GetRequestType(mod uint16, cmd uint16) (reflect.Type, error) {
	protoNo := b.getProtoNo(mod, cmd)
	req, ok := b.mapProtoNo2Req[protoNo]
	if !ok {
		return nil, ErrProtoBindProtoNotExist
	}

	return req, nil
}

// Get the response reflect type.
// @param mod, the module of the service.
// @param cmd, the command of the service.
// @return reflect.Type, reflect type of the response.
// @return error, error.
func (b *protoBinder) GetResponseType(mod uint16, cmd uint16) (reflect.Type, error) {
	protoNo := b.getProtoNo(mod, cmd)
	resp, ok := b.mapProtoNo2Resp[protoNo]
	if !ok {
		return nil, ErrProtoBindProtoNotExist
	}

	return resp, nil
}

func (b *protoBinder) getProtoNo(mod uint16, cmd uint16) uint16 {
	return mod*CMD_PER_MOD + cmd
}
