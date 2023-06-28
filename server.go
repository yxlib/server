// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/yxlib/yx"
)

var (
	ErrSrvNoNeedResponse = errors.New("no need response")
	ErrSrvInterNil       = errors.New("interceptor is nil")
	ErrSrvInterExist     = errors.New("interceptor is exist")
	ErrSrvInterNotExist  = errors.New("interceptor is not exist")
	ErrSrvServNil        = errors.New("service is nil")
	ErrSrvServExist      = errors.New("service is exist")
	ErrSrvServNotExist   = errors.New("service is not exist")
	ErrSrvModNotExist    = errors.New("mod is not exist")
	ErrSrvNetClose       = errors.New("net is close")
)

const (
	INTERCPT_STEP_PRE_HANDLE   = 1
	INTERCPT_STEP_HANDLE_COMPL = 2
	INTERCPT_STEP_RESP_COMPL   = 3
)

type Interceptor interface {
	// Call before handle the request.
	// @param req, the request
	// @param resp, the response of the request
	// @return int32, result code.
	// @return error, error.
	OnPreHandle(req Request, resp Response) (int32, error)

	// Call after handle the request.
	// @param req, the request
	// @param resp, the response of the request
	// @return int32, result code.
	// @return error, error.
	OnHandleCompletion(req Request, resp Response) (int32, error)

	// Call after send the response.
	// @param req, the request
	// @param resp, the response of the request
	// @return error, error.
	OnResponseCompletion(req Request, resp Response) error
}

type InterceptorList = []Interceptor

type Server interface {
	SetEnumerable(bEnumerable bool, protoNo uint32, mapProcName2ProtoNo map[string]uint32)
	AddService(srv Service, mod uint32) error
}

type EnumerationResp struct {
	MapProcName2ProtoNo map[string]uint32 `json:"func_mapper"`
}

type BaseServer struct {
	name                string
	bDebugMode          bool
	bEnumerable         bool
	enumProtoNo         uint32
	mapProcName2ProtoNo map[string]uint32
	mapProtoNo2ProcName map[uint32]string
	mapMod2Service      map[uint32]Service
	globalInterceptors  InterceptorList
	mapMod2Interceptors map[uint32]InterceptorList

	ec     *yx.ErrCatcher
	logger *yx.Logger
}

func NewBaseServer(name string) *BaseServer {
	tag := "Server(" + name + ")"

	s := &BaseServer{
		name:                name,
		bDebugMode:          false,
		bEnumerable:         false,
		enumProtoNo:         0,
		mapProcName2ProtoNo: nil,
		mapProtoNo2ProcName: nil,
		mapMod2Service:      make(map[uint32]Service),
		globalInterceptors:  make(InterceptorList, 0),
		mapMod2Interceptors: make(map[uint32]InterceptorList),
		ec:                  yx.NewErrCatcher(tag),
		logger:              yx.NewLogger(tag),
	}

	return s
}

func (s *BaseServer) SetEnumerable(bEnumerable bool, protoNo uint32, mapProcName2ProtoNo map[string]uint32) {
	s.enumProtoNo = protoNo
	s.mapProcName2ProtoNo = mapProcName2ProtoNo
	s.bEnumerable = (bEnumerable && len(mapProcName2ProtoNo) > 0)

	s.mapProtoNo2ProcName = make(map[uint32]string)
	for procName, protoNo := range s.mapProcName2ProtoNo {
		s.mapProtoNo2ProcName[protoNo] = procName
	}
}

func (s *BaseServer) GetProcMapper() map[string]uint32 {
	return s.mapProcName2ProtoNo
}

// Add a service bind with module mod.
// @param srv, the service.
// @param mod, the module of the service.
// @return error, error.
func (s *BaseServer) AddService(srv Service, mod uint32) error {
	var err error = nil
	defer s.ec.DeferThrow("AddService", &err)

	if srv == nil {
		err = ErrSrvServNil
		return err
	}

	_, ok := s.mapMod2Service[mod]
	if ok {
		err = ErrSrvServExist
		return err
	}

	s.mapMod2Service[mod] = srv
	return nil
}

// Get a service by module mod.
// @param mod, the module of the service.
// @return *Service, the service.
// @return bool, true mean success, false mean failed.
func (s *BaseServer) GetService(mod uint32) (Service, bool) {
	srv, ok := s.mapMod2Service[mod]
	return srv, ok
}

