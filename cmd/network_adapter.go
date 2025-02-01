package main

// NetworkAdapter defines how nodes send messages
type NetworkAdapter interface {
	SendMessage(targetID string, message string) error
}
