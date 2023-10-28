package main

import (
	"context"
	"flag"
	"fmt"

	// "io"
	"log"
	"net"
	"os"
	"sync"

	// this has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.
	gRPC "github.com/hannaStokes/handin3/proto"

	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedChittyChatServer        // You need this line if you have a server. "ChittyChat" should be the equivalent name of your service
	name                               string // Not required but useful if you want to name your server
	port                               string // Not required but useful if your server needs to know what port it's listening to

	channelList []chan gRPC.ChatMessage

	currentTime int64      // value that clients can increment.
	mutex       sync.Mutex // used to lock the server to avoid race conditions.
}

// flags are used to get arguments from the terminal. Flags take a value, a default value and a description of the flag.
// to use a flag then just add it as an argument when running the program.
var serverName = flag.String("name", "default", "Senders name") // set with "-name <name>" in terminal
var port = flag.String("port", "5400", "Server port")           // set with "-port <port>" in terminal

func main() {
	// f := setLog() //uncomment this line to log to a log.txt file instead of the console
	// defer f.Close()

	// This parses the flags and sets the correct/given corresponding values.
	flag.Parse()
	fmt.Println(".:server is starting:.")

	// launch the server
	launchServer()

	// code here is unreachable because launchServer occupies the current thread.
}

func launchServer() {
	log.Printf("Server %s: Attempts to create listener on port %s\n", *serverName, *port)

	// Create listener tcp on given port or default port 5400
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	// makes gRPC server using the options
	// you can add options here if you want or remove the options part entirely
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	// makes a new server instance using the name and port from the flags.
	server := &Server{
		name:        *serverName,
		port:        *port,
		currentTime: 0,
		channelList: make([]chan gRPC.ChatMessage, 0),
	}

	gRPC.RegisterChittyChatServer(grpcServer, server) //Registers the server to the gRPC server.

	log.Printf("Server %s: Listening at %v\n", *serverName, list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
	// code here is unreachable because grpcServer.Serve occupies the current thread.
}

// The method format can be found in the pb.go file. If the format is wrong, the server type will give an error.
func (s *Server) Subscribe(in *gRPC.SubMessage, stream gRPC.ChittyChat_SubscribeServer) error {

	name := in.ClientName
	//s.mutex.Lock() ??
	//lamport time missing
	IncreaseLamport(s, in.Timestamp)

	channel := make(chan gRPC.ChatMessage)
	s.channelList = append(s.channelList, channel)

	go recv(s, channel, stream)

	msg := fmt.Sprintf("User %s subscribed", name)
	broadcast(s, gRPC.ChatMessage{ClientName: name, Timestamp: s.currentTime, Message: msg})
	<-stream.Context().Done()
	//remove stream from s.channelList and send out "user logged off" message to remaining channels
	//Locate channel in channelList
	for i, c := range s.channelList {
		if c == channel {
			s.channelList[i] = nil
			if i == 0 {
				s.channelList = s.channelList[1:]
			} else if i == len(s.channelList) {
				s.channelList = s.channelList[:i]
			} else {
				s.channelList = append(s.channelList[:i], s.channelList[i+1:]...)
			}
			break
		}
	}

	lvmsg := fmt.Sprintf("User %s left the server", name)
	broadcast(s, gRPC.ChatMessage{ClientName: name, Timestamp: s.currentTime, Message: lvmsg})
	return nil
}

func IncreaseLamport(s *Server, timestamp int64) {
	if s.currentTime < timestamp {
		s.currentTime = timestamp
	}
	s.currentTime++
}

func broadcast(s *Server, message gRPC.ChatMessage) {
	s.currentTime++ //receive and send are separate events
	for _, channel := range s.channelList {
		channel <- message
	}
}

func recv(s *Server, channel chan gRPC.ChatMessage, stream gRPC.ChittyChat_SubscribeServer) {
	for {
		var recv = <-channel
		//IncreaseLamport(s, recv.Timestamp) already handled in Publish
		stream.Send(&gRPC.ChatMessage{ClientName: recv.ClientName, Timestamp: s.currentTime, Message: recv.Message})
		//defer s.mutex.Unlock()??
	}
}

func (s *Server) Publish(ctx context.Context, ChatMessage *gRPC.ChatMessage) (*gRPC.ChatAccept, error) {

	IncreaseLamport(s, ChatMessage.Timestamp)
	broadcast(s, *ChatMessage)
	name := s.name
	//IncreaseLamport(s,ChatMessage.Timestamp)
	return &gRPC.ChatAccept{ServerName: name, Timestamp: s.currentTime}, nil
}

// Get preferred outbound ip of this machine
// Usefull if you have to know which ip you should dial, in a client running on an other computer
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate("log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