// Remove a service by module mod.
// @param mod, the module of the service.
// @return error, error.
func (s *BaseServer) RemoveService(mod uint32) error {
	_, ok := s.mapMod2Service[mod]
	if !ok {
		return s.ec.Throw("RemoveService", ErrSrvServNotExist)
	}

	delete(s.mapMod2Service, mod)
	return nil
}

// Open or close dubug mode.
// @param bDebugMode, true mean use debug mode, false mean close debug mode.
func (s *BaseServer) SetDebugMode(bDebugMode bool) {
	s.bDebugMode = bDebugMode
}

// Add a global interceptor.
// @param it, the interceptor.
// @return error, error.
func (s *BaseServer) AddGlobalInterceptor(it Interceptor) error {
	var err error = nil
	defer s.ec.DeferThrow("AddGlobalInterceptor", &err)

	if it == nil {
		err = ErrSrvInterNil
		return err
	}

	s.globalInterceptors = append(s.globalInterceptors, it)
	return nil
}

// Remove a global interceptor.
// @param it, the interceptor.
// @return error, error.
func (s *BaseServer) RemoveGlobalInterceptor(it Interceptor) error {
	var err error = nil
	defer s.ec.DeferThrow("RemoveGlobalInterceptor", &err)

	if it == nil {
		err = ErrSrvInterNil
		return err
	}

	for i, itTmp := range s.globalInterceptors {
		if itTmp == it {
			s.globalInterceptors = append(s.globalInterceptors[:i], s.globalInterceptors[i+1:]...)
			return nil
		}
	}

	err = ErrSrvInterNotExist
	return err
}

// Add a module interceptor.
// @param it, the interceptor.
// @param mod, module of the interceptor.
// @return error, error.
func (s *BaseServer) AddModInterceptor(it Interceptor, mod uint32) error {
	var err error = nil
	defer s.ec.DeferThrow("AddModInterceptor", &err)

	if it == nil {
		err = ErrSrvInterNil
		return err
	}

	_, ok := s.mapMod2Interceptors[mod]
	if !ok {
		s.mapMod2Interceptors[mod] = make(InterceptorList, 0)
	}

	s.mapMod2Interceptors[mod] = append(s.mapMod2Interceptors[mod], it)
	return nil
}

// Remove a module interceptor.
// @param it, the interceptor.
// @param mod, module of the interceptor.
// @return error, error.
func (s *BaseServer) RemoveModInterceptor(it Interceptor, mod uint32) error {
	var err error = nil
	defer s.ec.DeferThrow("RemoveModInterceptor", &err)

	if it == nil {
		err = ErrSrvInterNil
		return err
	}

	list, ok := s.mapMod2Interceptors[mod]
	if !ok {
		err = ErrSrvModNotExist
		return err
	}

	for i, itTmp := range list {
		if itTmp == it {
			s.mapMod2Interceptors[mod] = append(list[:i], list[i+1:]...)
			return nil
		}
	}

	err = ErrSrvInterNotExist
	return err
}

// Handle a request.
// @param req, the request.
// @param resp, the response for this request.
// @return error, error.
func (s *BaseServer) HandleRequest(req Request, resp Response) error {
	if s.isEnumReq(req) {
		return s.handleEnumReq(req, resp)
	}

	code, err := s.preHandle(req, resp)
	if err != nil {
		resp.SetResCode(code)
		s.ec.Catch("HandleRequest", &err)
		s.logger.E("PreHandle failed!! (", req.GetProtoNo(), "): SNo. ", req.GetSeqNum(), ", resCode ", resp.GetResCode())
		return err
	}

	err = s.handleRequestDirect(req, resp)

	s.ec.Catch("HandleRequest", &err)
	return err
}

// func (s *BaseServer) HandleHttpRequest(req *Request, resp *Response) error {
// 	code, err := s.preHandle(req, resp)
// 	if err != nil {
// 		resp.Code = code
// 		s.ec.Catch("HandleHttpRequest", &err)
// 		s.logger.E("Http request (", req.Mod, ", ", req.Cmd, "): serial No. ", req.SerialNo, ", resCode ", resp.Code)
// 		return err
// 	}

// 	err = s.handleRequestImpl(req, resp)
// 	if err != nil {
// 		s.ec.Catch("HandleHttpRequest", &err)
// 	}

// 	if resp.Code != RESP_CODE_SUCCESS {
// 		s.logger.E("Http request (", req.Mod, ", ", req.Cmd, "): serial No. ", req.SerialNo, ", resCode ", resp.Code)
// 	}

