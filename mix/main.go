package main

import (
	"fmt"
	"log"
	"time"

	pss "github.com/fabsrobotics/pubsubsocket"
)

func main(){
	conn,err := pss.Dial("fabs",1024,true,100)
	if err != nil { panic(err) }
	dataChannel := make(chan []byte)
	errChannel := make(chan error)
	go conn.Listen(dataChannel,errChannel)
	counter := 0
	for {
		text := fmt.Sprintf("app2, counter: %v",counter)
		log.Println("Writing -> ",text)
		err = conn.Write([]byte(text))
		if err != nil { panic(err) }
		time.Sleep(1*time.Second)
		counter++
		select{
		case d := <-dataChannel:
			log.Println(string(d))
		case err := <-errChannel:
			log.Println(err)
			return
		default:
			continue
		}
	}
}