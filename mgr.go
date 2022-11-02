// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"sync"
	"time"

	"github.com/yxlib/yx"
)

const SRV_WORKER_CHECK_INTV time.Duration = (2 * time.Second)

type SessionMgr struct {
	maxRequestNum    uint16
	maxTaskNum       uint16
	lckWorker        *sync.RWMutex
	mapID2Worker     map[uint64]*SessionWorker
	mapID2StopWorker map[uint64]*SessionWorker
	logger           *yx.Logger
}

func NewSessionMgr(maxRequestNum uint16, maxTaskNum uint16) *SessionMgr {
	return &SessionMgr{
		maxRequestNum:    maxRequestNum,
		maxTaskNum:       maxTaskNum,
		lckWorker:        &sync.RWMutex{},
		mapID2Worker:     make(map[uint64]*SessionWorker),
		mapID2StopWorker: make(map[uint64]*SessionWorker),
		logger:           yx.NewLogger("SessionMgr"),
	}
}

func (m *SessionMgr) Stop() {
	m.stopAllWorkers()
	m.waitAllStopWorkerExit()
}

func (m *SessionMgr) AddWorker(id uint64, owner WorkerOwner) *SessionWorker {
	m.logger.I("start worker ", id)

	m.lckWorker.Lock()
	defer m.lckWorker.Unlock()

	worker := NewSessionWorker(id, owner, m.maxRequestNum, m.maxTaskNum)
	m.mapID2Worker[id] = worker
	worker.Start()

	return worker
}

func (m *SessionMgr) RemoveWorker(id uint64) {
	m.lckWorker.Lock()
	defer m.lckWorker.Unlock()

	worker, ok := m.mapID2Worker[id]
	if ok {
		m.mapID2StopWorker[id] = worker
		delete(m.mapID2Worker, id)
		go m.WaitWorkerExit(id, worker)
	}
}

func (m *SessionMgr) WaitWorkerExit(id uint64, worker *SessionWorker) {
	worker.Stop()
	worker.WaitExit()

	m.lckWorker.Lock()
	defer m.lckWorker.Unlock()

	_, ok := m.mapID2StopWorker[id]
	if ok {
		delete(m.mapID2StopWorker, id)
	}

	m.logger.I("remove worker ", id)
}

func (m *SessionMgr) GetWorker(id uint64) (*SessionWorker, bool) {
	m.lckWorker.RLock()
	defer m.lckWorker.RUnlock()

	worker, ok := m.mapID2Worker[id]
	return worker, ok
}

func (m *SessionMgr) ForEachWorker(cb func(wid uint64, w *SessionWorker)) {
	workers := m.cloneWorkers()
	for _, worker := range workers {
		cb(worker.GetID(), worker)
	}
}

func (m *SessionMgr) cloneWorkers() []*SessionWorker {
	m.lckWorker.RLock()
	defer m.lckWorker.RUnlock()

	workers := make([]*SessionWorker, 0, len(m.mapID2Worker))
	for _, worker := range m.mapID2Worker {
		workers = append(workers, worker)
	}

	return workers
}

func (m *SessionMgr) stopAllWorkers() {
	m.lckWorker.Lock()
	defer m.lckWorker.Unlock()

	for id, worker := range m.mapID2Worker {
		m.mapID2StopWorker[id] = worker
		go m.WaitWorkerExit(id, worker)
	}

	m.mapID2Worker = make(map[uint64]*SessionWorker)
}

func (m *SessionMgr) waitAllStopWorkerExit() {
	ticker := time.NewTicker(SRV_WORKER_CHECK_INTV)

	for {
		<-ticker.C
		if m.isAllStopWorkerExit() {
			break
		}
	}

	ticker.Stop()
}

func (m *SessionMgr) isAllStopWorkerExit() bool {
	m.lckWorker.Lock()
	defer m.lckWorker.Unlock()

	return len(m.mapID2StopWorker) == 0
}
