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
	mapProtoNo2ReqName  map[uint32]string
	mapProtoNo2RespName map[uint32]string
	factory             *yx.ObjectFactory
	ec                  *yx.ErrCatcher
}

var ProtoBinder = &protoBinder{
	mapProtoNo2ReqName:  make(map[uint32]string),
	mapProtoNo2RespName: make(map[uint32]string),
	factory:             yx.NewObjectFactory(),
	ec:                  yx.NewErrCatcher("protoBinder"),
}

// Register proto type.
// @param proto, the proto.
func (b *protoBinder) RegisterProto(proto interface{}) error {
	_, err := b.factory.RegisterObject(proto, nil, MAX_REUSE_COUNT)
	return b.ec.Throw("RegisterProto", err)
}

// Get the proto type by type name.
// @param name, the proto type name.
// @return reflect.Type, the reflect type of the proto.
// @return bool, true mean success, false mean failed.
func (b *protoBinder) GetProtoType(name string) (reflect.Type, bool) {
	return b.factory.GetReflectType(name)
}

// Bind protos.
// @param protoNo, the proto number.
// @param reqProtoName, the request proto name.
// @param respProtoName, the response proto name.
// @return error, error.
func (b *protoBinder) BindProto(protoNo uint32, reqProtoName string, respProtoName string) error {
	_, ok := b.mapProtoNo2ReqName[protoNo]
	if ok {
		return b.ec.Throw("BindProto", ErrProtoBindProtoExist)
	}

	_, ok = b.mapProtoNo2RespName[protoNo]
	if ok {
		return b.ec.Throw("BindProto", ErrProtoBindProtoExist)
	}

	_, ok = b.factory.GetReflectType(reqProtoName)
	if ok {
		b.mapProtoNo2ReqName[protoNo] = reqProtoName
	}

	_, ok = b.factory.GetReflectType(respProtoName)
	if ok {
		b.mapProtoNo2RespName[protoNo] = respProtoName
	}

	return nil
}

// Get the request reflect type.
// @param protoNo, the proto number.
// @return reflect.Type, reflect type of the request.
// @return error, error.
func (b *protoBinder) GetRequestType(protoNo uint32) (reflect.Type, error) {
	name, ok := b.mapProtoNo2ReqName[protoNo]
	if !ok {
		return nil, b.ec.Throw("GetRequestType", ErrProtoBindProtoNotExist)
	}

	refType, err := b.getReflectTypeByName(name)
	return refType, b.ec.Throw("GetRequestType", err)
}

// Get the response reflect type.
// @param protoNo, the proto number.
// @return reflect.Type, reflect type of the response.
// @return error, error.
func (b *protoBinder) GetResponseType(protoNo uint32) (reflect.Type, error) {
	name, ok := b.mapProtoNo2RespName[protoNo]
	if !ok {
		return nil, b.ec.Throw("GetResponseType", ErrProtoBindProtoNotExist)
	}

	refType, err := b.getReflectTypeByName(name)
	return refType, b.ec.Throw("GetResponseType", err)
}

// Get an request object.
// @param protoNo, the proto number.
// @return interface{}, the request object.
// @return error, error.
func (b *protoBinder) GetRequest(protoNo uint32) (interface{}, error) {
	name, ok := b.mapProtoNo2ReqName[protoNo]
	if !ok {
		return nil, b.ec.Throw("GetRequest", ErrProtoBindProtoNotExist)
	}

	obj, err := b.factory.CreateObject(name)
	return obj, b.ec.Throw("GetRequest", err)
}

// Reuse an request object.
// @param v, the reuse request.
// @param protoNo, the proto number.
// @return error, error.
func (b *protoBinder) ReuseRequest(v interface{}, protoNo uint32) error {
	if v == nil {
		return b.ec.Throw("ReuseRequest", ErrProtoBindReuseIsNil)
	}

	name, ok := b.mapProtoNo2ReqName[protoNo]
	if !ok {
		return b.ec.Throw("ReuseRequest", ErrProtoBindProtoNotExist)
	}

	err := b.factory.ReuseObject(v, name)
	return b.ec.Throw("ReuseRequest", err)
}

// Get an response object.
// @param protoNo, the proto number.
// @return interface{}, the response object.
// @return error, error.
func (b *protoBinder) GetResponse(protoNo uint32) (interface{}, error) {
	name, ok := b.mapProtoNo2RespName[protoNo]
	if !ok {
		return nil, b.ec.Throw("GetResponse", ErrProtoBindProtoNotExist)
	}

	obj, err := b.factory.CreateObject(name)
	return obj, b.ec.Throw("GetResponse", err)
}

// Reuse an response object.
// @param v, the reuse response.
// @param protoNo, the proto number.
// @return error, error.
func (b *protoBinder) ReuseResponse(v interface{}, protoNo uint32) error {
	if v == nil {
		return b.ec.Throw("ReuseResponse", ErrProtoBindReuseIsNil)
	}

	name, ok := b.mapProtoNo2RespName[protoNo]
	if !ok {
		return b.ec.Throw("ReuseResponse", ErrProtoBindProtoNotExist)
	}

	err := b.factory.ReuseObject(v, name)
	return b.ec.Throw("ReuseResponse", err)
}

func (b *protoBinder) getReflectTypeByName(name string) (reflect.Type, error) {
	objType, ok := b.factory.GetReflectType(name)
	if !ok {
		return nil, b.ec.Throw("getReflectTypeByName", ErrProtoBindProtoNotExist)
	}

	return objType, nil
}
