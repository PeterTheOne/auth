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

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/whawty/auth/store"
)

type webAuthenticateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type webAuthenticateResponse struct {
	Session     string    `json:"session,omitempty"`
	Username    string    `json:"username"`
	IsAdmin     bool      `json:"admin"`
	LastChanged time.Time `json:"lastchanged"`
	Error       string    `json:"error,omitempty"`
}

func handleWebAuthenticate(store *StoreChan, sessions *webSessionFactory, w http.ResponseWriter, r *http.Request) {
	wdl.Printf("web-api: got AUTHENTICATE request from %s", r.RemoteAddr)

	decoder := json.NewDecoder(r.Body)
	reqdata := &webAuthenticateRequest{}
	respdata := &webAuthenticateResponse{}

	if err := decoder.Decode(reqdata); err != nil {
		respdata.Error = fmt.Sprintf("Error parsing JSON response: %s", err)
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}

	if reqdata.Username == "" || reqdata.Password == "" {
		respdata.Error = "empty username or password is not allowed"
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}

	ok, isAdmin, _, lastChanged, err := store.Authenticate(reqdata.Username, reqdata.Password)
	if err != nil || !ok {
		respdata.Error = "authentication failed"
		if err != nil {
			respdata.Error = err.Error()
		}
		sendWebResponse(w, http.StatusUnauthorized, respdata)
		return
	}

	respdata.Username = reqdata.Username
	respdata.IsAdmin = isAdmin
	respdata.LastChanged = lastChanged
	status := http.StatusOK
	status, respdata.Error, respdata.Session = sessions.Generate(reqdata.Username, isAdmin)
	sendWebResponse(w, status, respdata)
}

