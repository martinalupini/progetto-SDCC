package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/http"
	"time"
	"strings"
	"reflect"
)

type Peer string

// struct used in RPC //
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


type Node struct{
	ID int
	Addr string
	Next *Node
	Prev *Node
} 

type ElectionMsg struct {
	ElectionID int
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
var numElection = 0
var elMsg ElectionMsg


func printPeers(){
	var i int
	for i=0; i<len(peers); i++ {
		log.Printf("%d:ID %d Address %s", i,peers[i].ID, peers[i].Addr)
	
	}
}

// RPC LeLan ///////////////////////

func (p *Peer) UpdateNext(next Node, reply *int) error {

	CurrNode.Next = &next
	log.Printf("New next is node %d address %s", CurrNode.Next.ID, CurrNode.Next.Addr)
	
	return nil
}



func (p *Peer) ElectionLeaderLeLann(msg ElectionMsg, reply *int) error {

	if msg.Candidate.ID == CurrNode.ID {
		log.Printf("ELECTION %d: I'm the leader", msg.ElectionID)
		iamLeader = true
		election = false
		*reply = 1
		NotifyLeader(CurrNode, msg)
	}else if msg.Candidate.ID > CurrNode.ID {
		election = true
		log.Printf("ELECTION %d: Node %d wants to be candidate. Sending message to next node",msg.ElectionID, msg.Candidate.ID)
		msg.From = CurrNode
		ElectionLeLann(msg, false)
	}else{
		election = true
		log.Printf("ELECTION %d: Node %d has lower ID then mine. Running for the election :)", msg.ElectionID, msg.Candidate.ID)
		msg.Candidate = CurrNode
		msg.From = CurrNode
		msg.ElectionID = CurrNode.ID+ numElection
		numElection++
		ElectionLeLann(msg, true)
	}
	
	
	return nil
}


func (p *Peer) ElectionBully(msg ElectionMsg, reply *int) error {
	election = true
	if msg.Candidate.ID < CurrNode.ID {
		log.Printf("ELECTION %d: Candidatee has lower ID then mine. Starting new ELECTION %d", msg.ElectionID, CurrNode.ID)
		msg.ElectionID = CurrNode.ID
		msg.Candidate = CurrNode
		ElectionBully(msg)
	}
	return nil
}


func (p *Peer) NewLeader(msg ElectionMsg, reply *int) error {
	iamLeader = false
	election = false
	Leader = msg.Candidate
	log.Printf("ELECTION %d FINISHED: New leader is node %d",msg.ElectionID, msg.Candidate.ID) 
	
	return nil

}

// RPC Bully ///////////////////////

func (p *Peer) NewPeer(newNode Node, reply *int) error {
	var newPeer Node
	newPeer.ID = newNode.ID
	newPeer.Addr = strings.Clone(newNode.Addr)  
	peers = append(peers, newPeer)
	log.Printf("Added peer with ID %d and address %s", newNode.ID, newNode.Addr)
	return nil

}


//////////////////////////////

func NotifyLeader(except Node, msg ElectionMsg){
	var next *Node = CurrNode.Next
	var reply int


	for next != nil && next.ID != CurrNode.ID && next.ID != except.ID {
	
		if IsAlive(CurrNode.Next.Addr) < 0 {
			log.Printf("Node %d not responding, passing to next...", next.ID)
			next = next.Next
			continue 
		}
		
		log.Printf("ELECTION %d FINISHED: Informing node %d I'm the leader",msg.ElectionID, next.ID)
		client, err := rpc.DialHTTP("tcp", CurrNode.Next.Addr)
		if err != nil {
			log.Fatal("Dial error", err)
		}
		
		msg.Candidate = CurrNode
		err = client.Call("Peer.NewLeader", msg, &reply)
		if err != nil {
			log.Fatal("Error while calling RPC:", err)
		}
		
		next = next.Next
		client.Close()
	
	}
	
	return

}

func ElectionBully(msg ElectionMsg){
	var i int 
	var reply int
	
	election = true
	
	log.Printf("len is %d", len(peers))
	if len(peers) == 0 { log.Printf("STARTING ELECTION %d", msg.ElectionID) }
	
	
	for i=0; i<len(peers) ; i++ {
		
		if peers[i].ID == CurrNode.ID || peers[i].ID < CurrNode.ID {
			log.Printf("STARTING ELECTION %d", msg.ElectionID)
		 	continue }
		
		log.Printf("STARTING ELECTION %d: Sending message to node %d", msg.ElectionID, peers[i].ID)
		numElection++
		
		client, err := rpc.DialHTTP("tcp", peers[i].Addr)
		if err != nil {
			log.Fatal("Dial error", err)
		}
		
		err = client.Call("Peer.ElectionBully", msg, &reply)
		if err != nil {
			log.Fatal("Error while calling RPC:", err)
		}
		
		if IsAlive(peers[i].Addr) < 0 { continue }
		c := make(chan error, 1)
		go func() { c <- client.Call("Peer.ElectionBully", msg, &reply) } ()
			select {
  				case err := <-c:
  					if err != nil {
						log.Fatal("Error while calling RPC:", err)
					}
					iamLeader = false
					log.Printf("ELECTION %d FINISHED: I'm not the peer with highest ID",msg.ElectionID)
					client.Close()
					return
				
				//simulating the fact that the other peer does not respond	
  				case <-time.After(3 * time.Second):
    					iamLeader = false
					log.Printf("ELECTION %d FINISHED: I'm not the peer with highest ID",msg.ElectionID)
					client.Close()
					return
			}
		
		client.Close()
	}
	
	iamLeader = true
	election = false
	Leader = CurrNode
	
	log.Printf("ELECTION %d FINISHED: Informing other peers I'm the leader", msg.ElectionID)
	
	for i=0; i<len(peers) ; i++ {
		
		if peers[i].ID == CurrNode.ID { continue }
		
		client, err := rpc.DialHTTP("tcp", peers[i].Addr)
		if err != nil {
			log.Fatal("Dial error", err)
		}
		
		err = client.Call("Peer.NewLeader", msg, &reply)
		if err != nil {
			log.Fatal("Error while calling RPC:", err)
		}

	}
}





func ElectionLeLann(msg ElectionMsg, starting bool){
	var next *Node = CurrNode.Next
	var reply int

	election = true
	//only one node in the system
	if next == nil{
		iamLeader = true
		election = false
		Leader = CurrNode
		log.Printf("ELECTION %d: No one in the network. I'am the leader", CurrNode.ID)
		return
	}
	
	//there are other nodes in the system
	for next.ID != CurrNode.ID {
	
		//checking if next node is responding. If not, passing to next one in the ring until one is reachable
		if IsAlive(next.Addr) < 0 {
			log.Printf("Node %d not responding, passing to next %s...", next.ID, next.Next.Addr)
			next = next.Next
			continue 
		}
		
		if starting {
			log.Printf("STARTING ELECTION %d: Sending message to node %d", msg.ElectionID, next.ID)
			numElection++
		} else {
			log.Printf("ELECTION %d: Sending message to node %d",msg.ElectionID, next.ID)
		}
		
		client, err := rpc.DialHTTP("tcp", next.Addr)
		if err != nil {
			log.Fatal("Dial error", next.ID)
		}
		
		err = client.Call("Peer.ElectionLeaderLeLann", msg, &reply)
		if err != nil {
			log.Fatal("Error while calling RPC:", err)
		}
		
		if(reply == 1){
			iamLeader = false
			election = false 
			Leader = *next
			log.Printf("ELECTION %d FINISHED: New leader is %d",msg.ElectionID, Leader.ID)
			client.Close()
			return
		}
		
		client.Close()
		return
	}
	
	if next.ID == CurrNode.ID {
		iamLeader = true
		election = false
		Leader = CurrNode
		log.Printf("ELECTION %d: No working node in the network. I'm the leader", msg.ElectionID)
		return
	}


}



func CheckLeaderAlive(){

	
		var ret int
	
		if reflect.ValueOf(Leader).IsZero() {
			elMsg.ElectionID = CurrNode.ID
			elMsg.Candidate = CurrNode
			elMsg.From = CurrNode
			if algorithm == "lelann" {  ElectionLeLann(elMsg, true) 
			} else { ElectionBully(elMsg) }
		
		}
		
	for {
		time.Sleep(2 * time.Second) 
		if iamLeader == false && election == false {
			//log.Printf("Checking if leader %d is alive...", Leader.ID)
			ret = IsAlive(Leader.Addr)
			if ret < 0 {
				elMsg.ElectionID = CurrNode.ID
				elMsg.Candidate = CurrNode
				elMsg.From = CurrNode
				log.Printf("Leader is not responding")
				if algorithm == "lelann" {  ElectionLeLann(elMsg, true) 
				} else { ElectionBully(elMsg) }
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



func NotifyPrev(){
	var reply int
	
	client, err := rpc.DialHTTP("tcp", CurrNode.Prev.Addr)
	if err != nil {
		log.Fatal("Error while connecting to predecessor:", err)
	}
	
	err = client.Call("Peer.UpdateNext", CurrNode, &reply)
	if err != nil {
		log.Fatal("Error while calling RPC:", err)
	}

	client.Close()

}

func updatePeers(proc string) {

	var i int
	var reply int
	
	for i=0; i<len(peers); i++ {
		if peers[i].ID == CurrNode.ID { continue }
		client, err := rpc.DialHTTP("tcp", peers[i].Addr)
		if err != nil {
			log.Fatal("Error while connecting to peer:", err)
		}
		
		err = client.Call(fmt.Sprintf("Peer.%s", proc), CurrNode, &reply)
			if err != nil {
			log.Fatal("Error while calling RPC:", err)
			}
			
		client.Close()
		}
	
}



func GetPeers() {

	serviceRegistry := "localhost:1234"
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
	algorithm = reply.Algorithm
	
	log.Printf("Peer with ID %d listens on %s",CurrNode.ID, CurrNode.Addr)
	
	if algorithm == "lelann" {
	
		if reply.Next ==  (Node{}) {
			CurrNode.Next = nil
			CurrNode.Prev = nil
		} else {
			CurrNode.Next = &(reply.Next)
			CurrNode.Prev = &(reply.Prev)
		}
		
		if CurrNode.Next != nil {
			log.Printf("Next node is %d address %s", CurrNode.Next.ID, CurrNode.Next.Addr )
			log.Printf("Prev node is %d address %s", CurrNode.Prev.ID, CurrNode.Prev.Addr )
		}
		
		if CurrNode.Prev != nil && reply.Present == false{
			NotifyPrev()
		}

		
	} else {
		peers = reply.Peers
		printPeers()
		if reply.Present == false { updatePeers("NewPeer") }
	}

}

func main() {
	
	var port string
	_, err := fmt.Scanln(&port)
    	if err != nil {
        	log.Fatal(err)
    	}
    	port = strings.TrimRight(port, "\n")

	peer := new(Peer)
	rpc.Register(peer)
	rpc.HandleHTTP()
	
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	
	if err != nil {
		log.Fatal("Error while starting RPC server:", err)
	}
	
	CurrNode.Addr = fmt.Sprintf("localhost:%s", port)
	CurrNode.Next = nil

	GetPeers()
	
	go CheckLeaderAlive()

	http.Serve(lis, nil) 
	
}
