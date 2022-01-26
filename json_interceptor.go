// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"encoding/json"
	"reflect"
)

type JsonInterceptor struct {
}

func (i *JsonInterceptor) OnPreHandle(req *Request, resp *Response) (int32, error) {
	// request
	reqType, err := ProtoBinder.GetRequestType(req.Mod, req.Cmd)
	if err != nil {
		return RESP_CODE_NOT_SUPPORT_PROTO, err
	}

	v := reflect.New(reqType)
	reqData := v.Interface()
	err = json.Unmarshal(req.Payload, reqData)
	if err != nil {
		return RESP_CODE_UNMARSHAL_REQ_FAILED, err
	}

	req.ExtData = reqData

	// response
	respType, err := ProtoBinder.GetResponseType(req.Mod, req.Cmd)
	if err != nil {
		return RESP_CODE_NOT_SUPPORT_PROTO, err
	}

	v = reflect.New(respType)
	resp.ExtData = v.Interface()

	return 0, nil
}

func (i *JsonInterceptor) OnHandleCompletion(req *Request, resp *Response) (int32, error) {
	respPayload, err := json.Marshal(resp.ExtData)
	if err != nil {
		return RESP_CODE_MARSHAL_RESP_FAILED, err
	}

	resp.Payload = respPayload
	return 0, nil
}

func (i *JsonInterceptor) OnResponseCompletion(req *Request, resp *Response) error {
	return nil
}
