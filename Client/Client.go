package main

import (
	"bufio"
	"context"
	"fmt"
	"grpcChatServer/chatserver"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
)

func main() {

	fmt.Println("Enter Server IP:Port ::: ")
	reader := bufio.NewReader(os.Stdin)
	serverID, err := reader.ReadString('\n')

	if err != nil {
		log.Printf("Failed to read from console :: %v", err)
	}
	serverID = strings.Trim(serverID, "\r\n")

	log.Println("Connecting : " + serverID)

	//connect to grpc server
	conn, err := grpc.Dial(serverID, grpc.WithInsecure())
	getConn(conn)
	if err != nil {
		log.Fatalf("Failed to conncet to gRPC server :: %v", err)
	}
	defer conn.Close()

	//call ChatService to create a stream
	client := chatserver.NewServicesClient(conn)

	stream, err := client.ChatService(context.Background())
	if err != nil {
		log.Fatalf("Failed to call ChatService :: %v", err)
	}

	// implement communication with gRPC server
	ch := clienthandle{stream: stream}
	ch.clientConfig()
	go ch.sendMessage()
	go ch.receiveMessage()

	//blocker

	bl := make(chan bool)
	<-bl

}

func leave() {
	k1.Close()
}

var k1 *grpc.ClientConn = nil

func getConn(madeconn *grpc.ClientConn) {
	k1 = madeconn
}

// clienthandle
type clienthandle struct {
	stream     chatserver.Services_ChatServiceClient
	clientName string
}

func (ch *clienthandle) clientConfig() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Your Name : ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf(" Failed to read from console :: %v", err)
	}
	ch.clientName = strings.Trim(name, "\r\n")
	//send "has joined here"
	ch.notifyJoin()
}

func (ch *clienthandle) notifyJoin() {
	cMessage := &chatserver.FromClient{
		Name: ch.clientName,
		Body: "{has joined the chat}",
	}
	ch.stream.Send(cMessage)
}

// send message
func (ch *clienthandle) sendMessage() {

	// create a loop
	for {

		reader := bufio.NewReader(os.Stdin)
		clientMessage, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf(" Failed to read from console :: %v", err)
		}
		clientMessage = strings.Trim(clientMessage, "\r\n")

		clientMessageBox := &chatserver.FromClient{
			Name:    ch.clientName,
			Body:    clientMessage,
			LogTime: time.Now().GoString(),
		}

		err = ch.stream.Send(clientMessageBox)

		if err != nil {
			log.Printf("Error while sending message to server :: %v", err)
		}

	}

}

// receive message
func (ch *clienthandle) receiveMessage() {

	//create a loop
	for {
		mssg, err := ch.stream.Recv()
		if err != nil {

		}

		if mssg.Body == "leave" && mssg.Name == ch.clientName {
			leave()
		}

		//print message to console
		fmt.Printf("%s : %s \n", mssg.Name, mssg.Body+" ["+mssg.LogTime+"]")

	}
}
