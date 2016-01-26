//
// Copyright (c) 2016 Christian Pointner <equinox@spreadspace.org>
//               2016 Markus Grüneis <gimpf@gimpf.org>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// * Neither the name of whawty.auth nor the names of its
//   contributors may be used to endorse or promote products derived from
//   this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//

// Package sasl implements the authentication protocol as used by salsauthd.
package sasl

import (
	"net"
)

const (
	// MaxRequestLength is the maximum length allowed for login, password
	// service and realm
	MaxRequestLength = 256
)

// Request contains all values for a saslauthd authentication request.
type Request struct {
	Login    string
	Password string
	Service  string
	Realm    string
}

// Marshal encodes the request values into a byte slice. The format is
// compatible to the requests as expected by salsauthd.
func (r *Request) Marshal() (data []byte, err error) {
	// TODO: implemente this
	return
}

// Unmarshal decodes the request values from it's byte representaion.
func (r *Request) UnMarshal(data []byte) (err error) {
	// TODO: implemente this
	return
}

// Response holds the result as well as the message as returned by the
// authentication provider.
type Response struct {
	Result  bool
	Message string
}

// Marshal encodes the response into a byte slice. The format is the same
// as used by salsauthd.
func (r *Response) Marshal() (data []byte, err error) {
	// TODO: implemente this
	return
}

// Unmarshal decodes the response from the it's byte representaion.
func (r *Response) UnMarshal(data []byte) (err error) {
	// TODO: implemente this
	return
}

// AuthCB is the function signature of callbacks as used by the server to
// handle authentication requests.
type AuthCB func(login, password, service, realm string) (ok bool, msg string, err error)

// Server holds all information needed to run the server. Use NewServer to
// create the struct.
type Server struct {
	sockPath string
	cb       AuthCB
	ln       net.Listener
}

// NewServer creates a server struct and starts listening on the unix socket
// as specified by socketpath. cb is the callback function which will get
// called for any authentication request.
func NewServer(socketpath string, cb AuthCB) (s *Server, err error) {
	s = &Server{}
	s.sockPath = socketpath
	s.cb = cb
	if s.ln, err = net.Listen("unix", s.sockPath); err != nil {
		return
	}
	return
}

func (s *Server) handleConnection(conn net.Conn) {
	// TODO: receive request
	//       unmarshal request
	//       call auth callback
	//       marshal response
	//       send response
}

// Run actually runs the server. In calls Accept() on the server socket and
// runs go-routines for new connections.
func (s *Server) Run() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			// TODO: handle errors!!!
			continue
		}
		go s.handleConnection(conn)
	}
}

// Client holds all information needed to send and authentication request as well as to
// receive responses from saslauthd compatible servers. Use NewClient to create the socket.
type Client struct {
	sockPath string
}

// NewClient creates a client struct.
func NewClient(socketpath string) (c *Client, err error) {
	c = &Client{}
	c.sockPath = socketpath
	return
}

// Auth connects to the server socket and sends a authentication request.
func (c *Client) Auth(login, password, service, realm string) (ok bool, msg string, err error) {
	// TODO: connect to server
	//       marshal request
	//       send request
	//       receive response
	//       unmarshal response
	return
}
