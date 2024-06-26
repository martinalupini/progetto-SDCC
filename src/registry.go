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
	"time"
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

type Info struct {
	Peers []Node
	AssignedID []int

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
//the id of the registry: 1 if it's the main, 2 if it's the backup
var id string 
//signals the main registry is working
var working = true
//signals if the backup registry has already started to check if main registry is alive
var start_checking = false



// auxiliary functions of the service registry //
/*
*  printPeers shows the peers in the system
*/
func printPeers(){

	var p = fmt.Sprintf("REGISTRY %s --- PEERS:  ", id)
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
*  Sync is invoked by the main registry to keep its beckup replica up to date each time a new peer is added
*  @newNode: the new node to add
*  @reply: the reply to the registry
*/
func (r *ServiceRegistry) Sync(newNode Node, reply *int) error {
	lastPeer++
	peers = append(peers, newNode)
	log.Printf("REGISTRY %s SYNCHRONIZATION ---  New peer address: %s ID:%d",id, newNode.Addr,newNode.ID)
	assignedID = append(assignedID, newNode.ID)
	
	return nil

}



/*
*  AddNode is invoked by the node when it wants to enter the network
*  @newNode: the new node to add
*  @reply: the reply to the node
*/
func (r *ServiceRegistry) AddNode(newNode Node, reply *Neighbours) error {
	var response int
	
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
		assignedID = append(assignedID, newNode.ID)
		log.Printf("REGISTRY %s ---  New peer address: %s ID:%d",id, newNode.Addr,newNode.ID)
			
	} else {
		reply.Pos = oldPos
		log.Printf("REGISTRY %s --- Peer %d with address %s connects again",id, reply.ID, newNode.Addr)
		
	}

	//updating the backup
	if present == false && id == "1" {
		 
		//connection to service registry 
		client, err := rpc.DialHTTP("tcp", "registry2:5678")
		if err != nil {
			goto out
		}
	
		err = client.Call("ServiceRegistry.Sync", &newNode, &response)
		if err != nil {
			goto out
		}
	
		client.Close()
	
	}
out:	
	reply.Peers = peers

	return nil
}

/*
*  RetrieveInfo is invoked by the main registry to obtain information from the backup registry (in case the main registry was down for a period)
*  @id: the id of the main registry
*  @reply: all the information to give to the main registry
*/
func (r *ServiceRegistry) RetrieveInfo(id string, reply *Info) error {
	
	//to start checking the main registry once one message from it is received
	if start_checking == false {
		start_checking = true
		go checkMainAlive()
	
	}
	
	reply.Peers = peers
	reply.AssignedID = assignedID
	working = true

	return nil
}


/*
*  checkMainAlive is executed continuously in the background by the backup registry to check if the main registry is still working
*/
func checkMainAlive(){
	for {
		if working == true {
			time.Sleep(2 * time.Second)
			client, err := net.DialTimeout("tcp", "registry:1234", 5*time.Second )
			if err != nil {
				log.Printf("REGISTRY BACKUP --- I'm the main registry")
				working = false
	
			} else {
				client.Close()
			}
			
		}
	
	
	}

}


func main() {
	var port string
	var hostname string

	id = os.Getenv("ID")
	port = os.Getenv("PORT")
	hostname = os.Getenv("HOSTNAME")
	
	if id == "1" {
		var reply Info 
		//connection to service registry backup to obtain updates
		client, err := rpc.DialHTTP("tcp", "registry2:5678")
		if err != nil {
			goto out
		}
	
		err = client.Call("ServiceRegistry.RetrieveInfo", id, &reply)
		if err != nil {
			goto out
		}
	
		peers = reply.Peers
		assignedID = reply.AssignedID
		client.Close()
	
	} 

out:	
	
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
    	
    	if id == "2" { log.Printf("REGISTRY --- Chosen algorithm is %s", algorithm) }
    	
	serviceRegistry := new(ServiceRegistry)
	rpc.Register(serviceRegistry)
	rpc.HandleHTTP()
	
	lis, err := net.Listen("tcp", hostname+":"+port)
	if err != nil {
		log.Fatal("REGISTRY --- Error while starting registry:", err)
	}
	log.Printf("REGISTRY %s --- Registry listens on port %s:%s", id, hostname, port)

	http.Serve(lis, nil)
	
	
}
