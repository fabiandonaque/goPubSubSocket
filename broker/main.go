package main

import (
	pss "github.com/fabsrobotics/pubsubsocket"
)

func main(){
	_,err := pss.StartBroker("fabs",1024,100)
	if err != nil { panic(err) }
	for {
		continue
	}
}