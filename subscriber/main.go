package main

import (
	"log"

	pss "github.com/fabsrobotics/pubsubsocket"
)

func main(){
	conn,err := pss.Dial("fabs",1024,true,100)
	if err != nil { panic(err) }
	dataChannel := make(chan []byte)
	errChannel := make(chan error)
	go conn.Listen(dataChannel,errChannel)
	for{
		select{
		case d := <-dataChannel:
			log.Println(string(d))
		case err := <- errChannel:
			log.Println(err)
			return
		default:
			continue
		}
	}
}