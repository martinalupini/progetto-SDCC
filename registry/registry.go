package main

import (
	"log"
	"net"
	"fmt"
	"net/rpc"
	"net/http"
	"strings"
	"math/rand"
	"slices"
)

type ServiceRegistry string

// struct used in RPC //
type Node struct{
	ID int
	Addr string
	Pos int
} 


type Neighbours struct {
	ID int
	Algorithm string
	Pos int
	Peers []Node
	Present bool
}


// variables of the registry //
var peers []Node
var lastPeer = 0
var algorithm string
var assignedID []int
var present = false 
var oldPos int


// auxiliary functions of the service registry //
func printPeers(){
	var i int
	for i=0; i<len(peers); i++ {
		log.Printf("%d:ID %d Address %s", i,peers[i].ID, peers[i].Addr)
	
	}
}


// RPC functions //

/*
AddNode: function invoked by the peer when it wants to enter the network. The service registry generates the unique ID for the peer.
*/

func generateID(newNode Node) int {
	var id int
	var i int
	
	present = false
	//checking if node were already in the system
	for i=0; i<len(peers); i++ {
		if peers[i].Addr == newNode.Addr {
			present = true
			oldPos = i
			return peers[i].ID
		}	
	
	}
	
	//if node does not exists I need to generate an ID
	for {
		id = rand.Intn(200)
		//checking if generated id is already assigned to another node
		if slices.Contains(assignedID, id) == false { return id }
	}
}


func (r *ServiceRegistry) AddNode(newNode Node, reply *Neighbours) error {
	
	//generating ID
	reply.ID = generateID(newNode)
	newNode.ID = reply.ID
	
	//selecting the algoritm 
	reply.Algorithm = algorithm
	reply.Present = present
	
	if present == false {
		reply.Pos = lastPeer
		lastPeer++
		peers = append(peers, newNode)
			
	} else {
		reply.Pos = oldPos
		
	}
	
	reply.Peers = peers
	log.Printf("New peer address: %s ID:%d",newNode.Addr,newNode.ID)

	log.Printf("Current peers in the system:") 
	printPeers()
	
	return nil
}


func main() {
	var alg string
	
    	for {
    		
		fmt.Println("Select one algorithm between LeLann and Bully:")
    		_, err := fmt.Scanln(&alg)
    		if err != nil {
        		log.Fatal(err)
    		}
    		alg = strings.ToLower(alg)
    		alg = strings.TrimRight(alg, "\n")
    		if alg != "lelann" && alg != "bully" { 
    			fmt.Println("Your selection is not valide. Please select one algorithm between LeLann and Bully:")
    		}else{
    			break
    		}
    	}
    	
    	algorithm = alg	

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
