package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	s "root/server"
	"time"
)

//writer have chanel and node connection
//reader have connection listener and chanel for writer

type SlaveManager struct {
	net.Conn
	listener chan []byte
	sender   chan []byte
}

func NewSlaveManager(addr string, masterCh chan []byte) (*SlaveManager, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	timeDeadline := time.Now().Add(5 * time.Second)
	if err := conn.SetDeadline(timeDeadline); err != nil {
		return nil, err
	}

	s := &SlaveManager{
		Conn:     conn,
		listener: masterCh,
		sender:   make(chan []byte),
	}

	go s.Writer()
	go s.Reader()

	return s, nil
}

func (s *SlaveManager) Writer() {
	defer s.Conn.Close()

	for {
		payload := <-s.listener
		//make req

		_, err := s.Conn.Write(payload) //send to server reader
		if err != nil {
			fmt.Println(err)

			if io.EOF != nil {
				return
			}

			continue
		}
	}
}

func (s *SlaveManager) Reader() {
	for {
		buf := make([]byte, 4096)
		n, err := s.Conn.Read(buf)
		if err != nil {
			log.Println(err)

			if io.EOF != nil {
				return
			}

			continue
		}

		s.sender <- buf[:n] //send to writer
	}
}

type Master struct {
	slave    []*SlaveManager
	receiver chan []byte
}

func (m *Master) Writer(conn net.Conn) {
	for {
		payload := <-m.receiver

		//parse the resultes

		_, err := conn.Write(payload) //Return to our client
		if err != nil {
			log.Println(err)

			if io.EOF != nil {
				return
			}

			continue
		}
	}
}

func NewMaster(addrs []string) *Master {
	receiver, slave := make(chan []byte), make([]*SlaveManager, len(addrs))

	for index, value := range addrs {
		slaveManager, err := NewSlaveManager(value, receiver)
		if err != nil {
			log.Println(err)
			continue
		}

		slave[index] = slaveManager
	}

	return &Master{
		receiver: receiver,
		slave:    slave,
	}
}

func LoadConfig(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	addrs := struct {
		Addrs []string `json:"addrs"`
	}{}

	if err := json.Unmarshal(data, &addrs); err != nil {
		return nil, err
	}

	return addrs.Addrs, nil
}

func main() {
	configPath := "./config.json"
	addrs, err := LoadConfig(configPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	fmt.Println(addrs)

	master := NewMaster(addrs)

	mux := s.NewRouter()
	mux.HandleFunc("conn", master.Conn) //ws simulation
	//add new file and update in real time

	s.ListenAndServe(":5000", mux)
}

func (m *Master) Conn(c *s.Ctx) {
	go m.Writer(c.Conn) //set writer to return answer to our client

	buf := make([]byte, 4096)

	for {
		_, err := c.Conn.Read(buf)
		if err != nil {
			log.Println(err)

			if io.EOF != nil {
				break
			}

			continue
		}

		//make command router

		// send msg
	}
}
