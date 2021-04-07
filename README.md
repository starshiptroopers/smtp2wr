# SMTP2WR - An smtp to http (or another smtp) relay written in Go

SMTP2WR enables you to route incoming SMTP emails to other SMTP destinations and HTTPS endpoints.
See smtp2wr.conf and routes.conf on how to configure SMTP2WR.

SMTP2WR was forked from RTMAIL (https://github.com/themecloud/rtmail) and completely rewritten.

## Usage

Set your email routes in a routes.conf file (see example) and add the path to your routes file in the main config file (smtp2wr.conf).
You can set the TLS certificate and key as well as what SMTP relay to use in the main config file. 

Change your configs and start SMTP2WR with `smtp2wr` or use the docker image

## Building from source

`git clone https://github.com/starshiptroopers/smtp2wr.git`

`cd smtp2wr && make build`

## Creating the docker image

`make docker-image`

## Running 

###### from compiled binary

`./smtp2wr -config ./smtp2wr.conf`

###### from docker container

Use docker-compose.yml file

## Routes

You can configure your email routes in routes.conf, a route looks like this:
```
    [
        {
            "Recipient":"hello@starshiptroopers.dev",
            "Type":"HTTP",
            "Relay":"http://localhost:8080/mail/handler",
            "LocalhostOnly": false,
            "Timeout": 10
        },
        {
            "Recipient":".+@starshiptroopers.dev",
            "Relay":"some-else-mail-server.com:25",
            "Type":"SMTP",
            "Destination":"null@starshiptroopers.dev",
            "LocalhostOnly": false
        },
        {
            "Type":"SMTP",
            "Relay":"some-else-mail-server.com:25",
            "Username":"test",
            "Password":"secure",
            "LocalhostOnly": true
        }
    ]

```
Where recipient is a regex string of which incoming mail will have to match to activate the route. 
Type is the forwarding type, may be ither HTTP or SMTP.
Relay is where to forward the mail (smtp server address or http endpoint URL) 
Destination, Username and Password is applicable only for SMTP relays.
Destination is where to send the mail, if this is blank the destination address will not be changed.
Username and Password are SMTP relay credentials used to auth (plain text auth used by default) 

To create a local forwarding server (local open relay) you can add the following to the routes.
```
[
    {
        "Type":"SMTP",
        "LocalhostOnly": true
    }
]
```
This will match all recipients and will not change the recipient when routing. It is also only permitted from localhost to protect us from creating an open-replay that spam-bots can use.

## Config

```
{
    "Routes":"./routes.conf",
    "SMTPCert":"",
    "SMTPKey":"",
    "SMTPHostname":"localhost",
    "SMTPListen":"127.0.0.1:10025",
    "SMTPForceTLS":false,
    "SMTPVerboseLogging":false
}
```
This is the standard configuration. 
SMTPCert and SMTPKey specifies which TLS certificate and key to use on the incoming SMTP server.
Routes is the path to the routing file.

## Why

It can use to forward some private mail to different mailboxes or http endpoints for some automation.
I created this to process some automated emails from a service provider  

## Some security considerations

You have to start service with the root user if you want to bind service at default smtp port (25)
Don't do this if you can. For security purposes use docker image instead with a port forwarding.
