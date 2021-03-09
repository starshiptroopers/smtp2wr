#!/bin/bash

FROM="test@localhost"
TO="hello@starshiptroopers.dev"
SMTP="127.0.0.1:10025"

curl --url "smtp://${SMTP}" --mail-from "${FROM}" --mail-rcpt "${TO}" -T <(echo -e "From: ${FROM}\nTo: ${TO}\nSubject: Mail Test\n\nHello")

