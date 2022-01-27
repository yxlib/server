// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"errors"
	"reflect"
	"sync"

	"github.com/yxlib/yx"
)

var (
	ErrProtoBindRegProtoIsNil  = errors.New("proto is nil")
	ErrProtoBindUnknownPattern = errors.New("unknown pattern")
	ErrProtoBindUnknownOpr     = errors.New("unknown operation")
	ErrProtoBindProtoExist     = errors.New("proto has exist")
	ErrProtoBindProtoNotExist  = errors.New("proto not exist")
	ErrProtoBindNotTheSameType = errors.New("not the same type")
	ErrProtoBindReuseIsNil     = errors.New("reuse object is nil")
)

const (
	CMD_PER_MOD      = 100
	INIT_REUSE_COUNT = 10
	MAX_REUSE_COUNT  = 100
)

type protoBinder struct {
	mapName2ProtoType   map[string]reflect.Type
	mapProtoNo2ReqType  map[uint16]reflect.Type
	mapProtoNo2RespType map[uint16]reflect.Type

	mapProtoNo2ReqPool map[uint16]*yx.LinkedQueue
	lckReqPool         *sync.Mutex

	mapProtoNo2RespPool map[uint16]*yx.LinkedQueue
	lckRespPool         *sync.Mutex
}

var ProtoBinder = &protoBinder{
	mapName2ProtoType:   make(map[string]reflect.Type),
	mapProtoNo2ReqType:  make(map[uint16]reflect.Type),
	mapProtoNo2RespType: make(map[uint16]reflect.Type),

	mapProtoNo2ReqPool:  make(map[uint16]*yx.LinkedQueue),
	lckReqPool:          &sync.Mutex{},
	mapProtoNo2RespPool: make(map[uint16]*yx.LinkedQueue),
	lckRespPool:         &sync.Mutex{},
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

	_, ok := b.mapProtoNo2ReqType[protoNo]
	if ok {
		return ErrProtoBindProtoExist
	}

	_, ok = b.mapProtoNo2RespType[protoNo]
	if ok {
		return ErrProtoBindProtoExist
	}

	reqType, ok := b.GetProtoType(reqProtoName)
	if !ok {
		return ErrProtoBindProtoNotExist
	}
	b.mapProtoNo2ReqType[protoNo] = reqType

	respType, ok := b.GetProtoType(respProtoName)
	if !ok {
		return ErrProtoBindProtoNotExist
	}
	b.mapProtoNo2RespType[protoNo] = respType

	b.initReuseProto(reqType, respType, protoNo)
	return nil
}

// Get the request reflect type.
// @param mod, the module of the service.
// @param cmd, the command of the service.
// @return reflect.Type, reflect type of the request.
// @return error, error.
func (b *protoBinder) GetRequestType(mod uint16, cmd uint16) (reflect.Type, error) {
	protoNo := b.getProtoNo(mod, cmd)
	reqType, ok := b.mapProtoNo2ReqType[protoNo]
	if !ok {
		return nil, ErrProtoBindProtoNotExist
	}

	return reqType, nil
}

// Get the response reflect type.
// @param mod, the module of the service.
// @param cmd, the command of the service.
// @return reflect.Type, reflect type of the response.
// @return error, error.
func (b *protoBinder) GetResponseType(mod uint16, cmd uint16) (reflect.Type, error) {
	protoNo := b.getProtoNo(mod, cmd)
	resp, ok := b.mapProtoNo2RespType[protoNo]
	if !ok {
		return nil, ErrProtoBindProtoNotExist
	}

	return resp, nil
}

func (b *protoBinder) GetRequest(mod uint16, cmd uint16) (interface{}, error) {
	protoNo := b.getProtoNo(mod, cmd)
	req, ok := b.popRequstFromPool(protoNo)
	if !ok {
		reqType, ok := b.mapProtoNo2ReqType[protoNo]
		if !ok {
			return nil, ErrProtoBindProtoNotExist
		}

		v := reflect.New(reqType)
		req = v.Interface()
	}

	return req, nil
}

