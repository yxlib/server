// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"errors"
	"sync"

	"github.com/yxlib/yx"
)

var (
	ErrWorkerExit     = errors.New("worker has exit")
	ErrRequestNil     = errors.New("request is nil")
	ErrRequestQueFull = errors.New("request queue is full")
	ErrTaskNil        = errors.New("task is nil")
	ErrTaskQueFull    = errors.New("task queue is full")
)

type WorkerOwner interface {
	OnWorkerClose(w *SessionWorker)
	OnHandleRequest(w *SessionWorker, req *Request, resp *Response) error
}

type SessionTask interface {
	Exec(worker *SessionWorker) error
}

type RequestInfo struct {
	Req  *Request
	Resp *Response
}

func NewRequestInfo(req *Request, resp *Response) *RequestInfo {
	return &RequestInfo{
		Req:  req,
		Resp: resp,
	}
}

type SessionWorker struct {
	id       uint64
	requests chan *RequestInfo
	tasks    chan SessionTask
	evtExit  *yx.Event
	owner    WorkerOwner
	lckStop  *sync.Mutex
	bStop    bool
	ec       *yx.ErrCatcher
}

func NewSessionWorker(id uint64, owner WorkerOwner, maxRequestNum uint16, maxTaskNum uint16) *SessionWorker {
	return &SessionWorker{
		id:       id,
		requests: make(chan *RequestInfo, maxRequestNum),
		tasks:    make(chan SessionTask, maxTaskNum),
		evtExit:  yx.NewEvent(),
		owner:    owner,
		lckStop:  &sync.Mutex{},
		bStop:    false,
		ec:       yx.NewErrCatcher("SessionWorker"),
	}
}

func (w *SessionWorker) GetID() uint64 {
	return w.id
}

func (w *SessionWorker) Start() {
	go w.loop()
}

func (w *SessionWorker) Stop() {
	w.lckStop.Lock()
	defer w.lckStop.Unlock()

	if w.bStop {
		return
	}

	w.bStop = true
	close(w.requests)
	close(w.tasks)
}

func (w *SessionWorker) IsStop() bool {
	w.lckStop.Lock()
	defer w.lckStop.Unlock()

	bStop := w.bStop
	return bStop
}

func (w *SessionWorker) WaitExit() {
	w.evtExit.Wait()
}

func (w *SessionWorker) AddRequest(r *RequestInfo) error {
	var err error = nil
	defer w.ec.DeferThrow("PushRequest", &err)

	if r == nil {
		err = ErrRequestNil
		return err
	}

	if w.IsStop() {
		err = ErrWorkerExit
		return err
	}

	if len(w.requests) == cap(w.requests) {
		err = ErrRequestQueFull
		return err
	}

	w.requests <- r
	return nil
}

func (w *SessionWorker) AddTask(t SessionTask) error {
	var err error = nil
	defer w.ec.DeferThrow("AddTask", &err)

	if t == nil {
		err = ErrTaskNil
		return err
	}

	if w.IsStop() {
		err = ErrWorkerExit
		return err
	}

	if len(w.tasks) == cap(w.tasks) {
		err = ErrTaskQueFull
		return err
	}

	w.tasks <- t
	return nil
}

func (w *SessionWorker) loop() {
	if w.owner == nil {
		return
	}

	bExit := false

	for {
		select {
		case reqInfo, ok := <-w.requests:
			bExit = w.handleRequest(reqInfo, ok)

		case t, ok := <-w.tasks:
			bExit = w.handleTask(t, ok)
		}

		if bExit {
			break
		}
	}

	w.evtExit.Send()
}

func (w *SessionWorker) handleRequest(reqInfo *RequestInfo, ok bool) bool {
	if !ok {
		return true
	}

	w.owner.OnHandleRequest(w, reqInfo.Req, reqInfo.Resp)
	// if err != nil {
	// 	w.errCatcher.Catch("handleRequest", &err)
	// }

	return false
}

func (w *SessionWorker) handleTask(t SessionTask, ok bool) bool {
	if !ok {
		return true
	}

	err := t.Exec(w)
	if err != nil {
		w.ec.Catch("handleTask", &err)
	}

	return false
}
