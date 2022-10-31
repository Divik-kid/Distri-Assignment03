package chatserver

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type messageUnit struct {
	ClientName        string
	MessageBody       string
	MessageUniqueCode int
	ClientUniqueCode  int
	LogTime           string
}

type messageHandle struct {
	MQue []messageUnit
	mu   sync.Mutex
}

var clients = make(map[int]Services_ChatServiceServer, 10)

var messageHandleObject = messageHandle{}

type ChatServer struct {
}

// define ChatService
func (is *ChatServer) ChatService(csi Services_ChatServiceServer) error {

	clientUniqueCode := rand.Intn(1e6)
	errch := make(chan error)

	// receive messages - init a go routine
	go receiveFromStream(csi, clientUniqueCode, errch)

	// send messages - init a go routine
	go sendToStream(csi, clientUniqueCode, errch)

	return <-errch

}

// receive messages
func receiveFromStream(csi_ Services_ChatServiceServer, clientUniqueCode_ int, errch_ chan error) {

	//implement a loop
	for {
		mssg, err := csi_.Recv()
		_, ok := clients[clientUniqueCode_]
		//if new user
		if ok == false {
			clients[clientUniqueCode_] = csi_
		}

		if err != nil {
			log.Printf("Error in receiving message from client :: %v", err)
			errch_ <- err
		} else if mssg.Body == "_leave" {

			log.Printf(mssg.Name + " is leaving the chat")
			askToLeave(csi_, clientUniqueCode_, errch_)
			//askToLeave(csi_ Services_ChatServiceServer, clientUniqueCode_ int, errch_ chan error);

		} else {

			messageHandleObject.mu.Lock()

			messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
				ClientName:        mssg.Name,
				MessageBody:       mssg.Body,
				MessageUniqueCode: rand.Intn(1e8),
				ClientUniqueCode:  clientUniqueCode_,
				//fix the time
				LogTime: time.Now().GoString(),
			})

			log.Printf("%v", messageHandleObject.MQue[len(messageHandleObject.MQue)-1])

			messageHandleObject.mu.Unlock()

		}
	}
}

// send message
func sendToStream(csi_ Services_ChatServiceServer, clientUniqueCode_ int, errch_ chan error) {

	//implement a loop
	for {

		//loop through messages in MQue
		for {

			time.Sleep(500 * time.Millisecond)

			messageHandleObject.mu.Lock()

			if len(messageHandleObject.MQue) == 0 {
				messageHandleObject.mu.Unlock()
				break
			}

			senderUniqueCode := messageHandleObject.MQue[0].ClientUniqueCode
			senderName4Client := messageHandleObject.MQue[0].ClientName
			message4Client := messageHandleObject.MQue[0].MessageBody
			timeFromClient := messageHandleObject.MQue[0].LogTime
			var keys []int
			messageHandleObject.mu.Unlock()
			//send the message to every client
			for k, element := range clients {
				if clients[senderUniqueCode] != element {
					keys = append(keys, k)

					//send message to designated client (do not send to the same client)
					//if clients[senderUniqueCode] != keys {

					err := element.Send(&FromServer{Name: senderName4Client, Body: message4Client, LogTime: timeFromClient})

					if err != nil {
						errch_ <- err
					}

					messageHandleObject.mu.Lock()

					// delete the message at index 0 after sending to receiver
					if len(messageHandleObject.MQue) > 1 {
						messageHandleObject.MQue = messageHandleObject.MQue[1:]
					} else {
						messageHandleObject.MQue = []messageUnit{}
					}

					messageHandleObject.mu.Unlock()

					//}
				}
			}

		}

		time.Sleep(100 * time.Millisecond)
	}
}

// send message
func askToLeave(csi_ Services_ChatServiceServer, clientUniqueCode_ int, errch_ chan error) {

	//implement a loop
	for {

		//loop through messages in MQue
		for {

			time.Sleep(500 * time.Millisecond)

			messageHandleObject.mu.Lock()

			if len(messageHandleObject.MQue) == 0 {
				messageHandleObject.mu.Unlock()
				break
			}

			senderUniqueCode := messageHandleObject.MQue[0].ClientUniqueCode
			senderName4Client := messageHandleObject.MQue[0].ClientName
			timeFromClient := messageHandleObject.MQue[0].LogTime

			messageHandleObject.mu.Unlock()

			//send message to designated client (do not send to the same client)
			if senderUniqueCode == clientUniqueCode_ {

				err := csi_.Send(&FromServer{Name: senderName4Client, Body: "leave", LogTime: timeFromClient})

				if err != nil {
					errch_ <- err
				}

				messageHandleObject.mu.Lock()

				// delete the message at index 0 after sending to receiver
				if len(messageHandleObject.MQue) > 1 {
					messageHandleObject.MQue = messageHandleObject.MQue[1:]
				} else {
					messageHandleObject.MQue = []messageUnit{}
				}

				messageHandleObject.mu.Unlock()

			}

		}

		time.Sleep(100 * time.Millisecond)
	}
}
