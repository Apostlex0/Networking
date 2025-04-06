package main

import (
	"fmt"
	"log"
	"net"
)

type Message struct {
	from string
	payload []byte
}
type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
	msgch chan Message
}
func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		msgch: make(chan Message, 10), // buffered channel with a capacity of 10
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln

	fmt.Println("Server started on", s.listenAddr)
	go s.AcceptLoop()

	<-s.quitch
	close(s.msgch)
	return nil
}

func (s *Server) AcceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Println("New connection to the remote client", conn.RemoteAddr())
		go s.ReadLoop(conn)
	}
}

func (s *Server) ReadLoop(conn net.Conn) {
	defer conn.Close()
	buff := make([]byte, 2048)
	for {
		n, err := conn.Read(buff)
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			continue
		}
		s.msgch <- Message{
			from : conn.RemoteAddr().String(),
			payload : buff[:n],
		}

		conn.Write([]byte("thank you for your message!"))
	}
}

func main() {
	server := NewServer(":8080")
	go func(){
		for msg := range server.msgch {
			fmt.Printf("Received message from connection (%s):\n%s", msg.from, string(msg.payload))
		}
		fmt.Println("Done reading all messages from the channel")
		server.quitch <- struct{}{}
	}()
	log.Fatal(server.Start())
}
