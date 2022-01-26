// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

type serviceBinder struct {
	mapNameService map[string]Service
}

var ServiceBinder = &serviceBinder{
	mapNameService: make(map[string]Service),
}

// Register service.
// @param s, the service.
func (r *serviceBinder) BindService(s Service) {
	if s == nil {
		return
	}

	name := s.GetName()
	r.mapNameService[name] = s
}

// Get service by name.
// @param name, the service name.
// @return server.Service, the service with the name.
// @return bool, true mean success, false mean failed.
func (r *serviceBinder) GetService(name string) (Service, bool) {
	s, ok := r.mapNameService[name]
	return s, ok
}
