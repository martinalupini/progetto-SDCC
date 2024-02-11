package main

import (
	"log"
	"net"
	"net/rpc"
	"net/http"
)

type ServiceRegistry string


// struct used in RPC //
type Node struct{
	ID int
	Addr string
	Next *Node
	Prev *Node

} 

type Neighbours struct {
	Prev Node
	Next Node
}

// variables of the registry //
var peers []Node
var mapPeers = make(map[int]string)
var lastPeer = -1


// auxiliary functions of the service registry //
func printPeers(){
	var i int
	for i=0; i<len(peers); i++ {
		log.Printf("%d: %s", i, peers[i].Addr)
	
	}
}


// RPC functions //

/*
AddNode: function invoked by the peer when it wants to enter the network. The service registry registers
the new node's id and sends to it the list of the peers in the network.
*/
func (r *ServiceRegistry) AddNode(newNode Node, reply *Neighbours) error {
	
	if lastPeer < 0 {
		reply.Next = Node{} //empty struct
		reply.Prev = Node{}
	}else{
		reply.Next = peers[0]
		reply.Prev = peers[lastPeer]
	}
	
	lastPeer++

	peers = append(peers, newNode)
	mapPeers[newNode.ID] = newNode.Addr
	log.Printf("New peer address: %s ID:%d", mapPeers[newNode.ID], newNode.ID)
	
	log.Printf("Current peers in the system:") //TO REMOVE
	printPeers()
	
	return nil
}



func main() {

	serviceRegistry := new(ServiceRegistry)
	rpc.Register(serviceRegistry)
	rpc.HandleHTTP()
	
	lis, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Error while starting registry:", err)
	}
	log.Printf("Registry listens on port %d", 1234)

	http.Serve(lis, nil)
	
	
}
