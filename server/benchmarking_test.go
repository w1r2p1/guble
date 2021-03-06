package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/smancke/guble/client"
	"github.com/smancke/guble/protocol"
	"github.com/smancke/guble/testutil"
)

type testgroup struct {
	t                   *testing.T
	groupID             int
	addr                string
	done                chan bool
	messagesToSend      int
	consumer, publisher client.Client
	topic               string
}

func newTestgroup(t *testing.T, groupID int, addr string, messagesToSend int) *testgroup {
	return &testgroup{
		t:              t,
		groupID:        groupID,
		addr:           addr,
		done:           make(chan bool),
		messagesToSend: messagesToSend,
	}
}

func TestThroughput(t *testing.T) {
	// TODO: We disabled this test because the receiver implementation of fetching messages
	// should be reimplemented according to the new message store
	testutil.SkipIfDisabled(t)
	testutil.SkipIfShort(t)
	defer testutil.ResetDefaultRegistryHealthCheck()

	dir, _ := ioutil.TempDir("", "guble_benchmarking_test")

	*Config.HttpListen = "localhost:0"
	*Config.KVS = "memory"
	*Config.MS = "file"
	*Config.StoragePath = dir

	service := StartService()

	testgroupCount := 4
	messagesPerGroup := 100
	log.Printf("init the %v testgroups", testgroupCount)
	testgroups := make([]*testgroup, testgroupCount, testgroupCount)
	for i := range testgroups {
		testgroups[i] = newTestgroup(t, i, service.WebServer().GetAddr(), messagesPerGroup)
	}

	// init test
	log.Print("init the testgroups")
	for i := range testgroups {
		testgroups[i].Init()
	}

	defer func() {
		// cleanup tests
		log.Print("cleanup the testgroups")
		for i := range testgroups {
			testgroups[i].Clean()
		}

		service.Stop()

		os.RemoveAll(dir)
	}()

	// start test
	log.Print("start the testgroups")
	start := time.Now()
	for i := range testgroups {
		go testgroups[i].Start()
	}

	log.Print("wait for finishing")
	for i, test := range testgroups {
		select {
		case successFlag := <-test.done:
			if !successFlag {
				t.Logf("testgroup %v returned with error", i)
				t.FailNow()
				return
			}
		case <-time.After(time.Second * 20):
			t.Log("timeout. testgroups not ready before timeout")
			t.Fail()
			return
		}
	}

	end := time.Now()
	totalMessages := testgroupCount * messagesPerGroup
	throughput := float64(totalMessages) / end.Sub(start).Seconds()
	log.Printf("finished! Throughput: %v/sec (%v message in %v)", int(throughput), totalMessages, end.Sub(start))

	time.Sleep(time.Second * 1)
}

func (tg *testgroup) Init() {
	tg.topic = fmt.Sprintf("/%v-foo", tg.groupID)
	var err error
	location := "ws://" + tg.addr + "/stream/user/xy"
	//location := "ws://gathermon.mancke.net:8080/stream/"
	//location := "ws://127.0.0.1:8080/stream/"
	tg.consumer, err = client.Open(location, "http://localhost/", 10, false)
	if err != nil {
		panic(err)
	}
	tg.publisher, err = client.Open(location, "http://localhost/", 10, false)
	if err != nil {
		panic(err)
	}

	tg.expectStatusMessage(protocol.SUCCESS_CONNECTED, "You are connected to the server.")

	tg.consumer.Subscribe(tg.topic)
	time.Sleep(time.Millisecond * 1)
	//test.expectStatusMessage(protocol.SUCCESS_SUBSCRIBED_TO, test.topic)
}

func (tg *testgroup) expectStatusMessage(name string, arg string) {
	select {
	case notify := <-tg.consumer.StatusMessages():
		assert.Equal(tg.t, name, notify.Name)
		assert.Equal(tg.t, arg, notify.Arg)
	case <-time.After(time.Second * 1):
		tg.t.Logf("[%v] no notification of type %s until timeout", tg.groupID, name)
		tg.done <- false
		tg.t.Fail()
		return
	}
}

func (tg *testgroup) Start() {
	go func() {
		for i := 0; i < tg.messagesToSend; i++ {
			body := fmt.Sprintf("Hallo-%d", i)
			tg.publisher.Send(tg.topic, body, "")
		}
	}()

	for i := 0; i < tg.messagesToSend; i++ {
		body := fmt.Sprintf("Hallo-%d", i)

		select {
		case msg := <-tg.consumer.Messages():
			assert.Equal(tg.t, tg.topic, string(msg.Path))
			if !assert.Equal(tg.t, body, msg.BodyAsString()) {
				tg.t.FailNow()
				tg.done <- false
			}
		case msg := <-tg.consumer.Errors():
			tg.t.Logf("[%v] received error: %v", tg.groupID, msg)
			tg.done <- false
			tg.t.Fail()
			return
		case <-time.After(time.Second * 5):
			tg.t.Logf("[%v] no message received until timeout, expected message %v", tg.groupID, i)
			tg.done <- false
			tg.t.Fail()
			return
		}
	}
	tg.done <- true
}

func (tg *testgroup) Clean() {
	tg.consumer.Close()
	tg.publisher.Close()
}
