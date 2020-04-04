package spammer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type Spammer struct {
	addressToSpam string
	closeChan     chan bool
}

type Spam struct {
	Name  string
	Email string
}

func NewSpammer(addressToSpam string) *Spammer {
	return &Spammer{addressToSpam, make(chan bool)}
}

func (s Spammer) Start() {
	go func() {
		for {
			select {
			case <-s.closeChan:
				return
			default:
				s.spam()
				time.Sleep(250 * time.Millisecond)
			}
		}
	}()
}

func (s Spammer) spam() {
	name := base64.URLEncoding.EncodeToString([]byte(strconv.Itoa(rand.Int())))

	requestBody, err := json.Marshal(Spam{
		Name:  name,
		Email: fmt.Sprintf("%s@gmail.com", name),
	})

	if err != nil {
		panic(err)
	}

	_, err = http.Post(fmt.Sprintf("http://%s/%s", s.addressToSpam, "spam"), "application/json", bytes.NewBuffer(requestBody))

	if err != nil {
		panic(err)
	}
}

func (s *Spammer) Stop() {
	s.closeChan <- true
}
