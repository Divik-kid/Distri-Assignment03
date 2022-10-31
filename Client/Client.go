package main

import (
	"bufio"
	"context"
	"fmt"
	"grpcChatServer/chatserver"
	"log"
	"os"
	"strconv"
	"strings"

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

// clienthandle
type clienthandle struct {
	stream     chatserver.Services_ChatServiceClient
	clientName string
	clientTime string
}

func (ch *clienthandle) clientConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Your Name : ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf(" Failed to read from console :: %v", err)
	}
	ch.clientName = strings.Trim(name, "\r\n")
	ch.clientTime = "0"
	//send "has joined here"
	ch.notifyJoin()
}

func (ch *clienthandle) notifyJoin() {

	futureTime := lamportSend(ch.clientTime)
	ch.clientTime = futureTime

	cMessage := &chatserver.FromClient{
		Name:    ch.clientName,
		Body:    "has joined the chat",
		LogTime: futureTime,
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

		futureTime := lamportSend(ch.clientTime)

		clientMessageBox := &chatserver.FromClient{
			Name:    ch.clientName,
			Body:    clientMessage,
			LogTime: futureTime,
		}
		ch.clientTime = futureTime

		err = ch.stream.Send(clientMessageBox)

		if err != nil {
			log.Printf("Error while sending message to server :: %v", err)
		}
	}
}

func lamportSend(a string) string {
	ai, _ := strconv.Atoi(a)
	return strconv.Itoa(ai + 1)
}

func lamportRecive(a, b string) string {
	ai, _ := strconv.Atoi(a)
	bi, _ := strconv.Atoi(b)
	out := max(ai, bi) + 1
	return strconv.Itoa(out)
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// receive message
func (ch *clienthandle) receiveMessage() {

	//create a loop
	for {
		mssg, err := ch.stream.Recv()
		if err != nil {
			log.Printf("Error in receiving message from server :: %v", err)
		}

		ch.clientTime = lamportRecive(ch.clientTime, mssg.LogTime)

		//print message to console
		fmt.Printf("%s : %s \n", mssg.Name, mssg.Body+" ["+ch.clientTime+"]")
	}
}
