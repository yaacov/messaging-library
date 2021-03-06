/*
Copyright (c) 2018 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package stomp

import (
	"encoding/json"
	"fmt"

	"github.com/go-stomp/stomp"

	"github.com/container-mgmt/messaging-library/pkg/client"
)

// Subscribe creates a subscription on the messaging server.
// The subscription has a destination, and messages sent to that destination
// will be received by this subscription.
//
// Once a message or an error is received, the callback function will be trigered.
func (c *Connection) Subscribe(destination string, callback client.SubscriptionCallback) (err error) {
	var subscription *stomp.Subscription
	var data client.MessageData

	// Check if we already subscibe to this destination,
	// We do not allow for multiple subscriptions for one destination.
	if _, ok := c.subscriptions[destination]; ok {
		err = fmt.Errorf("Only one subscription per destination is allowed")
		return
	}

	// Receive messages:
	subscription, err = c.connection.Subscribe(destination, stomp.AckAuto)
	if err != nil {
		return
	}

	c.subscriptions[destination] = subscription

	// Wait for messages:
	go func() {
		for message := range subscription.C {
			// Try to unmarshal the byte array coming from the broker into a
			// message body of type map[string]interface{}
			err = json.Unmarshal(message.Body, &data)
			if err != nil {
				// Call the callback function with the json unmarshal error.
				callback(
					client.Message{
						Data: client.MessageData{"byteArray": message.Body},
						Err:  err},
					destination)
			}

			// Call the callback function.
			callback(
				client.Message{
					Data: data,
					Err:  message.Err},
				destination)
		}
	}()

	return
}

// Unsubscribe unsubscribes from a destination.
func (c *Connection) Unsubscribe(destination string) (err error) {
	// Check if we subscribe to this destination, o/w return an error.
	subscription, ok := c.subscriptions[destination]
	if ok == false {
		err = fmt.Errorf("Unsubscribe faild, no destination %s", destination)
		return
	}

	err = subscription.Unsubscribe()
	return
}