// 	// handle response completion
// 	err2 := s.responseCompletion(req, resp)
// 	if err2 != nil {
// 		s.ec.Catch("HandleHttpRequest", &err2)
// 	}

// 	return err
// }

//================================================
//                 private
//================================================
func (s *BaseServer) interceptHandle(it Interceptor, step uint8, req Request, resp Response) (int32, error) {
	var code int32 = 0
	var err error = nil

	if step == INTERCPT_STEP_PRE_HANDLE {
		code, err = it.OnPreHandle(req, resp)
	} else if step == INTERCPT_STEP_HANDLE_COMPL {
		code, err = it.OnHandleCompletion(req, resp)
	} else if step == INTERCPT_STEP_RESP_COMPL {
		err = it.OnResponseCompletion(req, resp)
	}

	return code, s.ec.Throw("interceptHandle", err)
}

func (s *BaseServer) listInterceptHandle(list InterceptorList, step uint8, req Request, resp Response) (int32, error) {
	if len(list) == 0 {
		return RESP_CODE_SUCCESS, nil
	}

	for _, it := range list {
		code, err := s.interceptHandle(it, step, req, resp)
		if err != nil {
			return code, s.ec.Throw("listInterceptHandle", err)
		}
	}

	return RESP_CODE_SUCCESS, nil
}

func (s *BaseServer) reverseListInterceptHandle(list InterceptorList, step uint8, req Request, resp Response) (int32, error) {
	if len(list) == 0 {
		return RESP_CODE_SUCCESS, nil
	}

	for i := len(list) - 1; i >= 0; i-- {
		code, err := s.interceptHandle(list[i], step, req, resp)
		if err != nil {
			return code, s.ec.Throw("reverseListInterceptHandle", err)
		}
	}

	return RESP_CODE_SUCCESS, nil
}

func (s *BaseServer) intercept(step uint8, req Request, resp Response) (int32, error) {
	// global first
	code, err := s.listInterceptHandle(s.globalInterceptors, step, req, resp)
	if err != nil {
		return code, s.ec.Throw("intercept", err)
	}

	if req != nil {
		protoNo := req.GetProtoNo()
		mod := GetMod(protoNo)
		list, ok := s.mapMod2Interceptors[mod]
		if ok {
			code, err := s.listInterceptHandle(list, step, req, resp)
			if err != nil {
				return code, s.ec.Throw("intercept", err)
			}
		}
	}

	return RESP_CODE_SUCCESS, nil
}

func (s *BaseServer) reverseIntercept(step uint8, req Request, resp Response) (int32, error) {
	// mod first, reverse visit
	if req != nil {
		protoNo := req.GetProtoNo()
		mod := GetMod(protoNo)
		list, ok := s.mapMod2Interceptors[mod]
		if ok {
			code, err := s.reverseListInterceptHandle(list, step, req, resp)
			if err != nil {
				return code, s.ec.Throw("reverseIntercept", err)
			}
		}
	}

	code, err := s.reverseListInterceptHandle(s.globalInterceptors, step, req, resp)
	if err != nil {
		return code, s.ec.Throw("reverseIntercept", err)
	}

	return RESP_CODE_SUCCESS, nil
}

func (s *BaseServer) preHandle(req Request, resp Response) (int32, error) {
	code, err := s.intercept(INTERCPT_STEP_PRE_HANDLE, req, resp)
	return code, s.ec.Throw("preHandle", err)
}

func (s *BaseServer) handleCompletion(req Request, resp Response) (int32, error) {
	code, err := s.reverseIntercept(INTERCPT_STEP_HANDLE_COMPL, req, resp)
	return code, s.ec.Throw("handleCompletion", err)
}

func (s *BaseServer) responseCompletion(req Request, resp Response) error {
	_, err := s.reverseIntercept(INTERCPT_STEP_RESP_COMPL, req, resp)
	return s.ec.Throw("responseCompletion", err)
}

// func (s *BaseServer) loop() error {
// 	req, err := s.srvNet.ReadRequest()
// 	if err != nil {
// 		return ErrSrvNetClose
// 	}

// 	resp := NewResponse(req)

// 	code, err := s.preHandle(req, resp)
// 	if err != nil {
// 		resp.Code = code
// 		s.ec.Catch("loop", &err)
// 		s.logger.E("PreHandle failed!! (", req.Mod, ", ", req.Cmd, "): SNo. ", req.SerialNo, ", resCode ", resp.Code)

