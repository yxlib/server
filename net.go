package server

import (
	"errors"

	"github.com/yxlib/yx"
)

var (
	ErrNetReadQueClose = errors.New("read queue closed")
)

type Net interface {
	// Read a request.
	// @return *Request, a request or nil.
	// @return error, error.
	ReadRequest() (*Request, error)

	// Write a response.
	// @param resp, the response
	// @return error, error.
	WriteResponse(resp *Response) error

	// Close the net.
	Close()
}

type BaseNet struct {
	chanRequest chan *Request
	logger      *yx.Logger
	ec          *yx.ErrCatcher
}

func NewBaseNet(maxReadQue uint32) *BaseNet {
	return &BaseNet{
		chanRequest: make(chan *Request, maxReadQue),
		logger:      yx.NewLogger("BaseNet"),
		ec:          yx.NewErrCatcher("BaseNet"),
	}
}

// ServerNet
func (n *BaseNet) ReadRequest() (*Request, error) {
	req, ok := <-n.chanRequest
	if !ok {
		return nil, ErrNetReadQueClose
	}

	return req, nil
}

func (n *BaseNet) WriteResponse(resp *Response) error {
	return nil
}

func (n *BaseNet) Close() {
	close(n.chanRequest)
}

func (n *BaseNet) AddRequest(req *Request) {
	n.chanRequest <- req
}
