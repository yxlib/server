// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"encoding/json"

	"github.com/yxlib/yx"
)

type JsonInterceptor struct {
	ec *yx.ErrCatcher
}

func NewJsonInterceptor() *JsonInterceptor {
	return &JsonInterceptor{
		ec: yx.NewErrCatcher("JsonInterceptor"),
	}
}

func (i *JsonInterceptor) OnPreHandle(req *Request, resp *Response) (int32, error) {
	// request
	// reqType, err := ProtoBinder.GetRequestType(req.Mod, req.Cmd)
	// if err != nil {
	// 	return RESP_CODE_NOT_SUPPORT_PROTO, err
	// }

	// v := reflect.New(reqType)
	// reqData := v.Interface()
	if len(req.Payload) > 0 {
		reqData, err := ProtoBinder.GetRequest(req.Mod, req.Cmd)
		if err != nil {
			return RESP_CODE_NOT_SUPPORT_PROTO, i.ec.Throw("OnPreHandle", err)
		}

		err = json.Unmarshal(req.Payload, reqData)
		if err != nil {
			return RESP_CODE_UNMARSHAL_REQ_FAILED, i.ec.Throw("OnPreHandle", err)
		}

		req.ExtData = reqData
	}

	// response
	// respType, err := ProtoBinder.GetResponseType(req.Mod, req.Cmd)
	// if err != nil {
	// 	return RESP_CODE_NOT_SUPPORT_PROTO, err
	// }

	// v = reflect.New(respType)
	// resp.ExtData = v.Interface()

	respData, err := ProtoBinder.GetResponse(resp.Mod, resp.Cmd)
	if err == nil {
		resp.ExtData = respData
		// return RESP_CODE_NOT_SUPPORT_PROTO, err
	}

	// resp.ExtData = respData

	return 0, nil
}

func (i *JsonInterceptor) OnHandleCompletion(req *Request, resp *Response) (int32, error) {
	if resp.ExtData == nil {
		resp.Payload = []byte("")
		return 0, nil
	}

	respPayload, err := json.Marshal(resp.ExtData)
	if err != nil {
		return RESP_CODE_MARSHAL_RESP_FAILED, i.ec.Throw("OnHandleCompletion", err)
	}

	resp.Payload = respPayload
	return 0, nil
}

func (i *JsonInterceptor) OnResponseCompletion(req *Request, resp *Response) error {
	if req != nil && req.ExtData != nil {
		ProtoBinder.ReuseRequest(req.ExtData, req.Mod, req.Cmd)
	}

	if resp != nil && resp.ExtData != nil {
		ProtoBinder.ReuseResponse(resp.ExtData, resp.Mod, resp.Cmd)
	}

	return nil
}