// 		s.srvNet.WriteResponse(resp)
// 		return err
// 	}

// 	if !s.bWorkerMode {
// 		err = s.handleRequest(req, resp)
// 		return err
// 	}

// 	worker, ok := s.mgr.GetWorker(req.ConnId)
// 	if !ok || worker.IsStop() {
// 		worker = s.mgr.AddWorker(req.ConnId, s)
// 	}

// 	reqInfo := NewRequestInfo(req, resp)
// 	err = worker.AddRequest(reqInfo)
// 	if err != nil {
// 		s.ec.Catch("loop", &err)
// 		s.mgr.RemoveWorker(req.ConnId)
// 	}

// 	return err
// }

func (s *BaseServer) handleRequestDirect(req Request, resp Response) error {
	s.printRequestInfo(req)

	err := s.handleRequestImpl(req, resp)
	// if err != nil {
	// 	s.ec.Catch("handleRequest", &err)
	// }

	code := resp.GetResCode()
	if code != RESP_CODE_SUCCESS {
		s.printResponseInfo(yx.LOG_LV_ERROR, resp)
	} else {
		s.printResponseInfo(yx.LOG_LV_INFO, resp)
	}

	// handle response completion
	err2 := s.responseCompletion(req, resp)
	if err2 != nil {
		s.ec.Catch("handleRequestDirect", &err2)
	}

	return s.ec.Throw("handleRequestDirect", err)
}

func (s *BaseServer) printRequestInfo(req Request) {
	s.logger.I("=====> START REQUEST:")

	protoNo := req.GetProtoNo()
	logs := make([][]interface{}, 0)
	logs = append(logs, yx.LogArgs("[L] *******************************"))
	logs = append(logs, yx.LogArgs("[0] Sequence No.: ", req.GetSeqNum()))
	logs = append(logs, yx.LogArgs("[1] Proto No.: ", protoNo))

	procName, ok := s.mapProtoNo2ProcName[protoNo]
	if ok {
		logs = append(logs, yx.LogArgs("[2] Processor: ", procName))
	}

	logs = append(logs, yx.LogArgs("[L] *******************************"))

	s.logger.Detail(yx.LOG_LV_INFO, logs)
}

func (s *BaseServer) printResponseInfo(lv int, resp Response) {
	s.logger.I("=====> START RESPONSE:")

	protoNo := resp.GetProtoNo()
	logs := make([][]interface{}, 0)
	logs = append(logs, yx.LogArgs("[L] *******************************"))
	logs = append(logs, yx.LogArgs("[0] Sequence No.: ", resp.GetSeqNum()))
	logs = append(logs, yx.LogArgs("[1] Proto No.: ", protoNo))
	logs = append(logs, yx.LogArgs("[2] Result Code: ", resp.GetResCode()))
	logs = append(logs, yx.LogArgs("[L] *******************************"))

	s.logger.Detail(lv, logs)
}

func (s *BaseServer) handleRequestImpl(req Request, resp Response) error {
	var err error = nil
	defer s.ec.DeferThrow("handleRequestImpl", &err)

	protoNo := req.GetProtoNo()
	mod := GetMod(protoNo)
	serv, ok := s.GetService(mod)
	if !ok {
		resp.SetResCode(RESP_CODE_SYS_UNKNOWN_MOD)
		err = fmt.Errorf("service not exist for mod: %d", mod)
		return err
	}

	// handle
	code, err := serv.OnHandleRequest(req, resp, s.bDebugMode)
	resp.SetResCode(code)
	if err != nil {
		data := resp.GetData()
		if len(data) == 0 {
			resp.SetData([]byte(err.Error()))
		}

		return err
	}

	// handle completion
	code, err = s.handleCompletion(req, resp)
	if err != nil {
		resp.SetResCode(code)
		return err
	}

	return nil
}

func (s *BaseServer) isEnumReq(req Request) bool {
	if !s.bEnumerable {
		return false
	}

	protoNo := req.GetProtoNo()
	return (s.enumProtoNo == protoNo)
}

func (s *BaseServer) handleEnumReq(req Request, resp Response) error {
	respObj := &EnumerationResp{
		MapProcName2ProtoNo: s.mapProcName2ProtoNo,
	}

	data, err := json.Marshal(respObj)
	if err != nil {
		return err
	}

	resp.SetData(data)
	return nil
}
