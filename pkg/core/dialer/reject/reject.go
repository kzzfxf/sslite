// Copyright 2023 kzzfxf
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reject

import (
	"errors"
	"net"
)

var (
	ErrDialRejected = errors.New("dial rejected")
)

type Reject struct {
}

// NewReject
func NewReject() (d *Reject) {
	return &Reject{}
}

// Addr
func (*Reject) Addr() (addr string) {
	return ""
}

// Dial
func (*Reject) Dial(network string, addr string) (conn net.Conn, err error) {
	return nil, ErrDialRejected
}

// Close
func (*Reject) Close() (err error) {
	return
}
