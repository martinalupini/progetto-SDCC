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
	Next *Node
	Prev *Node

} 


type Neighbours struct {
	ID int
	Algorithm string
	//LeLann
	Prev Node
	Next Node
	//Bully
	Peers []Node
	Present bool
}


// variables of the registry //
var peers []Node
var mapIDAddr = make(map[string]Node)
var lastPeer = -1
var algorithm string
var assignedID []int
var present = false


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
	var node Node
	
	present = false
	//checking if node were already in the system
	node, ok := mapIDAddr[newNode.Addr]
	if ok {  
		present = true
		return node.ID }
	
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
	
	if algorithm == "lelann" {
		if present == false {
		
			//adding prev and next node
			if lastPeer < 0 {
				reply.Next = Node{} //empty struct
				reply.Prev = Node{}
				newNode.Next = nil
				newNode.Prev = nil
			}else{
				//peers[0].Prev = &newNode
				//peers[lastPeer].Next = &newNode
				reply.Next = peers[0]
				reply.Prev = peers[lastPeer]
				newNode.Next = &peers[0]
				newNode.Prev = &peers[lastPeer]
				
				//updating prev and next
				if entry, ok := mapIDAddr[peers[0].Addr]; ok {
					entry.Prev = &newNode
      					mapIDAddr[peers[0].Addr] = entry
      				}
      				
      				if entry, ok := mapIDAddr[peers[lastPeer].Addr]; ok {
					entry.Next = &newNode
      					mapIDAddr[peers[lastPeer].Addr] = entry
      				}
			}
			lastPeer++
		} else {
			//node already present in the system
			reply.Next = *(mapIDAddr[newNode.Addr].Next)
			reply.Prev = *(mapIDAddr[newNode.Addr].Prev)
		
		}
	} else {
		var i int
		for i=0; i<len(peers); i++ {
			var newPeer Node
			newPeer.ID = peers[i].ID
			newPeer.Addr = strings.Clone(peers[i].Addr)
			reply.Peers = append(reply.Peers, newPeer)
		
		}
	}
	
	if present == false {
		//adding peer to the service registry data structures
		mapIDAddr[newNode.Addr] = newNode
		peers = append(peers, newNode)
		log.Printf("New peer address: %s ID:%d",newNode.Addr,mapIDAddr[newNode.Addr].ID)
	}
	
	//TO REMOVE
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
