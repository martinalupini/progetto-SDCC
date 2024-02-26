package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/http"
	"time"
	"reflect"
	"os"
)

type Peer string

// struct used in RPC //
type Neighbours struct {
	ID int
	Algorithm string
	Pos int
	Peers []Node
	Present bool 
}


type Node struct{
	ID int
	Addr string
	Pos int
} 

type ElectionMsg struct {
	StarterID int
	Number int
	Phase int
	Candidate Node
	From Node

}

// variables of the registry //
var CurrNode Node
var Leader = Node{}
var peers []Node
var iamLeader = false
var election = false
var algorithm string
var totalPeers = 0
var electionStarted = 0



func printPeers(){
	var p = fmt.Sprintf("NODE %d --- PEERS:  ", CurrNode.ID)
	var i int
	for i=0; i<len(peers); i++ {
		p = p+ fmt.Sprintf("%d:ID %d Address %s  ", i,peers[i].ID, peers[i].Addr)
	
	}
	log.Printf(p)
}

// RPC LeLan ///////////////////////

func (p *Peer) ElectionLeaderLeLann(msg ElectionMsg, reply *int) error {

	if msg.Candidate.ID == CurrNode.ID {
		log.Printf("NODE %d --- ELECTION %d number %d phase %d: I'm the leader",CurrNode.ID, msg.StarterID, msg.Number, msg.Phase)
		iamLeader = true
		election = false
		*reply = 1
		NotifyLeader(msg)
	}else if msg.Candidate.ID > CurrNode.ID {
		election = true
		log.Printf("NODE %d --- ELECTION %d number %d phase %d: Node %d wants to be candidate. Sending message to next node", CurrNode.ID, msg.StarterID, msg.Number, msg.Phase, msg.Candidate.ID)
		msg.From = CurrNode
		msg.Phase++
		ElectionLeLann(msg, false)
	}else{
		election = true
		log.Printf("NODE %d --- ELECTION %d number %d phase %d: Node %d has lower ID then mine. Running for the election :)", CurrNode.ID, msg.StarterID, msg.Number, msg.Phase, msg.Candidate.ID)
		msg.Candidate = CurrNode
		msg.From = CurrNode
		msg.StarterID = CurrNode.ID
		msg.Number = electionStarted
		msg.Phase = 0
		electionStarted++
		ElectionLeLann(msg, true)
	}
	
	
	return nil
}


func (p *Peer) ElectionBully(msg ElectionMsg, reply *int) error {
	election = true
	if msg.Candidate.ID < CurrNode.ID {
		log.Printf("NODE %d --- ELECTION %d number %d phase %d: Candidate has lower ID then mine. Starting new ELECTION %d", CurrNode.ID, msg.StarterID, msg.Number, msg.Phase, CurrNode.ID)
		msg.StarterID = CurrNode.ID
		msg.Number = electionStarted
		msg.Candidate = CurrNode
		msg.Phase = 0
		electionStarted++
		ElectionBully(msg)
	}
	*reply = -1
	return nil
}


func (p *Peer) NewLeader(msg ElectionMsg, reply *int) error {
	iamLeader = false
	election = false
	Leader = msg.Candidate
	log.Printf("NODE %d --- ELECTION %d number %d FINISHED: New leader is node %d", CurrNode.ID, msg.StarterID, msg. Number, msg.Candidate.ID) 
	
	return nil

}


func (p *Peer) NewPeer(newNode Node, reply *int) error {
 
 	peers = append(peers, newNode)
	totalPeers++
	log.Printf("NODE %d --- Added peer with ID %d and address %s at pos %d", CurrNode.ID, newNode.ID, newNode.Addr, newNode.Pos)
	return nil

}


//////////////////////////////

func NotifyLeader(msg ElectionMsg){
	var i int
	var reply int

	log.Printf("NODE %d --- ELECTION %d number %d: Informing other nodes I'm the leader", CurrNode.ID, msg.StarterID, msg.Number)

	for i=0; i<len(peers); i++ {
		
		if peers[i].ID == CurrNode.ID { continue }
		
		if IsAlive(peers[i].Addr) < 0 { continue }
		
		client, err := rpc.DialHTTP("tcp", peers[i].Addr)
		if err != nil {
			log.Fatal("Dial error", err)
		}
		
		err = client.Call("Peer.NewLeader", msg, &reply)
		if err != nil {
			log.Fatal("Error while calling RPC:", err)
		}
		
		client.Close()

	}

	return

}

