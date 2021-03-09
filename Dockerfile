# stage 1: build
#FROM golang:1.14 as build
#LABEL stage=intermediate
#WORKDIR /app
#COPY . .
#RUN make build

# stage 2: scratch
FROM alpine:latest as scratch
RUN apk --no-cache add ca-certificates
WORKDIR /opt/smtp2wr
COPY bin/smtp2wr /opt/smtp2wr/smtp2wr
COPY smtp2wr.conf /opt/smtp2wr
COPY routes.conf /opt/smtp2wr
CMD ["./smtp2wr"]
