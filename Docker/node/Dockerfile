FROM golang:1.21-alpine
 
WORKDIR /app
 
COPY . .
 
# Builds your app with optional configuration
RUN go build ./src/node.go 
 
# Specifies the executable command that runs when the container starts. 
CMD [ "./node"]
