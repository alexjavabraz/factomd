package state_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/state"
)

var _ = fmt.Println

func TestQueues(t *testing.T) {
	RegisterPrometheus()
	RegisterPrometheus()
	channel := make(chan interfaces.IMsg, 1000)
	general := GeneralMSGQueue(channel)
	inmsg := InMsgMSGQueue(channel)
	netOut := NetOutMsgQueue(channel)

	if !checkLensAndCap(channel, []interfaces.IQueue{general, inmsg, netOut}) {
		t.Error("Error: Lengths/Cap does not match")
	}

	c := 0
	for i := 0; i < 100; i++ {
		switch c {
		case 0:
			channel <- nil
		case 1:
			general.Enqueue(new(messages.DBStateMsg))
		case 2:
			inmsg.Enqueue(nil)
		}
		c++
		if c == 3 {
			c = 0
		}
		if !checkLensAndCap(channel, []interfaces.IQueue{general, inmsg, netOut}) {
			t.Error("Error: Lengths/Cap does not match")
		}

	}

	for i := 0; i < 100; i++ {
		switch c {
		case 0:
			<-channel
		case 1:
			general.Dequeue()
		case 2:
			inmsg.Dequeue()
		}
		c++
		if c == 3 {
			c = 0
		}
		if !checkLensAndCap(channel, []interfaces.IQueue{general, inmsg, netOut}) {
			t.Error("Error: Lengths/Cap does not match")
		}
	}

	if len(channel) != 0 {
		t.Errorf("Channel should be 0, found %d", len(channel))
	}

	// Check for blocking
	select {
	case <-channel:
	default:
	}
	go func() {
		time.Sleep(1100 * time.Millisecond)
		general.Enqueue(nil)
		inmsg.Enqueue(nil)
		netOut.Enqueue(nil)
	}()

	b := time.Now().Unix()
	general.BlockingDequeue()
	if time.Now().Unix()-b < 1 {
		t.Error("Did not properly block")
	}

	inmsg.BlockingDequeue()
	if time.Now().Unix()-b < 1 {
		t.Error("Did not properly block")
	}

	netOut.BlockingDequeue()
	if time.Now().Unix()-b < 1 {
		t.Error("Did not properly block")
	}

	// Test NonBlocking
	if v := general.Dequeue(); v != nil {
		t.Error("Should be nil")
	}
	if v := inmsg.Dequeue(); v != nil {
		t.Error("Should be nil")
	}
	if v := netOut.Dequeue(); v != nil {
		t.Error("Should be nil")
	}

	// Trip prometheus, unfortunately, we cannot actually check the values
	tripAllMessages(inmsg)
	tripAllMessages(general)
	tripAllMessages(netOut)

	if len(channel) != 0 {
		t.Errorf("Channel should be 0, found %d", len(channel))
	}
	if !checkLensAndCap(channel, []interfaces.IQueue{general, inmsg, netOut}) {
		t.Error("Error: Lengths/Cap does not match")
	}
}

func tripAllMessages(q interfaces.IQueue) {
	EnAndDeQueue(q, new(messages.EOM))
	EnAndDeQueue(q, new(messages.Ack))
	EnAndDeQueue(q, new(messages.AuditServerFault))
	EnAndDeQueue(q, new(messages.ServerFault))
	EnAndDeQueue(q, new(messages.FullServerFault))
	EnAndDeQueue(q, new(messages.CommitChainMsg))
	EnAndDeQueue(q, new(messages.CommitEntryMsg))
	EnAndDeQueue(q, new(messages.DirectoryBlockSignature))
	EnAndDeQueue(q, new(messages.EOMTimeout))
	EnAndDeQueue(q, new(messages.Heartbeat))
	EnAndDeQueue(q, new(messages.InvalidDirectoryBlock))
	EnAndDeQueue(q, new(messages.MissingMsg))
	EnAndDeQueue(q, new(messages.MissingMsgResponse))
	EnAndDeQueue(q, new(messages.MissingData))
	EnAndDeQueue(q, new(messages.RevealEntryMsg))
	EnAndDeQueue(q, new(messages.DBStateMsg))
	EnAndDeQueue(q, new(messages.DBStateMissing))
	EnAndDeQueue(q, new(messages.Bounce))
	EnAndDeQueue(q, new(messages.BounceReply))
	EnAndDeQueue(q, new(messages.SignatureTimeout))
	EnAndDeQueue(q, new(messages.FactoidTransaction))
	EnAndDeQueue(q, new(messages.DataResponse))
	EnAndDeQueue(q, new(messages.RequestBlock))

}

func EnAndDeQueue(q interfaces.IQueue, m interfaces.IMsg) {
	q.Enqueue(m)
	q.Dequeue()
}

func checkLensAndCap(channel chan interfaces.IMsg, qs []interfaces.IQueue) bool {
	for _, q := range qs {
		if len(channel) != q.Length() {
			return false
		}
		if cap(channel) != q.Cap() {
			return false
		}
	}
	return true
}
