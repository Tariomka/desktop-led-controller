package data

import "fmt"

// MessageReceiver is a function for handling incoming messaages, i.e. this function will be invoked
// each time a message is sent. Simplest example would be sending the message to a channel:
//
//	func(message any) {
//		myStruct.myChannel <- message
//	}
type MessageReceiver func(message any)

type Messenger struct {
	channels map[string]MessageReceiver
}

func NewMessanger() *Messenger {
	return &Messenger{
		channels: make(map[string]MessageReceiver),
	}
}

func (this *Messenger) RegisterReceiver(key string, receiver MessageReceiver) {
	this.channels[key] = receiver
}

func (this *Messenger) Send(key string, message any) {
	receiver, ok := this.channels[key]
	if !ok {
		// TODO: use slog?
		fmt.Printf("[MESSENGER] Receiver with key '%s' is not registered.\n", key)
		return
	}

	receiver(message)
}
