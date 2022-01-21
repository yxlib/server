// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"errors"

	"github.com/yxlib/yx"
)

var (
	ErrUnknown      = errors.New("unknown error")
	ErrProcNil      = errors.New("processor is nil")
	ErrProcExist    = errors.New("this cmd already has processor")
	ErrProcNotExist = errors.New("this cmd does not have processor")
)

type Processor func(req *Request, resp *Response) (int32, error)

// type Processor interface {
// 	OnHandleRequest(req *Request, resp *Response) (int32, error)
// }

type Service interface {
	// Call when handle a request.
	// @param req, the request.
	// @param resp, the response of the request.
	// @param bDebugMode, true mean use debug mode, false mean not use.
	// @return int32, the result code.
	// @return error, error.
	OnHandleRequest(req *Request, resp *Response, bDebugMode bool) (int32, error)
}

type BaseService struct {
	name             string
	mapCmd2Processor map[uint16]Processor
	errCatcher       *yx.ErrCatcher
}

func NewBaseService(name string) *BaseService {
	return &BaseService{
		name:             name,
		mapCmd2Processor: make(map[uint16]Processor),
		errCatcher:       yx.NewErrCatcher("BaseService(" + name + ")"),
	}
}

// Get the name of the service.
// @return string, the name.
func (s *BaseService) GetName() string {
	return s.name
}

// Add a processor bind with command cmd.
// @param p, the processor.
// @param cmd, the command of the processor.
// @return error, error.
func (s *BaseService) AddProcessor(p Processor, cmd uint16) error {
	var err error = nil
	defer s.errCatcher.DeferThrow("AddProcessor", &err)

	if p == nil {
		err = ErrProcNil
		return err
	}

	_, ok := s.mapCmd2Processor[cmd]
	if ok {
		err = ErrProcExist
		return err
	}

	s.mapCmd2Processor[cmd] = p
	return nil
}

// Get a processor by command cmd.
// @param cmd, the command of the processor.
// @return Processor, the processor.
// @return bool, true mean success, false mean failed.
func (s *BaseService) GetProcessor(cmd uint16) (Processor, bool) {
	p, ok := s.mapCmd2Processor[cmd]
	return p, ok
}

// Remove a processor by command cmd.
// @param cmd, the command of the processor.
// @return error, error.
func (s *BaseService) RemoveProcessor(cmd uint16) error {
	_, ok := s.mapCmd2Processor[cmd]
	if !ok {
		return s.errCatcher.Throw("RemoveProcessor", ErrProcNotExist)
	}

	delete(s.mapCmd2Processor, cmd)
	return nil
}

//================================================
//                    Service
//================================================
func (s *BaseService) OnHandleRequest(req *Request, resp *Response, bDebugMode bool) (int32, error) {
	var err error = nil
	defer s.errCatcher.DeferThrow("OnHandleRequest", &err)

	processor, ok := s.mapCmd2Processor[req.Cmd]
	if !ok {
		err = ErrProcNotExist
		return RESP_CODE_SYS_UNKNOWN_CMD, err
	}

	var code int32 = RESP_CODE_SYS_UNKNOWN_ERR
	err = ErrUnknown
	yx.RunDangerCode(func() {
		code, err = processor(req, resp)
	}, bDebugMode)

	return code, err
}
