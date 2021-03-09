package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/phayes/freeport"
	"io"
	"net/http"
	"net/smtp"
	"testing"
	"time"
)

type HttpRelayData struct {
	Sender     string
	Recipients []string
	Data       []byte
}

//start mail server with predefined configuration
//start http server to server http endpoint for forwarded messages
//create smtp client and send the mail to our mail server
//receive this message at http server endpoint and compare with original
func TestService(t *testing.T) {

	host := "127.0.0.1"
	SMTPTo := "nobody@starshiptroopers.dev"
	SMTPFrom := "test@test.test"
	const operationTimeout = time.Second

	SMTPPort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal("can't get a free tcp port")
	}

	HTTPPort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal("can't get a free tcp port")
	}

	config := Config{
		SMTPHostname:       "localhost",
		SMTPListen:         fmt.Sprintf("%s:%d", host, SMTPPort),
		SMTPVerboseLogging: true,
	}
	routes := []Route{
		{
			Recipient:     ".+@starshiptroopers.dev",
			Type:          "HTTP",
			LocalhostOnly: false,
			Relay:         fmt.Sprintf("http://%s:%d/data", host, HTTPPort),
			Timeout:       5,
		},
	}

	launcher, err := TestingSMTPServer(config, routes)
	if err != nil {
		t.Fatalf("can't create smtp server: %v", err)
	}

	err = TestingServiceStart(launcher)
	if err != nil {
		t.Fatalf("can't start smtp server: %v", err)
	}

	dataCh := make(chan interface{}, 1)

	launcher, err = TestingWebServer(fmt.Sprintf("%s:%d", host, HTTPPort), dataCh)
	if err != nil {
		t.Fatalf("can't create web server: %v", err)
	}

	err = TestingServiceStart(launcher)
	if err != nil {
		t.Fatalf("can't start web server: %v", err)
	}

	msg := []byte(fmt.Sprintf("To: %s\nFrom: %s\nSubject: Just a test\n\nHello\n.\n", SMTPTo, SMTPFrom))
	err = smtp.SendMail(
		config.SMTPListen,
		nil,
		SMTPFrom,
		[]string{SMTPTo},
		msg,
	)
	if err != nil {
		t.Errorf("can't send the mail: %v", err)
	}

	//waiting for data appears on http side
	select {
	case d := <-dataCh:
		if d != nil {
			if err, ok := d.(error); ok {
				t.Fatalf("we accept wrong data on http side: %v", err)
			}

			if data, ok := d.(*HttpRelayData); !ok {
				t.Fatalf("we accept wrong data on http side")
			} else {
				if bytes.Compare(msg, data.Data) != 0 {
					t.Fatalf("we accept wrong data on http side")
				}
			}
		}
		//wait for service startup
	case <-time.After(operationTimeout * 1000):
		t.Fatalf("no data received on the http server side")
	}
	return
}

func TestingSMTPServer(config Config, routes []Route) (launcher func() error, err error) {
	defer func() {
		if panicErr := recover(); err != nil {
			err = fmt.Errorf("mail server startup failed: %v", panicErr)
		}
	}()
	return server(config, routes)
}

func TestingWebServer(host string, dataChan chan interface{}) (launcher func() error, err error) {
	var mux http.ServeMux

	server := &http.Server{
		Addr:    host,
		Handler: &mux,
	}

	mux.HandleFunc("/data", func(res http.ResponseWriter, req *http.Request) {
		var d HttpRelayData
		if req.Method == http.MethodPost {
			err := readBodyAsJSON(req, &d)
			if err != nil {
				http.Error(res, "wrong request data", http.StatusBadRequest)
				return
			}
			dataChan <- &d
			http.Error(res, "", http.StatusOK)
		} else {
			http.Error(res, "Bad request", http.StatusBadRequest)
		}
	})

	return server.ListenAndServe, nil
}

//start service in background
func TestingServiceStart(launcher func() error) error {

	closeCh := make(chan error)

	go func() {
		err := launcher()
		if err != nil {
			closeCh <- err
		}
		close(closeCh)
	}()

	//wait for service became ready
	select {
	case err := <-closeCh:
		if err != nil {
			return fmt.Errorf("can't start service: %v", err)
		}
		//wait for service startup
	case <-time.After(time.Second):

	}
	return nil
}

func readBody(req *http.Request, maxLength int64) ([]byte, error) {
	b := bytes.NewBuffer(make([]byte, 0))
	if req.ContentLength > maxLength {
		return nil, fmt.Errorf("request body length greater then %d (maxRequestBody)", maxLength)
	}

	_, err := b.ReadFrom(io.LimitReader(req.Body, req.ContentLength))

	if err != nil {
		return nil, fmt.Errorf("can't read request body %v", err)
	}
	_ = req.Body.Close()

	return b.Bytes(), nil
}

func readBodyAsJSON(req *http.Request, j interface{}) error {
	b, err := readBody(req, 1024)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, j)
	if err != nil {
		return fmt.Errorf("can't parse json from request body %v", err)
	}
	return nil
}
