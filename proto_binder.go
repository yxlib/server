// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"errors"
	"reflect"

	"github.com/yxlib/yx"
)

var (
	ErrProtoBindProtoExist    = errors.New("proto has exist")
	ErrProtoBindProtoNotExist = errors.New("proto not exist")
	ErrProtoBindReuseIsNil    = errors.New("reuse object is nil")
)

const MAX_REUSE_COUNT = 100

type protoBinder struct {
	mapProtoNo2ReqName  map[uint16]string
	mapProtoNo2RespName map[uint16]string
	factory             *yx.ObjectFactory
}

var ProtoBinder = &protoBinder{
	mapProtoNo2ReqName:  make(map[uint16]string),
	mapProtoNo2RespName: make(map[uint16]string),
	factory:             yx.NewObjectFactory(),
}

// Register proto type.
// @param proto, the proto.
func (b *protoBinder) RegisterProto(proto interface{}) error {
	_, err := b.factory.RegisterObject(proto, nil, MAX_REUSE_COUNT)
	return err
}

// Get the proto type by type name.
// @param name, the proto type name.
// @return reflect.Type, the reflect type of the proto.
// @return bool, true mean success, false mean failed.
func (b *protoBinder) GetProtoType(name string) (reflect.Type, bool) {
	return b.factory.GetReflectType(name)
}

// Bind protos.
// @param mod, the module of the service.
// @param cmd, the command of the service.
// @param reqProtoName, the request proto name.
// @param respProtoName, the response proto name.
// @return error, error.
func (b *protoBinder) BindProto(mod uint16, cmd uint16, reqProtoName string, respProtoName string) error {
	protoNo := GetProtoNo(mod, cmd)

	_, ok := b.mapProtoNo2ReqName[protoNo]
	if ok {
		return ErrProtoBindProtoExist
	}

	_, ok = b.factory.GetReflectType(reqProtoName)
	if !ok {
		return ErrProtoBindProtoNotExist
	}

	_, ok = b.mapProtoNo2RespName[protoNo]
	if ok {
		return ErrProtoBindProtoExist
	}

	_, ok = b.factory.GetReflectType(respProtoName)
	if !ok {
		return ErrProtoBindProtoNotExist
	}

	b.mapProtoNo2ReqName[protoNo] = reqProtoName
	b.mapProtoNo2RespName[protoNo] = respProtoName
	return nil
}

// Get the request reflect type.
// @param mod, the module of the service.
// @param cmd, the command of the service.
// @return reflect.Type, reflect type of the request.
// @return error, error.
func (b *protoBinder) GetRequestType(mod uint16, cmd uint16) (reflect.Type, error) {
	protoNo := GetProtoNo(mod, cmd)
	name, ok := b.mapProtoNo2ReqName[protoNo]
	if !ok {
		return nil, ErrProtoBindProtoNotExist
	}

	return b.getReflectTypeByName(name)
}

// Get the response reflect type.
// @param mod, the module of the service.
// @param cmd, the command of the service.
// @return reflect.Type, reflect type of the response.
// @return error, error.
func (b *protoBinder) GetResponseType(mod uint16, cmd uint16) (reflect.Type, error) {
	protoNo := GetProtoNo(mod, cmd)
	name, ok := b.mapProtoNo2RespName[protoNo]
	if !ok {
		return nil, ErrProtoBindProtoNotExist
	}

	return b.getReflectTypeByName(name)
}

// Get an request object.
// @param mod, the module of the service.
// @param cmd, the command of the service.
// @return interface{}, the request object.
// @return error, error.
func (b *protoBinder) GetRequest(mod uint16, cmd uint16) (interface{}, error) {
	protoNo := GetProtoNo(mod, cmd)
	name, ok := b.mapProtoNo2ReqName[protoNo]
	if !ok {
		return nil, ErrProtoBindProtoNotExist
	}

	return b.factory.CreateObject(name)
}

// Reuse an request object.
// @param v, the reuse request.
// @param mod, the module of the service.
// @param cmd, the command of the service.
// @return error, error.
func (b *protoBinder) ReuseRequest(v interface{}, mod uint16, cmd uint16) error {
	if v == nil {
		return ErrProtoBindReuseIsNil
	}

	protoNo := GetProtoNo(mod, cmd)
	name, ok := b.mapProtoNo2ReqName[protoNo]
	if !ok {
		return ErrProtoBindProtoNotExist
	}

	return b.factory.ReuseObject(v, name)
}

// Get an response object.
// @param mod, the module of the service.
// @param cmd, the command of the service.
// @return interface{}, the response object.
// @return error, error.
func (b *protoBinder) GetResponse(mod uint16, cmd uint16) (interface{}, error) {
	protoNo := GetProtoNo(mod, cmd)
	name, ok := b.mapProtoNo2RespName[protoNo]
	if !ok {
		return nil, ErrProtoBindProtoNotExist
	}

	return b.factory.CreateObject(name)
}

// Reuse an response object.
// @param v, the reuse response.
// @param mod, the module of the service.
// @param cmd, the command of the service.
// @return error, error.
func (b *protoBinder) ReuseResponse(v interface{}, mod uint16, cmd uint16) error {
	if v == nil {
		return ErrProtoBindReuseIsNil
	}

	protoNo := GetProtoNo(mod, cmd)
	name, ok := b.mapProtoNo2RespName[protoNo]
	if !ok {
		return ErrProtoBindProtoNotExist
	}

	return b.factory.ReuseObject(v, name)
}

func (b *protoBinder) getReflectTypeByName(name string) (reflect.Type, error) {
	objType, ok := b.factory.GetReflectType(name)
	if !ok {
		return nil, ErrProtoBindProtoNotExist
	}

	return objType, nil
}
