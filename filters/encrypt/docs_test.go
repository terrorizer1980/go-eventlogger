package encrypt_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/eventlogger"
	"github.com/hashicorp/eventlogger/filters/encrypt"
	"github.com/hashicorp/eventlogger/sinks/writer"
	wrapping "github.com/hashicorp/go-kms-wrapping"
	"github.com/hashicorp/go-kms-wrapping/wrappers/aead"
)

func ExampleFilter() {
	then := time.Date(
		2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	// Create a broker
	b := eventlogger.NewBroker()

	b.StopTimeAt(then) // setting this so the output timestamps are predictable for testing.

	wrapper := exampleWrapper()

	// A gated.Filter for events
	f := &encrypt.Filter{
		Wrapper:  wrapper,
		HmacSalt: []byte("salt"),
		HmacInfo: []byte("info"),
	}
	// Marshal to JSON
	jsonFmt := &eventlogger.JSONFormatter{}

	// Send the output to stdout
	stdoutSink := &writer.Sink{
		Writer: os.Stdout,
	}

	// Register the nodes with the broker
	nodes := []eventlogger.Node{f, jsonFmt, stdoutSink}
	nodeIDs := make([]eventlogger.NodeID, len(nodes))
	for i, node := range nodes {
		id := eventlogger.NodeID(fmt.Sprintf("node-%d", i))
		err := b.RegisterNode(id, node)
		if err != nil {
			// handle error
		}
		nodeIDs[i] = id
	}

	et := eventlogger.EventType("test-event")
	// Register a pipeline for our event type
	err := b.RegisterPipeline(eventlogger.Pipeline{
		EventType:  et,
		PipelineID: "encrypt-filter-pipeline",
		NodeIDs:    nodeIDs,
	})
	if err != nil {
		panic(err)
	}

	p := &struct {
		NoClassification string
		Public           string `classification:"public"`
		Sensitive        string `classification:"sensitive,redact"`
		Secret           string `classification:"secret,redact"`
	}{
		NoClassification: "no classification",
		Public:           "public",
		Sensitive:        "sensitive",
		Secret:           "secret",
	}

	ctx := context.Background()

	if _, err := b.Send(ctx, et, p); err != nil {
		panic(err)
	}

	// Output:
	// {"created_at":"2009-11-17T20:34:58.651387237Z","event_type":"test-event","payload":{"NoClassification":"\u003cREDACTED\u003e","Public":"public","Sensitive":"\u003cREDACTED\u003e","Secret":"\u003cREDACTED\u003e"}}
}

// exampleWrapper initializes an AEAD wrapping.Wrapper for examples
func exampleWrapper() wrapping.Wrapper {
	rootKey := make([]byte, 32)
	n, err := rand.Read(rootKey)
	if err != nil {
		panic(err)
	}
	if n != 32 {
		panic("unable to read 32 bytes from rand")
	}
	root := aead.NewWrapper(nil)
	_, err = root.SetConfig(map[string]string{
		"key_id": base64.StdEncoding.EncodeToString(rootKey),
	})
	if err != nil {
		panic(err)
	}
	err = root.SetAESGCMKeyBytes(rootKey)
	if err != nil {
		panic(err)
	}
	return root
}
