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

func (s Spammer) Start() error {
	var err error

	go func() {
		for {
			select {
			case <-s.closeChan:
				return
			default:
				if err = s.spam(); err != nil {
					s.Stop()
					break
				}
				time.Sleep(250 * time.Millisecond)
			}
		}
	}()

	return err
}

func (s Spammer) spam() error {
	name := base64.URLEncoding.EncodeToString([]byte(strconv.Itoa(rand.Int())))

	requestBody, err := json.Marshal(Spam{
		Name:  name,
		Email: fmt.Sprintf("%s@gmail.com", name),
	})

	if err != nil {
		return err
	}

	_, err = http.Post(fmt.Sprintf("http://%s/%s", s.addressToSpam, "spam"), "application/json", bytes.NewBuffer(requestBody))

	return err
}

func (s *Spammer) Stop() {
	s.closeChan <- true
	close(s.closeChan)
}
