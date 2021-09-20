/*
Released under MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM GmbH. All Rights Reserved.

Provides simple Kamailio JSON-RPC over HTTP communication.
*/

package kamjsonrpc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

const (
	OK = "OK"
)

type KamJsonRpcRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      uint64        `json:"id"`
}

// {"jsonrpc":"2.0","error":{"code":-32000,"message":"Execution Error"},"id":0}
type KamError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type KamJsonRpcResponse struct {
	Jsonrpc string           `json:"jsonrpc"`
	Id      uint64           `json:"id"`
	Result  *json.RawMessage `json:"result"`
	Error   *KamError        `json:"error"`
}

func NewKamailioJsonRpc(url string, skipTlsVerify bool) (*KamailioJsonRpc, error) {
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTlsVerify}}}
	return &KamailioJsonRpc{url: url, client: client, mutex: new(sync.Mutex)}, nil
}

type KamailioJsonRpc struct {
	url    string
	client *http.Client
	id     uint64
	mutex  *sync.Mutex
}

// Generic function to remotely call a method and pass the parameters
func (self *KamailioJsonRpc) Call(serviceMethod string, args interface{}, reply *json.RawMessage) error {
	self.mutex.Lock()
	reqId := self.id
	self.id += 1
	self.mutex.Unlock()
	req := &KamJsonRpcRequest{Jsonrpc: "2.0", Method: serviceMethod, Id: reqId}
	if argSlice, isSlice := args.([]string); isSlice {
		req.Params = make([]interface{}, len(argSlice))
		for idx, val := range argSlice {
			req.Params[idx] = val
		}
	} else {
		req.Params = []interface{}{args}
	}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	resp, err := self.client.Post(self.url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var kamResponse KamJsonRpcResponse
	if err = json.Unmarshal(respBody, &kamResponse); err != nil {
		return err
	}
	if kamResponse.Error != nil {
		return errors.New(kamResponse.Error.Message)
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("Unexpected status code received: %d", resp.StatusCode)
	}
	if kamResponse.Id != reqId {
		return fmt.Errorf("Unsynchronized request, had: %d, received: %d", reqId, kamResponse.Id)
	}
	*reply = *kamResponse.Result
	return nil
}

// Add inidividual methods over the generic one

func (self *KamailioJsonRpc) CoreEcho(params []string, reply *[]string) error {
	var rplRaw json.RawMessage
	if err := self.Call("core.echo", params, &rplRaw); err != nil {
		return err
	}
	return json.Unmarshal(rplRaw, reply)
}

func (self *KamailioJsonRpc) UacRegEnable(params []string, reply *string) error {
	var regRaw json.RawMessage
	if err := self.Call("uac.reg_enable", params, &regRaw); err != nil {
		return err
	}
	*reply = OK
	return nil
}

func (self *KamailioJsonRpc) UacRegDisable(params []string, reply *string) error {
	var regRaw json.RawMessage
	if err := self.Call("uac.reg_disable", params, &regRaw); err != nil {
		return err
	}
	*reply = OK
	return nil
}

func (self *KamailioJsonRpc) UacRegReload(params []string, reply *string) error {
	var regRaw json.RawMessage
	if err := self.Call("uac.reg_reload", params, &regRaw); err != nil {
		return err
	}
	*reply = OK
	return nil
}

func (self *KamailioJsonRpc) UacRegRefresh(params []string, reply *string) error {
	var regRaw json.RawMessage
	if err := self.Call("uac.reg_refresh", params, &regRaw); err != nil {
		return err
	}
	*reply = OK
	return nil
}

func (self *KamailioJsonRpc) UacRegInfo(params []string, reply *RegistrationInfo) error {
	var regRaw json.RawMessage
	if err := self.Call("uac.reg_info", params, &regRaw); err != nil {
		return err
	}
	return json.Unmarshal(regRaw, reply)
}

func (self *KamailioJsonRpc) DomainReload(params []string, reply *string) error {
	var response json.RawMessage
	if err := self.Call("domain.reload", params, &response); err != nil {
		return err
	}
	*reply = OK
	return nil
}

func (self *KamailioJsonRpc) UsrlocDump(params []string, reply *ULDump) error {
	var ulRaw json.RawMessage
	if err := self.Call("ul.dump", params, &ulRaw); err != nil {
		return err
	}
	return json.Unmarshal(ulRaw, reply)
}

//func (self *KamailioJsonRpc) UsrlocLookup(params []string, reply *json.RawMessage) error {
func (self *KamailioJsonRpc) UsrlocLookup(params []string, reply *ULSingle) error {
	var ulRaw json.RawMessage
	if err := self.Call("ul.lookup", params, &ulRaw); err != nil {
		return err
	}
	return json.Unmarshal(ulRaw, reply)
}

type RegistrationInfo struct {
	LocalUuid      string          `json:"l_uuid"`
	LocalUsername  string          `json:"l_username"`
	LocalDomain    string          `json:"l_domain"`
	RemoteUsername string          `json:"r_username"`
	RemoteDomain   string          `json:"r_domain"`
	Realm          string          `json:"realm"`
	AuthUsername   string          `json:"auth_username"`
	AuthPassword   string          `json:"auth_password"`
	AuthProxy      string          `json:"auth_proxy"`
	Expires        json.RawMessage `json:"expires"`
	Flags          int64           `json:"flags"`
	DiffExpires    int64           `json:"diff_expires"`
	TimerExpires   int64           `json:"timer_expires"`
}

type ULContact struct {
	Contact struct {
		Address       string          `json:"Address"`
		Expires       json.RawMessage `json:"Expires"`
		Q             int64           `json:"Q"`
		CallID        string          `json:"Call-ID"`
		Cseq          int64           `json:"CSeq"`
		UserAgent     string          `json:"User-Agent"`
		Received      string          `json:"Received"`
		Path          string          `json:"Path"`
		State         string          `json:"State"`
		Flags         int             `json:"Falgs"`
		CFlags        int             `json:"CFlags"`
		Socket        string          `json:"Socket"`
		Methods       int64           `json:"Methods"`
		Ruid          string          `json:"Ruid"`
		Instance      string          `json:"Instance"`
		RegID         int64           `json:"Reg-Id"`
		ServerID      int             `json:"Server-Id"`
		TcpconnID     int64           `json:"Tcpconn-Id"`
		Keepalive     int             `json:"Keepalive"`
		LastKeepalive int64           `json:"Last-Keepalive"`
		KaRoundtrip   int             `json:"KA-Roundtrip"`
		LastModified  int64           `json:"Last-Modified"`
	} `json:"Contact"`
}

type ULSingle struct {
	Aor      string      `json:"AoR"`
	Contacts []ULContact `json:"Contacts"`
}

type ULAor struct {
	Info struct {
		Aor    string `json:"AoR"`
		HashID int64  `json:"HashID"`
		//HashID   uint64      `json:"HashID"`
		Contacts []ULContact `json:"Contacts"`
	} `json:"Info"`
}

type ULDomain struct {
	Domain struct {
		Domain string  `json:"Domain"`
		Size   int     `json:"Size"`
		Aors   []ULAor `json:"AoRs"`
		Stats  struct {
			Records  int64 `json:"Records"`
			MaxSlots int64 `json:"Max-Slots"`
		} `json:"Stats"`
	} `json:"Domain"`
}

type ULDump struct {
	Domains []ULDomain `json:"Domains"`
}
