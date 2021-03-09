package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/chrj/smtpd"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"regexp"
	"strings"
)

type Route struct {
	Recipient     string //recipient regexp
	Type          string //relay type (http or smtp)
	Destination   string //relay addr
	LocalhostOnly bool   //route only localhost connections
	Relay         string //relay addr
	Username      string //username for auth on the default relay
	Password      string //password for auth on the default relay
}

type Config struct {
	Routes       string //routes config file
	SMTPCert     string //tls cert and key
	SMTPKey      string
	SMTPHostname string
	SMTPListen   string //
	SMTPForceTLS bool
}

var gitTag, gitCommit, gitBranch string

func main() {

	var config Config
	var configPath string
	var routes []Route

	if gitTag != "" {
		fmt.Printf("smtp2wr service version %s (%s, %s)", gitTag, gitBranch, gitCommit)
	}

	flag.StringVar(&configPath, "config", "./smtp2wr.conf", "SMTP2WR config file")

	if err := readJSONConfig(configPath, &config); err != nil {
		panic(err)
	}
	if err := readJSONConfig(config.Routes, &routes); err != nil {
		panic(err)
	}

	for _, route := range routes {
		if route.Recipient != "" {
			log.Printf("Handling %s to %s relay %s", route.Recipient, route.Type, route.Relay)
		}
	}

	flag.Parse()

	if launcher, err := server(config, routes); err != nil {
		log.Fatal(err)
	} else {
		launcher()
	}

	log.Println("Mailserver exiting peacefully.")
}

func server(c Config, routes []Route) (launcher func() error, err error) {

	var TLSConfig *tls.Config

	if c.SMTPCert != "" && c.SMTPKey != "" {
		if cert, err := tls.LoadX509KeyPair(c.SMTPCert, c.SMTPKey); err != nil {
			panic(err)
		} else {
			log.Println("Mailserver certificates loaded...")
			TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
		}
	}

	server := &smtpd.Server{

		Hostname: c.SMTPHostname,

		Handler: func(peer smtpd.Peer, env smtpd.Envelope) error {
			done := false
			for _, route := range routes {
				if !(route.LocalhostOnly && peer.Addr.String() != "127.0.0.1") {
					for _, recipient := range env.Recipients {
						match, _ := regexp.MatchString(route.Recipient, recipient)
						if match || route.Recipient == "" {
							lhead := fmt.Sprintf("Email for %s received ", recipient)
							if route.Relay == "" {
								log.Printf("%sbut relay isn't defined for recipient %s, route skipped", lhead, route.Recipient)
								continue
							}

							var dst []string
							if route.Destination != "" {
								dst = []string{route.Destination}
							} else {
								dst = env.Recipients
							}
							if route.Type == "SMTP" {
								var auth smtp.Auth
								if route.Username != "" {
									auth = smtp.PlainAuth(route.Username, route.Username, route.Password, strings.Split(route.Relay, ":")[0])
								}
								err := smtp.SendMail(
									route.Relay,
									auth,
									env.Sender,
									dst,
									env.Data,
								)

								if err != nil {
									log.Printf("%sbut can't forwarded to %s with %s relay: %v", lhead, dst, route.Relay, err)
								} else {
									log.Printf("%sand forwarded to %s with %s relay", lhead, dst, route.Relay)
									done = true
								}
								continue
							}
							if route.Type == "HTTP" {
								data, err := json.Marshal(env)
								if err != nil {
									log.Println("Unable to parse envelope from", env.Sender)
								} else {
									resp, err := http.Post(route.Relay, "application/json", bytes.NewBuffer(data))
									if err == nil && resp != nil && resp.StatusCode != http.StatusOK {
										err = errors.New("HTTP " + resp.Status)
									}
									if err != nil {
										log.Printf("%sbut can't forwarded to %s relay: %v", lhead, route.Relay, err)
									} else {
										log.Printf("%sand forwarded to %s relay", lhead, route.Relay)
										done = true
									}
									continue
								}
							}
						}
					}
				}
			}
			if !done {
				log.Printf("The mail from %s to %s has been rejected, no sutable routes available", env.Sender, env.Recipients)
				return errors.New("Invalid Recipient. This server does not handle the recipient")
			}
			return nil
		},

		RecipientChecker: func(peer smtpd.Peer, addr string) error {
			return nil
		},

		ForceTLS:  c.SMTPForceTLS,
		TLSConfig: TLSConfig,
	}

	return func() error {
		log.Println("Mailserver is listening at " + c.SMTPListen)
		return server.ListenAndServe(c.SMTPListen)
	}, nil
}

func readJSONConfig(path string, config interface{}) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("can't open configuration file %s: %v", path, err)
	}
	err = json.Unmarshal(b, config)
	if err != nil {
		return fmt.Errorf("can't parse configuration file %s: %v", path, err)
	}
	return nil
}
