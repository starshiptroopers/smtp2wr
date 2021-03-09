# SMTP2WR - An smtp to http (or another smtp) relay written in Go

SMTP2WR enables you to route incoming SMTP emails to other SMTP destinations and HTTPS endpoints.
See smtp2wr.conf and routes.conf on how to configure SMTP2WR.

SMTP2WR is forked from RTMAIL (https://github.com/themecloud/rtmail) and completely rewritten.

## Usage

Set your email routes in a routes.conf file (see example) and add the path to your routes file in the main config file (smtp2wr.conf).
You can set the TLS certificate and key as well as what SMTP relay to use in the main config file. 

## Installation

## From source

`git clone https://github.com/starshiptroopers/smtp2wr.git`
`cd smtp2wr`
`make build`
`./bin/smtp2wr`

## From Docker image

Docker image isn't published yet. But you can build your own image from sources with command:

`make docker-build`

Then start the service with:

`docker-compose up`

## Routes

You can configure your email routes in routes.conf, a route looks like this:
```
    {
        "Recipient":".+@example.com",
        "Type":"SMTP",
        "Relay": "mailrelay.example.com"
        "Destination":"me@example.com",
        "LocalhostOnly": false
        "Username":"username",
        "Password":"password",
    }
```
Where recipient is a regex string of which incoming mail will have to match to activate the route. 
Type is the forwarding type, may be ither HTTP or SMTP.
Relay is where to forward the mail (smtp server address or http endpoint URL) 
Destination, Username and Password is applicable only for SMTP relays.
Destination is where to send the mail, if this is blank the destination address will not be changed.
Username and Password are SMTP relay credentials used to auth (plain text auth used by default) 

To create a local forwarding server (local open relay) you can add the following to the routes.
```
    {
        "Type":"SMTP",
        "LocalhostOnly": true
    }
```
This will match all recipients and will not change the recipient when routing. It is also only permitted from localhost to protect us from creating an open-replay that spam-bots can use.

## Config

```
{
    "SMTPCert":"/etc/rtmail/cert.crt",
    "SMTPKey":"/etc/rtmail/cert.key",
    "SMTPHostname":"mail.example.com",
    "Routes":"/etc/rtmail/routes.conf",
    "SMTPListen":"10025"
}
```
This is the standard configuration. 
SMTPCert and SMTPKey specifies which TLS certificate and key to use on the incoming SMTP server.
Routes is the path to the routing file.

## Why

It can use to forward some private mail to different mailboxes or microservices for some automation.
I created this to process some automated emails from a service provider with my web service  

## Some security considerations

You have to start service with the root user if you want to bind service at default 25 port
Don't do this if you can. Use docker image instead with a port forwarding.
