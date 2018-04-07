package snpp

import (
	"bufio"
	"log"
	"net"
	"strings"
	"time"
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
	response, err := r.ReadString('\n')
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
		select {
		case msg, ok := <-msgchan:
			if ok {
				//parse message from channel

				splitmsg := strings.Split(msg, ";")
				pin, text := splitmsg[0], splitmsg[1]
				log.Printf("Message pulled from queue:%v", msg)

				//initiate page
				log.Printf("sending page to snpp server for pin:%v\n", pin)
				snpp.Write([]byte("PAGE " + pin + "\r\n"))

				//read response from page initiate
				log.Println("Reading snpp server response")
				response, err = r.ReadString('\n')
				log.Printf("Response:%v", response)
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
					log.Printf("sending mess to snpp server")
					snpp.Write([]byte("MESS " + text + "\r\n"))
					log.Println("Reading snpp server response")
					response, err = r.ReadString('\n')
					log.Printf("Response:%v", response)
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
						log.Printf("sending send to snpp server")
						snpp.Write([]byte("SEND\r\n"))
						log.Println("Reading snpp server response")
						response, err = r.ReadString('\n')
						log.Printf("Response:%v", response)
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
		default:
			log.Println("No msgs to process sleeping for 5 seconds")
			time.Sleep(5 * time.Second)
		}
	}
}