func ElectionBully(msg ElectionMsg){
	var i int 
	var reply int
	var major []Node
	
	election = true
	
	if len(peers) == 1 { 
		election = false
		iamLeader = true
		Leader = CurrNode
		log.Printf("NODE %d --- STARTING ELECTION %d number %d: I'm the only node in the network. I'm the leader", CurrNode.ID, msg.StarterID, msg.Number) 
		return
		
	}
	
	for i=0; i<len(peers) ; i++ { 
		
		if peers[i].ID > CurrNode.ID {
			major = append(major, peers[i])
			log.Printf("NODE %d --- STARTING ELECTION %d number %d: Sending message to node %d", CurrNode.ID, msg.StarterID, msg.Number, peers[i].ID)
		}
		
		i = (i+1) % totalPeers
		
	
	}
	
	for i=0; i<len(major) ; i++ { 
	
		//checking if node is alive	
		if IsAlive(peers[i].Addr) < 0 { 
			log.Printf("NODE %d --- ELECTION %d number %d: node %d is not working", CurrNode.ID, msg.StarterID, msg.Number, major[i].ID)
			continue }
		
		client, err := rpc.DialHTTP("tcp", major[i].Addr)
		if err != nil {
			log.Fatal("Dial error", err)
		}
		
		log.Printf("NODE %d --- ELECTION %d number %d FINISHED: I'm not the leader. Received OK message from node %d", CurrNode.ID, msg.StarterID, msg.Number, major[i].ID)
		
		err = client.Call("Peer.ElectionBully", msg, &reply)
		if err != nil {
			log.Fatal("Error while calling RPC:", err)
		}
		
		if reply == -1 {  iamLeader = false}
		
		client.Close()
	}
	
	iamLeader = true
	election = false
	Leader = CurrNode
	
	NotifyLeader(msg)
}





func ElectionLeLann(msg ElectionMsg, starting bool){
	var i= (CurrNode.Pos + 1) % totalPeers //starting from next node
	var reply int

	election = true
	
	for i< len(peers){
	
		if len(peers) == 0 {
			iamLeader = true
			election = false
			Leader = CurrNode
			log.Printf("NODE %d --- STARTING ELECTION %d number %d: I'm the only node in the network. I'm the leader", CurrNode.ID, msg.StarterID, msg.Number)
			return
		
		}
		
		//checking if all nodes are not working
		if i == CurrNode.Pos {
			iamLeader = true
			election = false
			Leader = CurrNode
			if starting { log.Printf("NODE %d --- STARTING ELECTION %d number %d: No working node in the network. I'm the leader", CurrNode.ID, msg.StarterID, msg.Number)
			} else {
				log.Printf("NODE %d --- ELECTION %d number %d: No working node in the network. I'm the leader", CurrNode.ID, msg.StarterID, msg.Number) }
			return
		
		}
		
		//checking if next node is responding. If not, passing to next one in the ring until one is reachable
		if IsAlive(peers[i].Addr) < 0 {
			log.Printf("NODE %d --- Node %d not responding, passing to next %s...", CurrNode.ID, peers[i].ID, peers[i].Addr)
			i++
			i = i % totalPeers
			continue 
		}
		
		if starting {
			log.Printf("NODE %d --- STARTING ELECTION %d number %d: Sending message to node %d", CurrNode.ID, msg.StarterID, msg.Number, peers[i].ID)
		} else {
			log.Printf("NODE %d --- ELECTION %d number %d phase %d: Sending message to node %d",CurrNode.ID, msg.StarterID, msg.Number, msg.Phase, peers[i].ID)
		}
		
		msg.Phase++
		
		client, err := rpc.DialHTTP("tcp", peers[i].Addr)
		if err != nil {
			log.Fatal("Dial error", peers[i].ID)
		}
	
		err = client.Call("Peer.ElectionLeaderLeLann", msg, &reply)
		if err != nil {
			log.Fatal("Error while calling RPC:", err)
		}
		
		if(reply == 1){
			iamLeader = false
			election = false 
			Leader = peers[i]
			log.Printf("NODE %d --- ELECTION %d number %d FINISHED: New leader is %d", CurrNode.ID, msg.StarterID, msg.Number, Leader.ID)
		}
		
		client.Close()
		return
	
	}
		
}



