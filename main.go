package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"time"
)

const (
	matrixToken = ""
	matrixEndpoint = ""

	kchatAddr = "localhost:1337"
	kchatUser = "matrix_bridge"
	kchatPass = ""
)

type message struct {
	MsgType string `json:"msgtype"`
	Body string `json:"body"`
}

var R *regexp.Regexp

func authenticate(conn net.Conn, user, pass string) {
	fmt.Fprintln(conn, "/login " + user + " " + pass)
}

func handleMessage(conn net.Conn) {
	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println("Receive error")
			break;
		}
		log.Println("Received:", msg)
		if R.MatchString(msg) {
			res := R.FindStringSubmatch(msg)
			fmt.Println(res[1] + ": " + res[2])
			text := fmt.Sprintf("%s: %s", res[1], res[2])
			sendMessage(text);
		}
	}
}

func sendMessage(msg string) {
	m := message{
		MsgType: "m.text",
		Body: msg,
	}
	j, _ := json.Marshal(m)
	p, _ := rand.Prime(rand.Reader, 64)
	req, _ := http.NewRequest(http.MethodPut,
	    matrixEndpoint + p.String(), bytes.NewBuffer(j))
	req.Header.Set("Authorization", "Bearer " + matrixToken)
	client := &http.Client{}
	_, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	R = regexp.MustCompile(`^\r\x1b\[1;\d{2}m(?P<Nick>.+)\x1b\[0m: (?P<Message>.+)`)
	for {
		conn, err := net.Dial("tcp", kchatAddr)
		if err != nil {
			log.Println("Retrying in 3 seconds")
			time.Sleep(3 * time.Second)
			continue
		}
		log.Println("Succesfully connected")
		defer conn.Close()
		authenticate(conn, kchatUser, kchatPass)
		handleMessage(conn)
	}
}
