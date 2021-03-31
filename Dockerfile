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
COPY bin/smtp2wr ./smtp2wr
COPY *.conf ./configs/
RUN ln -s configs/smtp2wr.conf smtp2wr.conf 
RUN ln -s configs/routes.conf routes.conf 
CMD ["./smtp2wr"]
