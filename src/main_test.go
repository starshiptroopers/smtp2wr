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

//start mail server with default configuration
//start http server as http endpoint for forwarded messages
//create smtp client, do the test smtp message and send them to the mail server
//receive this message at http server endpoint
func TestService(t *testing.T) {

	host := "127.0.0.1"
	SMTPTo := "nobody@starshiptroopers.dev"
	SMTPFrom := "test@test.test"

	SMTPPort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal("can't get a free tcp port")
	}

	HTTPPort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal("can't get a free tcp port")
	}

	config := Config{
		SMTPHostname: "localhost",
		SMTPListen:   fmt.Sprintf("%s:%d", host, SMTPPort),
	}
	routes := []Route{{
		Recipient:     ".+@starshiptroopers.dev",
		Type:          "HTTP",
		LocalhostOnly: false,
		Relay:         fmt.Sprintf("http://%s:%d/data", host, HTTPPort),
	},
	}

	launcher, err := TestingSMTPServer(config, routes)
	if err != nil {
		t.Fatalf("can't create smtp server: %v", err)
	}

	err = TestingServiceStart(launcher)
	if err != nil {
		t.Fatalf("can't start web server: %v", err)
	}

	launcher, err = TestingWebServer(fmt.Sprintf("%s:%d", host, HTTPPort))
	if err != nil {
		t.Fatalf("can't create web server: %v", err)
	}

	err = TestingServiceStart(launcher)
	if err != nil {
		t.Fatalf("can't start web server: %v", err)
	}

	err = smtp.SendMail(
		routes[0].Relay,
		nil,
		SMTPFrom,
		[]string{SMTPTo},
		[]byte(fmt.Sprintf("To: %s\r\n"+
			"Subject: Just a test\r\n"+
			"\r\n"+
			"Hello\r\n", SMTPTo)),
	)

}

func TestingSMTPServer(config Config, routes []Route) (launcher func() error, err error) {
	defer func() {
		if panicErr := recover(); err != nil {
			err = fmt.Errorf("mail server startup failed: %v", panicErr)
		}
	}()
	return server(config, routes)
}

func TestingWebServer(host string) (launcher func() error, err error) {
	var mux http.ServeMux

	server := &http.Server{
		Addr:    host,
		Handler: &mux,
	}

	mux.HandleFunc("/data", func(res http.ResponseWriter, req *http.Request) {
		type Data struct {
			Sender     string
			Recipients []string
			Data       []byte
		}

		var d Data
		if req.Method == http.MethodPost {
			err := readBodyAsJSON(req, &d)
			if err != nil {
				http.Error(res, "wrong request data", http.StatusBadRequest)
				return
			}
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
