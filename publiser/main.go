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
	counter := 0
	for {
		text := fmt.Sprintf("app1, counter: %v",counter)
		log.Println("Writing -> ",text)
		err = conn.Write([]byte(text))
		if err != nil { panic(err) }
		time.Sleep(1*time.Second)
		counter++
	}
}