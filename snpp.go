package snpp

import (
	"bufio"
	"log"
	"net"
	"strings"
)

//snpp level 1 client
func Client(msgchan chan string, addr string) {
	log.Printf("Connecting SNPP Client to %v...\n\n", addr)
Init:
	//dial server
	snpp, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("Error dialing snpp server: %v\n", err.Error())
		return
	}
	r := bufio.NewReader(snpp)

	//server connection init
	response, err := r.ReadString('\r')
	if err != nil {
		log.Printf("Error reading response from SNPP server: %v\n", err.Error())
		log.Println("Closing connection and starting new connection.")
		snpp.Close()
		goto Init
	}
	if !strings.Contains(response, "220") {
		log.Printf("SNPP Server was not ready. response was: %v\n", response)
		snpp.Close()
		goto Init
	}
	log.Println("SNPP server ready")
	//start sending messages
	for {

		//parse message from channel
		msg := <-msgchan
		splitmsg := strings.Split(msg, ";")
		pin, text := splitmsg[0], splitmsg[1]

		//initiate page
		snpp.Write([]byte("PAGE " + pin + "\r"))

		//read response from page initiate
		response, err = r.ReadString('\r')
		if err != nil {
			log.Printf("Error reading response from SNPP server: %v\n", err.Error())
			log.Println("Closing connection and starting new connection.")
			snpp.Close()
			msgchan <- msg
			goto Init
		}
		//if page not accepted move on and throw out the message
		if !strings.Contains(response, "250") {
			log.Printf("SNPP Server did not accept pager id. response was: %v\n", response)
			log.Printf("THROWING OUT MSG: %v\n", msg)
		} else {
			snpp.Write([]byte("MESS " + text + "\r"))
			response, err = r.ReadString('\r')
			if err != nil {
				log.Printf("Error reading response from SNPP server: %v\n", err.Error())
				log.Println("Closing connection and starting new connection.")
				snpp.Close()
				msgchan <- msg
				log.Printf("REQUEUED MSG: %v\n", msg)
				goto Init
			}

			if !strings.Contains(response, "250") {
				log.Printf("SNPP Server did not accept message. response was: %v\n", response)
				msgchan <- msg
				log.Printf("REQUEUED MSG: %v\n", msg)
			} else {
				snpp.Write([]byte("SEND\r"))
				response, err = r.ReadString('\r')
				if err != nil {
					log.Printf("Error reading response from SNPP server: %v\n", err.Error())
					log.Println("Closing connection and starting new connection.")
					snpp.Close()
					msgchan <- msg
					log.Printf("REQUEUED MSG: %v\n", msg)
					goto Init
				}
				if !strings.Contains(response, "250") {
					log.Printf("SNPP Server did not send message. response was: %v\n", response)
					msgchan <- msg
					log.Printf("REQUEUED MSG: %v\n", msg)

				} else {
					log.Printf("<%v> Sent to SNPP Server", msg)
				}
			}

		}
	}
}
