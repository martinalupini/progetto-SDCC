FROM golang:1.21-alpine
 
WORKDIR /app
 
COPY . .

# Builds your app with optional configuration
RUN go build -o registry-exe ./registry/registry.go 

EXPOSE 1234
 
# Specifies the executable command that runs when the container starts
CMD [ “./registry-exe” ]
