package pubsub

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"time"
)

type Broker struct{
	Size int64
	Name string
	Wait int64
	Socket net.Listener
	Connections []net.Conn
}

type Conn struct {
	Size int64
	Name string
	Wait int64
	Reconnect bool
	Conn net.Conn
}

func checkName(name string) error {
	// Check null
	if name == "" { return fmt.Errorf("name can't be null") }
	// Check inital character
	if name[0] == '-' || name[0] == '_' { return fmt.Errorf("name can't begin with '-' or '_' ") }
	// Check final character
	if name[len(name)-1] == '-' || name[len(name)-1] == '_' || name[len(name)-1] == '.' { return fmt.Errorf("name can't end with '-' or '_' or '.' ") }
	re := regexp.MustCompile(`^[a-zA-Z0-9\-\_\.]*$`)
	if !re.MatchString(name){ return fmt.Errorf("name has to be conformed by 'a' to 'z' or 'A' to 'Z' or '0' to '9' or '-' or '_' or '.' ")}
	return nil
}

func (b *Broker)handleConnections(){
	for {
		conn,err := b.Socket.Accept()
		if err != nil { continue }
		b.Connections = append(b.Connections, conn)
	}
}

func (b *Broker)removeConnection(conn net.Conn){
	var temp []net.Conn
	for i,c := range b.Connections{
		if c == conn {
			temp = append(b.Connections[:i],b.Connections[i+1:]...)
			break
		}
	}
	b.Connections = temp
}

func (b *Broker)handleBroadcasting(){
	for{
		// Read block
		buf := make([]byte,b.Size)
		var wconn net.Conn
		n := 0
		var err error
		for _,c := range b.Connections {
			err = c.SetReadDeadline(time.Now().Add(time.Duration(b.Wait)*time.Millisecond))
			if err != nil { continue }
			n,err = c.Read(buf)
			if err != nil { b.removeConnection(c); break }
			if n > 0 { wconn = c; break }
		}
		// Write block
		if n > 0 {
			for _,c := range b.Connections{
				if wconn != c {
					err = c.SetWriteDeadline(time.Now().Add(time.Duration(b.Wait)*time.Millisecond))
					if err != nil { continue }
					_,err := c.Write(buf[:n])
					if err != nil { b.removeConnection(c) }
				}
			}
		}
	}
}

func StartBroker(name string, size int64, wait int64) (*Broker, error){
	// Check name
	err := checkName(name)
	if err != nil { return nil,err }
	// Check size
	if size <= 0 { return nil,fmt.Errorf("size has to be greater than 0") }
	// Check wait
	if wait <= 0 { return nil,fmt.Errorf("wait has to be greater than 0 milliseconds ") }
	// Remove socket if it exists
	socketPath := "/tmp/"+name+".sock"
	os.Remove(socketPath)
	// Listen to socket
	socket,err := net.Listen("unix",socketPath)
	if err != nil { return nil,err }
	// Create a Broker
	broker := Broker{
		Name: name,
		Size: size,
		Wait: wait,
		Socket: socket,
	}
	// Set connection handler
	go broker.handleConnections()
	// Set broadcasting handler
	go broker.handleBroadcasting()

	return &broker,nil
}

func (b *Broker) Close() error {
	for _,c := range b.Connections {
		err := c.Close()
		if err != nil { return err }
	}
	err := b.Socket.Close()
	if err != nil { return err }
	b = nil
	return nil
}

func Dial(name string, size int64, reconnect bool, wait int64)(*Conn,error){
	// Check name
	err := checkName(name)
	if err != nil { return nil,err }
	// Check size
	if size <= 0 { return nil,fmt.Errorf("size has to be greater than 0") }
	// Check wait
	if wait <= 0 { return nil,fmt.Errorf("wait has to be greater than 0 milliseconds ") }
	// Dial to socket
	socket,err := net.Dial("unix","/tmp/"+name+".sock")
	if err != nil { return nil,err }
	// Create connection
	conn := Conn{
		Name: name,
		Size: size,
		Wait: wait,
		Reconnect: reconnect,
		Conn: socket,
	}
	return &conn,nil
}

func (c *Conn)reDial(){
	for {
		socket,err := net.Dial("unix","/temp/"+c.Name+".sock")
		if err != nil {
			time.Sleep(1*time.Second)
		} else {
			c.Conn = socket
			break
		}
	}
}

func (c *Conn)Listen(dataChannel chan []byte, errChannel chan error){
	for {
		buf := make([]byte,c.Size)
		err := c.Conn.SetReadDeadline(time.Now().Add(time.Duration(c.Wait)*time.Millisecond))
		if err != nil { errChannel<-err; break }
		n,err := c.Conn.Read(buf)
		if err != nil {
			if c.Reconnect {
				c.reDial()
			} else {
				errChannel<-err
				break
			}
		}
		if n>0{
			dataChannel<-buf[:n]
		}
	}
}

func (c *Conn) Write(data []byte) error {
	err := c.Conn.SetWriteDeadline(time.Now().Add(time.Duration(c.Wait)*time.Millisecond))
	if err != nil { return err }
	_,err = c.Conn.Write(data)
	if err != nil { 
		if c.Reconnect {
			c.reDial()
		} else {
			return err
		}
	}
	return nil
}

func (c *Conn) Close() error {
	err := c.Conn.Close()
	if err != nil { return err }
	c = nil
	return nil
}