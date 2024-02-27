FROM golang:1.21-alpine
 
WORKDIR /app
 
COPY ./configuration.txt /app/configuration.txt
COPY ./src/registry.go /app/registry.go

# Builds your app with optional configuration
RUN go build registry.go 

EXPOSE 1234
 
# Specifies the executable command that runs when the container starts
CMD [ "./registry" ]

