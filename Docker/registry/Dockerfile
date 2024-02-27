FROM golang:1.21-alpine
 
WORKDIR /app
 
COPY ./configuration.txt /app/configuration.txt
COPY ./src/registry.go /app/registry.go

RUN go build registry.go 

EXPOSE 1234
 
CMD [ "./registry" ]

