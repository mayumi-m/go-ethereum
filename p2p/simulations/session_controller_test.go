package simulations

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
	"encoding/json"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/p2p/adapters"
)

/***
 * \todo rewrite this with a scripting engine to do http protocol xchanges more easily
 */
const (
	domain = "http://localhost"
	port   = "8888"
)

var quitc chan bool
var controller *ResourceController

func init() {
	glog.SetV(6)
	glog.SetToStderr(true)
	controller, quitc = NewSessionController()
	StartRestApiServer(port, controller)
}

func url(port, path string) string {
	return fmt.Sprintf("%v:%v/%v", domain, port, path)
}

func TestDelete(t *testing.T) {
	req, err := http.NewRequest("DELETE", url(port, ""), nil)
	if err != nil {
		t.Fatalf("unexpected error")
	}
	var resp *http.Response
	go func() {
		r, err := (&http.Client{}).Do(req)
		if err != nil {
			t.Fatalf("unexpected error")
		}
		resp = r
	}()
	timeout := time.NewTimer(1000 * time.Millisecond)
	select {
	case <-quitc:
	case <-timeout.C:
		t.Fatalf("timed out: controller did not quit, response: %v", resp)
	}
}

func TestCreate(t *testing.T) {
	s, err := json.Marshal(&struct{Id string}{Id: "testnetwork"})
	req, err := http.NewRequest("POST", domain + ":" + port, bytes.NewReader(s))
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatalf("unexpected error on http.Client request: %v", err)
	}
	req, err = http.NewRequest("POST", domain + ":" + port + "/testnetwork/debug/", nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}
	resp, err = (&http.Client{}).Do(req)
	if err != nil {
		t.Fatalf("unexpected error on http.Client request: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body: %v", err)
	}
	t.Logf("%s", body)
}

func TestNodes(t *testing.T) {
	networkname := "testnetworkfornodes"
	
	s, err := json.Marshal(&struct{Id string}{Id: networkname})
	req, err := http.NewRequest("POST", domain + ":" + port, bytes.NewReader(s))
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatalf("unexpected error on http.Client request: %v", err)
	}
	for i := 0; i < 3; i++ {
		req, err = http.NewRequest("POST", domain + ":" + port + "/" + networkname + "/node/", nil)
		if err != nil {
			t.Fatalf("unexpected error creating request: %v", err)
		}
		resp, err = (&http.Client{}).Do(req)
		if err != nil {
			t.Fatalf("unexpected error on http.Client request: %v", err)
		}
		t.Logf("%s", resp)
	}
	/*
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body: %v", err)
	}
	
	t.Logf("%s", body)
	*/
}

func testResponse(t *testing.T, method, addr string, r io.ReadSeeker) []byte {

	req, err := http.NewRequest(method, addr, r)
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatalf("unexpected error on http.Client request: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body: %v", err)
	}
	return body

}

func TestUpdate(t *testing.T) {

	ids := testIDs()
	journal := testJournal(ids)

	conf := &NetworkConfig{
		Id: "0",
	}
	mc := NewNetworkController(conf, &event.TypeMux{}, journal)
	controller.SetResource(conf.Id, mc)
	exp := `{
  "add": [
    {
      "data": {
        "id": "aa7c",
        "up": true
      },
      "group": "nodes"
    },
    {
      "data": {
        "id": "f5ae",
        "up": true
      },
      "group": "nodes"
    }
  ],
  "remove": [],
  "message": []
}`
	resp := testResponse(t, "GET", url(port, "0"), bytes.NewReader([]byte("{}")))
	if string(resp) != exp {
		t.Fatalf("incorrect response body. got\n'%v', expected\n'%v'", string(resp), exp)
	}
}

func mockNewNodes(eventer *event.TypeMux, ids []*adapters.NodeId) {
	glog.V(6).Infof("mock starting")
	for _, id := range ids {
		glog.V(6).Infof("mock adding node %v", id)
		eventer.Post(&NodeEvent{
			Action: "up",
			Type:   "node",
			node:   &Node{Id: id, config: &NodeConfig{Id: id}},
		})
	}
}
