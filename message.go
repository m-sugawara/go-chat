package main

import "time"

// message shows ONE message
type message struct {
	Name string
	Message string
	When time.Time
	AvatarURL string
}
