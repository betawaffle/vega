package vega

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestConsulNode(t *testing.T) {
	dir, err := ioutil.TempDir("", "mailbox")
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(dir)

	cn1, err := NewConsulClusterNode(
		&ConsulNodeConfig{
			AdvertiseAddr: "127.0.0.1",
			ListenPort:    8899,
			DataPath:      dir})

	if err != nil {
		panic(err)
	}

	defer cn1.Close()
	go cn1.Accept()

	dir2, err := ioutil.TempDir("", "mailbox")
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(dir2)

	cn2, err := NewConsulClusterNode(
		&ConsulNodeConfig{
			AdvertiseAddr: "127.0.0.1",
			ListenPort:    9900,
			DataPath:      dir2})

	if err != nil {
		panic(err)
	}

	defer cn2.Close()
	go cn2.Accept()

	cn1.Declare("a")

	// propagation delay
	time.Sleep(1000 * time.Millisecond)

	msg := Msg([]byte("hello"))

	debugf("pushing...\n")
	err = cn2.Push("a", msg)
	if err != nil {
		panic(err)
	}

	debugf("polling...\n")
	got, err := cn1.Poll("a")
	if err != nil {
		panic(err)
	}

	if got == nil || !got.Message.Equal(msg) {
		t.Fatal("didn't get the message")
	}
}

func TestConsulNodeRedeclaresOnStart(t *testing.T) {
	dir, err := ioutil.TempDir("", "mailbox")
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(dir)

	cn1, err := NewConsulClusterNode(
		&ConsulNodeConfig{
			AdvertiseAddr: "127.0.0.1",
			ListenPort:    8899,
			DataPath:      dir})

	if err != nil {
		panic(err)
	}

	// defer cn1.Close()
	go cn1.Accept()

	cl, err := NewClient(":8899")

	cn1.Declare("a")

	// propagation delay
	time.Sleep(1000 * time.Millisecond)

	msg := Msg([]byte("hello"))

	debugf("pushing...\n")
	err = cl.Push("a", msg)
	if err != nil {
		panic(err)
	}

	cl.Close()
	cn1.Cleanup()
	cn1.Close()

	cn1, err = NewConsulClusterNode(
		&ConsulNodeConfig{
			AdvertiseAddr: "127.0.0.1",
			ListenPort:    8899,
			DataPath:      dir})

	if err != nil {
		panic(err)
	}

	defer cn1.Close()
	go cn1.Accept()

	cl, _ = NewClient(":8899")

	err = cl.Push("a", msg)
	if err != nil {
		t.Fatal("routes were not readded")
	}
}
