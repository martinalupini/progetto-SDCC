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
	"bufio"
	"os"
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

//peers in the network
var peers []Node
//index of last node
var lastPeer = 0
//algorithm chosen
var algorithm = ""
//IDs already assigned
var assignedID []int
//if the node that joins the network was already in the network
var present = false 
//the old index of the node if already in the network
var oldPos int


// auxiliary functions of the service registry //
/*
*  printPeers shows the peers in the system
*/
func printPeers(){
	var p = "REGISTRY --- PEERS:  "
	var i int
	for i=0; i<len(peers); i++ {
		p = p+ fmt.Sprintf("%d:ID %d Address %s  ", i,peers[i].ID, peers[i].Addr)
	
	}
	log.Printf(p)
}


// RPC functions //

/*
*  generateID is used by the service registry to check if the node had already joined the newtork in the past 
*  (and if so, retrieves all the information about the node). If not, the registry generates the unique ID for the node.
*  @newNode: the new node to check
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

/*
*  AddNode is invoked by the node when it wants to enter the network
*  @newNode: the new node to add
*  @reply: the reply to the node
*/
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
		log.Printf("REGISTRY ---  New peer address: %s ID:%d",newNode.Addr,newNode.ID)
			
	} else {
		reply.Pos = oldPos
		log.Printf("REGISTRY --- Peer %d with address %s connects again", reply.ID, newNode.Addr)
		
	}
	
	reply.Peers = peers

	return nil
}


func main() {
	
	//reading from the configuration file what algorithm to use
	readFile, err := os.Open("configuration.txt")
  
    	if err != nil {
        	fmt.Println(err)
    	}
    	fileScanner := bufio.NewScanner(readFile)

    	fileScanner.Scan() 
    	algorithm = fileScanner.Text()
    
    	readFile.Close()
	
 	algorithm = strings.ToLower(algorithm)
    	algorithm = strings.TrimRight(algorithm, "\n")
    	
    	//the default algorithm is Bully
    	if algorithm != "chang-roberts" && algorithm != "bully" {   algorithm = "chang-roberts"  }
    	
	serviceRegistry := new(ServiceRegistry)
	rpc.Register(serviceRegistry)
	rpc.HandleHTTP()
	
	lis, err := net.Listen("tcp", "registry:1234")
	if err != nil {
		log.Fatal("REGISTRY --- Error while starting registry:", err)
	}
	log.Printf("REGISTRY --- Registry listens on port %d", 1234)

	http.Serve(lis, nil)
	
	
}