func CheckLeaderAlive(){

	var ret int
	var msg ElectionMsg
	
	if reflect.ValueOf(Leader).IsZero() && election == false {
		msg.StarterID = CurrNode.ID
		msg.Number = electionStarted
		msg.Phase = 0
		msg.Candidate = CurrNode
		msg.From = CurrNode
		electionStarted++
		if algorithm == "lelann" {  ElectionLeLann(msg, true) 
		} else { ElectionBully(msg) }
		
	}
		
	for {
		time.Sleep(2 * time.Second) 
		if iamLeader == false && election == false {
			//log.Printf("Checking if leader %d is alive...", Leader.ID)
			ret = IsAlive(Leader.Addr)
			if ret < 0 {
				msg.StarterID = CurrNode.ID
				msg.Number = electionStarted
				msg.Phase = 0
				msg.Candidate = CurrNode
				msg.From = CurrNode
				electionStarted++
				log.Printf("NODE %d --- Leader is not responding", CurrNode.ID)
				if algorithm == "lelann" {  ElectionLeLann(msg, true) 
				} else { ElectionBully(msg) }
			}
		}
	}
}

func IsAlive(addr string) int {
	ret :=0
	client, err := net.DialTimeout("tcp", addr, 5*time.Second )
	if err != nil {
		return -1
	
	} else {
		client.Close()
	}
	
	return ret

}



func updatePeers(proc string) {

	var i int
	var reply int
	
	for i=0; i<len(peers); i++ {
		if peers[i].ID == CurrNode.ID { continue }
		
		if IsAlive(peers[i].Addr) != 0 { continue }
		
		client, err := rpc.DialHTTP("tcp", peers[i].Addr)
		if err != nil {
			log.Fatal("Error while connecting to peer:", err)
		}
		
		if proc == "NewPeer" {
			err = client.Call("Peer.NewPeer", CurrNode, &reply)
			if err != nil {
				log.Fatal("Error while calling RPC:", err)
			}
		}
		
		
		client.Close()
	}
	
}



func GetPeers() {

	serviceRegistry := "registry:1234"
	var reply Neighbours
	
	//connection to service registry 
	client, err := rpc.DialHTTP("tcp", serviceRegistry)
	if err != nil {
		log.Fatal("Error while connecting to registry server:", err)
	}
	
	err = client.Call("ServiceRegistry.AddNode", &CurrNode, &reply)
	if err != nil {
		log.Fatal("Error while calling RPC:", err)
	}
	
	client.Close()
	
	CurrNode.ID = reply.ID
	CurrNode.Pos = reply.Pos
	algorithm = reply.Algorithm
	peers = reply.Peers
	totalPeers = len(peers)
	
	//log.Printf("NODE %d --- Hi I'm the node with ID %d and I listen on %s",CurrNode.ID, CurrNode.ID, CurrNode.Addr)
	//printPeers()
	
	if reply.Present == false { updatePeers("NewPeer") }

}


func main() {
	
	var port string
	var hostname string
	
    	port = os.Getenv("PORT")
    	hostname = os.Getenv("HOSTNAME")


	peer := new(Peer)
	rpc.Register(peer)
	rpc.HandleHTTP()
	
	CurrNode.Addr = fmt.Sprintf("%s:%s",hostname, port)
	
	lis, err := net.Listen("tcp", CurrNode.Addr )
	
	if err != nil {
		log.Fatal("Error while starting RPC server:", err)
	}

	GetPeers()
	
	go CheckLeaderAlive()

	http.Serve(lis, nil) 
	
}