type webAddRequest struct {
	Session  string `json:"session"`
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"admin"`
}

type webAddResponse struct {
	Username string `json:"username"`
	IsAdmin  bool   `json:"admin"`
	Error    string `json:"error,omitempty"`
}

func handleWebAdd(store *StoreChan, sessions *webSessionFactory, w http.ResponseWriter, r *http.Request) {
	wdl.Printf("web-api: got ADD request from %s", r.RemoteAddr)

	decoder := json.NewDecoder(r.Body)
	reqdata := &webAddRequest{}
	respdata := &webAddResponse{}

	if err := decoder.Decode(reqdata); err != nil {
		respdata.Error = fmt.Sprintf("Error parsing JSON response: %s", err)
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}

	if reqdata.Session == "" || reqdata.Username == "" || reqdata.Password == "" {
		respdata.Error = "empty session, username or password is not allowed"
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}

	status, errorStr, username, isAdmin := sessions.Check(reqdata.Session)
	if status != http.StatusOK {
		respdata.Error = errorStr
		sendWebResponse(w, status, respdata)
		return
	}

	if !isAdmin {
		respdata.Error = "only admins are allowed to add users"
		sendWebResponse(w, http.StatusForbidden, respdata)
		return
	}

	wdl.Printf("admin '%s' want's to add user '%s' and admin status: %t", username, reqdata.Username, reqdata.IsAdmin)

	if err := store.Add(reqdata.Username, reqdata.Password, reqdata.IsAdmin); err != nil {
		respdata.Error = err.Error()
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}
	respdata.Username = reqdata.Username
	respdata.IsAdmin = reqdata.IsAdmin
	sendWebResponse(w, http.StatusOK, respdata)
}

type webRemoveRequest struct {
	Session  string `json:"session"`
	Username string `json:"username"`
}

type webRemoveResponse struct {
	Username string `json:"username"`
	Error    string `json:"error,omitempty"`
}

func handleWebRemove(store *StoreChan, sessions *webSessionFactory, w http.ResponseWriter, r *http.Request) {
	wdl.Printf("web-api: got REMOVE request from %s", r.RemoteAddr)

	decoder := json.NewDecoder(r.Body)
	reqdata := &webRemoveRequest{}
	respdata := &webRemoveResponse{}

	if err := decoder.Decode(reqdata); err != nil {
		respdata.Error = fmt.Sprintf("Error parsing JSON response: %s", err)
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}

	if reqdata.Session == "" || reqdata.Username == "" {
		respdata.Error = "empty session or username is not allowed"
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}

	status, errorStr, username, isAdmin := sessions.Check(reqdata.Session)
	if status != http.StatusOK {
		respdata.Error = errorStr
		sendWebResponse(w, status, respdata)
		return
	}

	if !isAdmin {
		respdata.Error = "only admins are allowed to remove users"
		sendWebResponse(w, http.StatusForbidden, respdata)
		return
	}

	wdl.Printf("admin '%s' want's to remove user '%s'", username, reqdata.Username)

	if err := store.Remove(reqdata.Username); err != nil {
		respdata.Error = err.Error()
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}
	respdata.Username = reqdata.Username
	sendWebResponse(w, http.StatusOK, respdata)
}

type webUpdateRequest struct {
	Session     string `json:"session"`
	Username    string `json:"username"`
	OldPassword string `json:"oldpassword"`
	NewPassword string `json:"newpassword"`
}

type webUpdateResponse struct {
	Username string `json:"username"`
	Error    string `json:"error,omitempty"`
}

func handleWebUpdate(store *StoreChan, sessions *webSessionFactory, w http.ResponseWriter, r *http.Request) {
	wdl.Printf("web-api: got UPDATE request from %s", r.RemoteAddr)

	decoder := json.NewDecoder(r.Body)
	reqdata := &webUpdateRequest{}
	respdata := &webUpdateResponse{}

	if err := decoder.Decode(reqdata); err != nil {
		respdata.Error = fmt.Sprintf("Error parsing JSON response: %s", err)
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}

	if reqdata.Username == "" {
		respdata.Error = "empty username is not allowed"
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}

	if reqdata.Session != "" && reqdata.OldPassword == "" {
		if reqdata.NewPassword == "" {
			respdata.Error = "empty newpassword is not allowed when using session based authentication"
			sendWebResponse(w, http.StatusBadRequest, respdata)
			return
		}

		status, errorStr, username, isAdmin := sessions.Check(reqdata.Session)
		if status != http.StatusOK {
			respdata.Error = errorStr
			sendWebResponse(w, status, respdata)
			return
		}

		if !isAdmin && username != reqdata.Username {
			respdata.Error = "only admins are allowed to update any users' password"
			sendWebResponse(w, http.StatusForbidden, respdata)
			return
		}
		wdl.Printf("user '%s' want's to update user '%s', using a valid session", username, reqdata.Username)
	} else if reqdata.Session == "" && reqdata.OldPassword != "" {
		ok, _, _, _, err := store.Authenticate(reqdata.Username, reqdata.OldPassword)
		if err != nil || !ok {
			respdata.Error = "authentication failed"
			if err != nil {
				respdata.Error = err.Error()
			}
			sendWebResponse(w, http.StatusUnauthorized, respdata)
			return
		}
		if reqdata.NewPassword == "" {
			reqdata.NewPassword = reqdata.OldPassword
		}
		wdl.Printf("update user '%s', using current(old) password", reqdata.Username)
	} else {
		respdata.Error = "exactly one of session or old-password must be supplied"
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}

	if err := store.Update(reqdata.Username, reqdata.NewPassword); err != nil {
		respdata.Error = err.Error()
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}
	respdata.Username = reqdata.Username
	sendWebResponse(w, http.StatusOK, respdata)
}

type webSetAdminRequest struct {
	Session  string `json:"session"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"admin"`
}

type webSetAdminResponse struct {
	Username string `json:"username"`
	IsAdmin  bool   `json:"admin"`
	Error    string `json:"error,omitempty"`
}

func handleWebSetAdmin(store *StoreChan, sessions *webSessionFactory, w http.ResponseWriter, r *http.Request) {
	wdl.Printf("web-api: got SET_ADMIN request from %s", r.RemoteAddr)

	decoder := json.NewDecoder(r.Body)
	reqdata := &webSetAdminRequest{}
	respdata := &webSetAdminResponse{}

	if err := decoder.Decode(reqdata); err != nil {
		respdata.Error = fmt.Sprintf("Error parsing JSON response: %s", err)
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}

	if reqdata.Session == "" || reqdata.Username == "" {
		respdata.Error = "empty session or username is not allowed"
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}

	status, errorStr, username, isAdmin := sessions.Check(reqdata.Session)
	if status != http.StatusOK {
		respdata.Error = errorStr
		sendWebResponse(w, status, respdata)
		return
	}

	if !isAdmin {
		respdata.Error = "only admins are allowed to change the admin status of users"
		sendWebResponse(w, http.StatusForbidden, respdata)
		return
	}

	wdl.Printf("admin '%s' want's to set admin status of user '%s' to %t", username, reqdata.Username, reqdata.IsAdmin)

	if err := store.SetAdmin(reqdata.Username, reqdata.IsAdmin); err != nil {
		respdata.Error = err.Error()
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}
	respdata.Username = reqdata.Username
	respdata.IsAdmin = reqdata.IsAdmin
	sendWebResponse(w, http.StatusOK, respdata)
}

type webListRequest struct {
	Session string `json:"session"`
}

type webListResponse struct {
	List  store.UserList `json:"list"`
	Error string         `json:"error,omitempty"`
}

func handleWebList(store *StoreChan, sessions *webSessionFactory, w http.ResponseWriter, r *http.Request) {
	wdl.Printf("web-api: got LIST request from %s", r.RemoteAddr)

	decoder := json.NewDecoder(r.Body)
	reqdata := &webListRequest{}
	respdata := &webListResponse{}

	if err := decoder.Decode(reqdata); err != nil {
		respdata.Error = fmt.Sprintf("Error parsing JSON response: %s", err)
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}

	if reqdata.Session == "" {
		respdata.Error = "empty session is not allowed"
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}

	status, errorStr, username, isAdmin := sessions.Check(reqdata.Session)
	if status != http.StatusOK {
		respdata.Error = errorStr
		sendWebResponse(w, status, respdata)
		return
	}

	if !isAdmin {
		respdata.Error = "only admins are allowed to list users"
		sendWebResponse(w, http.StatusForbidden, respdata)
		return
	}

	wdl.Printf("admin '%s' want's to list all supported users", username)

	var err error
	if respdata.List, err = store.List(); err != nil {
		respdata.Error = err.Error()
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}
	sendWebResponse(w, http.StatusOK, respdata)
}

type webListFullRequest struct {
	Session string `json:"session"`
}

type webListFullResponse struct {
	List  store.UserListFull `json:"list"`
	Error string             `json:"error,omitempty"`
}

func handleWebListFull(store *StoreChan, sessions *webSessionFactory, w http.ResponseWriter, r *http.Request) {
	wdl.Printf("web-api: got LIST_FULL request from %s", r.RemoteAddr)

	decoder := json.NewDecoder(r.Body)
	reqdata := &webListFullRequest{}
	respdata := &webListFullResponse{}

	if err := decoder.Decode(reqdata); err != nil {
		respdata.Error = fmt.Sprintf("Error parsing JSON response: %s", err)
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}

	if reqdata.Session == "" {
		respdata.Error = "empty session is not allowed"
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}

	status, errorStr, username, isAdmin := sessions.Check(reqdata.Session)
	if status != http.StatusOK {
		respdata.Error = errorStr
		sendWebResponse(w, status, respdata)
		return
	}

	if !isAdmin {
		respdata.Error = "only admins are allowed to list users"
		sendWebResponse(w, http.StatusForbidden, respdata)
		return
	}

	wdl.Printf("admin '%s' want's to list all users", username)

	var err error
	if respdata.List, err = store.ListFull(); err != nil {
		respdata.Error = err.Error()
		sendWebResponse(w, http.StatusBadRequest, respdata)
		return
	}
	sendWebResponse(w, http.StatusOK, respdata)
}

func sendWebResponse(w http.ResponseWriter, status int, respdata interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.Encode(respdata)
}

type webHandler struct {
	store    *StoreChan
	sessions *webSessionFactory
	H        func(*StoreChan, *webSessionFactory, http.ResponseWriter, *http.Request)
}

func (self webHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.H(self.store, self.sessions, w, r)
}

// This is from golang http package - why is this not exported?
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func runWepApi(listener *net.TCPListener, store *StoreChan, staticDir string) (err error) {
	var sessions *webSessionFactory
	if sessions, err = NewWebSessionFactory(600 * time.Second); err != nil { // TODO: hardcoded value
		return err
	}

	http.Handle("/api/authenticate", webHandler{store, sessions, handleWebAuthenticate})
	http.Handle("/api/add", webHandler{store, sessions, handleWebAdd})
	http.Handle("/api/remove", webHandler{store, sessions, handleWebRemove})
	http.Handle("/api/update", webHandler{store, sessions, handleWebUpdate})
	http.Handle("/api/set-admin", webHandler{store, sessions, handleWebSetAdmin})
	http.Handle("/api/list", webHandler{store, sessions, handleWebList})
	http.Handle("/api/list-full", webHandler{store, sessions, handleWebListFull})

	http.Handle("/admin/", http.StripPrefix("/admin/", http.FileServer(http.Dir(staticDir))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/admin/", http.StatusTemporaryRedirect)
	})

	wl.Printf("web-api: listening on '%s'", listener.Addr())
	server := &http.Server{ReadTimeout: 60 * time.Second, WriteTimeout: 60 * time.Second}
	return server.Serve(tcpKeepAliveListener{listener})
}

func runWebApiAddr(addr string, store *StoreChan, staticDir string) (err error) {
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return runWepApi(ln.(*net.TCPListener), store, staticDir)
}

func runWebApiListener(listener *net.TCPListener, store *StoreChan, staticDir string) (err error) {
	return runWepApi(listener, store, staticDir)
}