func (b *protoBinder) ReuseRequest(v interface{}, mod uint16, cmd uint16) error {
	if v == nil {
		return ErrProtoBindReuseIsNil
	}

	protoNo := b.getProtoNo(mod, cmd)
	reqType, ok := b.mapProtoNo2ReqType[protoNo]
	if !ok {
		return ErrProtoBindProtoNotExist
	}

	t := reflect.TypeOf(v)
	t = t.Elem()
	if reqType != t {
		return ErrProtoBindNotTheSameType
	}

	b.pushRequest(v, protoNo)
	return nil
}

func (b *protoBinder) GetResponse(mod uint16, cmd uint16) (interface{}, error) {
	protoNo := b.getProtoNo(mod, cmd)
	resp, ok := b.popResponseFromPool(protoNo)
	if !ok {
		respType, ok := b.mapProtoNo2RespType[protoNo]
		if !ok {
			return nil, ErrProtoBindProtoNotExist
		}

		v := reflect.New(respType)
		resp = v.Interface()
	}

	return resp, nil
}

func (b *protoBinder) ReuseResponse(v interface{}, mod uint16, cmd uint16) error {
	if v == nil {
		return ErrProtoBindReuseIsNil
	}

	protoNo := b.getProtoNo(mod, cmd)
	respType, ok := b.mapProtoNo2RespType[protoNo]
	if !ok {
		return ErrProtoBindProtoNotExist
	}

	t := reflect.TypeOf(v)
	t = t.Elem()
	if respType != t {
		return ErrProtoBindNotTheSameType
	}

	b.pushResponse(v, protoNo)
	return nil
}

func (b *protoBinder) getProtoNo(mod uint16, cmd uint16) uint16 {
	return mod*CMD_PER_MOD + cmd
}

func (b *protoBinder) initReuseProto(reqType reflect.Type, respType reflect.Type, protoNo uint16) {
	for i := 0; i < INIT_REUSE_COUNT; i++ {
		v := reflect.New(reqType)
		b.pushRequest(v.Interface(), protoNo)

		v = reflect.New(respType)
		b.pushResponse(v.Interface(), protoNo)
	}
}

func (b *protoBinder) popRequstFromPool(protoNo uint16) (interface{}, bool) {
	b.lckReqPool.Lock()
	defer b.lckReqPool.Unlock()

	queue, ok := b.mapProtoNo2ReqPool[protoNo]
	if !ok {
		return nil, false
	}

	return b.popFromPool(queue)
}

func (b *protoBinder) pushRequest(v interface{}, protoNo uint16) {
	b.lckReqPool.Lock()
	defer b.lckReqPool.Unlock()

	queue, ok := b.mapProtoNo2ReqPool[protoNo]
	if !ok {
		queue = yx.NewLinkedQueue()
		b.mapProtoNo2ReqPool[protoNo] = queue
	}

	b.pushToPool(queue, v)
}

func (b *protoBinder) popResponseFromPool(protoNo uint16) (interface{}, bool) {
	b.lckRespPool.Lock()
	defer b.lckRespPool.Unlock()

	queue, ok := b.mapProtoNo2RespPool[protoNo]
	if !ok {
		return nil, false
	}

	return b.popFromPool(queue)
}

func (b *protoBinder) pushResponse(v interface{}, protoNo uint16) {
	b.lckRespPool.Lock()
	defer b.lckRespPool.Unlock()

	queue, ok := b.mapProtoNo2RespPool[protoNo]
	if !ok {
		queue = yx.NewLinkedQueue()
		b.mapProtoNo2RespPool[protoNo] = queue
	}

	b.pushToPool(queue, v)
}

func (b *protoBinder) popFromPool(queue *yx.LinkedQueue) (interface{}, bool) {
	if queue.GetSize() == 0 {
		return nil, false
	}

	v, err := queue.Dequeue()
	if err != nil {
		return nil, false
	}

	return v, true
}

func (b *protoBinder) pushToPool(queue *yx.LinkedQueue, v interface{}) {
	if queue.GetSize() >= MAX_REUSE_COUNT {
		return
	}

	queue.Enqueue(v)
}
