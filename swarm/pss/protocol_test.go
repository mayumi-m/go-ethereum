package pss

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
)

type protoCtrl struct {
	C        chan bool
	protocol *Protocol
	run      func(*p2p.Peer, p2p.MsgReadWriter) error
}

// simple ping pong protocol test for the pss devp2p emulation
func TestProtocol(t *testing.T) {
	t.Run("32", testProtocol)
	t.Run("8", testProtocol)
	t.Run("0", testProtocol)
}

func testProtocol(t *testing.T) {

	// address hint size
	var addrsize int64
	var err error
	paramstring := strings.Split(t.Name(), "/")
	addrsize, _ = strconv.ParseInt(paramstring[1], 10, 0)
	log.Info("protocol test", "addrsize", addrsize)

	clients, err := setupNetwork(2)
	if err != nil {
		t.Fatal(err)
	}

	var loaddr []byte
	err = clients[0].Call(&loaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 1 baseaddr fail: %v", err)
	}
	loaddr = loaddr[:addrsize]
	var roaddr []byte
	err = clients[1].Call(&roaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 2 baseaddr fail: %v", err)
	}
	roaddr = roaddr[:addrsize]
	lnodeinfo := &p2p.NodeInfo{}
	err = clients[0].Call(&lnodeinfo, "admin_nodeInfo")
	if err != nil {
		t.Fatalf("rpc nodeinfo node 11 fail: %v", err)
	}

	lpubkey := make([]byte, 32)
	err = clients[0].Call(&lpubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 1 pubkey fail: %v", err)
	}
	rpubkey := make([]byte, 32)
	err = clients[1].Call(&rpubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 2 pubkey fail: %v", err)
	}

	time.Sleep(time.Millisecond * 500) // replace with hive healthy code

	lmsgC := make(chan APIMsg)
	lctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	lsub, err := clients[0].Subscribe(lctx, "pss", lmsgC, "receive", PingTopic)
	log.Trace("lsub", "id", lsub)
	defer lsub.Unsubscribe()
	rmsgC := make(chan APIMsg)
	rctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	rsub, err := clients[1].Subscribe(rctx, "pss", rmsgC, "receive", PingTopic)
	log.Trace("rsub", "id", rsub)
	defer rsub.Unsubscribe()

	// set reciprocal public keys
	err = clients[0].Call(nil, "pss_setPeerPublicKey", rpubkey, PingTopic, roaddr)
	if err != nil {
		t.Fatal(err)
	}
	err = clients[1].Call(nil, "pss_setPeerPublicKey", lpubkey, PingTopic, loaddr)
	if err != nil {
		t.Fatal(err)
	}

	// add right peer's public key as protocol peer on left
	nid, _ := discover.HexID("0x00") // this hack is needed to satisfy the p2p method
	p := p2p.NewPeer(nid, fmt.Sprintf("%x", loaddr), []p2p.Cap{})
	pssprotocols[lnodeinfo.ID].protocol.AddPeer(p, pssprotocols[lnodeinfo.ID].run, PingTopic, true, common.ToHex(rpubkey))

	// sends ping asym, checks delivery
	pssprotocols[lnodeinfo.ID].C <- false
	select {
	case <-lmsgC:
		log.Debug("lnode ok")
	case cerr := <-lctx.Done():
		t.Fatalf("test message timed out: %v", cerr)
	}
	select {
	case <-rmsgC:
		log.Debug("rnode ok")
	case cerr := <-lctx.Done():
		t.Fatalf("test message timed out: %v", cerr)
	}

	// sends ping asym, checks delivery
	pssprotocols[lnodeinfo.ID].C <- false
	select {
	case <-lmsgC:
		log.Debug("lnode ok")
	case cerr := <-lctx.Done():
		t.Fatalf("test message timed out: %v", cerr)
	}
	select {
	case <-rmsgC:
		log.Debug("rnode ok")
	case cerr := <-lctx.Done():
		t.Fatalf("test message timed out: %v", cerr)
	}

}
