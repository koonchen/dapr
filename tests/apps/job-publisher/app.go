// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation and Dapr Contributors.
// Licensed under the MIT License.
// ------------------------------------------------------------

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	daprPort       = 3500
	pubsubName     = "messagebus"
	pubsubTopic    = "pubsub-job-topic-http"
	message        = "message-from-job"
	publishRetries = 25
)

func stopSidecar() {
	log.Printf("Shutting down the sidecar at %s", fmt.Sprintf("http://localhost:%d/v1.0/shutdown", daprPort))
	r, err := http.Get(fmt.Sprintf("http://localhost:%d/v1.0/shutdown", daprPort))
	if r != nil {
		r.Body.Close()
	}
	if err != nil {
		log.Printf("Error stopping the sidecar %s", err)
	}
	log.Printf("Sidecar stopped")
}

func publishMessagesToPubsub() error {
	daprPubsubURL := fmt.Sprintf("http://localhost:%d/v1.0/publish/%s/%s", daprPort, pubsubName, pubsubTopic)
	jsonValue, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling %s to JSON", message)
	}
	log.Printf("Publishing to %s", daprPubsubURL)
	// nolint: gosec
	r, err := http.Post(daprPubsubURL, "application/json", bytes.NewBuffer(jsonValue))
	if r != nil {
		defer r.Body.Close()
	}
	if err != nil {
		log.Printf("Error publishing messages to pubsub: %+v", err)
	}
	return err
}

func main() {
	for retryCount := 0; retryCount < publishRetries; retryCount++ {
		err := publishMessagesToPubsub()
		if err != nil {
			log.Printf("Unable to publish, retrying.")
			time.Sleep(2 * time.Second)
		} else {
			stopSidecar()
			os.Exit(0)
		}
	}
	stopSidecar()
	os.Exit(1)
}
