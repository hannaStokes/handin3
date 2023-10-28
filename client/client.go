package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	//"strconv"
	"strings"

	// this has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.
	gRPC "github.com/hannaStokes/handin3/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Same principle as in client. Flags allows for user specific arguments/values
var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", "5400", "Tcp server")
var clientsTime int64 = 0

var server gRPC.ChittyChatClient //the server
var ServerConn *grpc.ClientConn  //the server connection

func main() {
	//parse flag/arguments
	flag.Parse()

	fmt.Println("--- CLIENT APP ---")

	//log to file instead of console
	//f := setLog()
	//defer f.Close()

	//connect to server and close the connection when program closes
	fmt.Println("--- join Server ---")
	ConnectToServer()
	defer ServerConn.Close()

	message := &gRPC.SubMessage{
		ClientName: *clientsName,
		Timestamp:  clientsTime,
	}
	go subscribe(server, message)

	//start the biding
	parseInput()
}

// connect to server
func ConnectToServer() {

	//dial options
	//the server is not using TLS, so we use insecure credentials
	//(should be fine for local testing but not in the real world)
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	//dial the server, with the flag "server", to get a connection to it
	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, *serverPort)
	conn, err := grpc.Dial(fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}
	//fmt.Printf("Error here")
	// makes a client from the server connection and saves the connection
	// and prints rather or not the connection was is READY
	server = gRPC.NewChittyChatClient(conn)
	ServerConn = conn
	log.Println("the connection is: ", conn.GetState().String())
}

func subscribe(client gRPC.ChittyChatClient, r *gRPC.SubMessage) error {
	clientsTime++
	stream, _ := client.Subscribe(context.Background(), r)
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		IncreaseLamport(res.Timestamp)
		log.Printf("\"%s\" at timestamp %d", res.Message, clientsTime)
	}
	return nil
}

func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Type the message you wish to send here")
	fmt.Println("--------------------")

	//Infinite loop to listen for clients input.
	for {
		fmt.Print("-> ")

		//Read input into var input and any errors into err
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input) //Trim input

		if !conReady(server) {
			log.Printf("Client %s: something was wrong with the connection to the server :(", *clientsName)
			continue
		}
		clientsTime++
		//Convert string to int64, return error if the int is larger than 32bit or not a number
		message := &gRPC.ChatMessage{
			ClientName: *clientsName,
			Timestamp:  clientsTime,
			Message:    fmt.Sprintf("received message \"%s\" from user %s", input, *clientsName),
		}
		//
		ack, err := server.Publish(context.Background(), message)
		//IncreaseLamport(ack.TimeStamp)
		ack.Timestamp++
		if err != nil {
			log.Printf("Client %s: no response from the server, attempting to reconnect", *clientsName)
			log.Println(err)
		}
	}
}

// Function which returns a true boolean if the connection to the server is ready, and false if it's not.
func conReady(s gRPC.ChittyChatClient) bool {
	return ServerConn.GetState().String() == "READY"
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}

func IncreaseLamport(timestamp int64) {
	if clientsTime < timestamp {
		clientsTime = timestamp
	}
	clientsTime++
}
