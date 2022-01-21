// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import "errors"

var (
	ErrPackFrameIsNil = errors.New("pack frame is nil")
)

const (
	RESP_CODE_SUCCESS         = 0
	RESP_CODE_SYS_UNKNOWN_ERR = -1
	RESP_CODE_SYS_UNKNOWN_MOD = -2
	RESP_CODE_SYS_UNKNOWN_CMD = -3
)

type PeerInfo struct {
	PeerType uint8
	PeerNo   uint16
}

func NewPeerInfo(peerType uint8, peerNo uint16) *PeerInfo {
	return &PeerInfo{
		PeerType: peerType,
		PeerNo:   peerNo,
	}
}

type Pack struct {
	SerialNo uint16
	Mod      uint16
	Cmd      uint16
	Src      *PeerInfo
	Tran     *PeerInfo
	Dst      *PeerInfo
	Payload  []byte
	ExtData  interface{}
}

func NewPack() *Pack {
	return &Pack{
		SerialNo: 0,
		Mod:      0,
		Cmd:      0,
		Src:      nil,
		Tran:     nil,
		Dst:      nil,
		Payload:  nil,
		ExtData:  nil,
	}
}

type Request struct {
	*Pack
	ConnId uint32
}

func NewRequest(connId uint32) *Request {
	return &Request{
		Pack:   NewPack(),
		ConnId: connId,
	}
}

type Response struct {
	*Pack
	Code int32
}

func NewResponse(req *Request) *Response {
	resp := &Response{
		Pack: NewPack(),
		Code: 0,
	}

	if req != nil {
		resp.SerialNo = req.SerialNo
		resp.Mod = req.Mod
		resp.Cmd = req.Cmd
		resp.Src = req.Dst
		resp.Tran = req.Tran
		resp.Dst = req.Src
	}

	return resp
}
