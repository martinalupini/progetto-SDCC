package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/http"
	"time"
)

type Peer string

// struct used in RPC //
type Neighbours struct {
	Prev Node
	Next Node
}

type Node struct{
	ID int
	Addr string
	Next *Node
	Prev *Node
} 

type ElectionMsg struct {
	Candidate Node
	From Node

}

// variables of the registry //
var CurrNode Node
var Leader = Node{}
var iamLeader = false
var election = false


// RPC ///////////////////////

func (p *Peer) UpdateNext(next Node, reply *int) error {

	CurrNode.Next = &next
	log.Printf("New next is node %d address %s", CurrNode.Next.ID, CurrNode.Next.Addr)
	
	return nil
}



func (p *Peer) ElectionLeader(msg ElectionMsg, reply *int) error {

	if msg.Candidate.ID == CurrNode.ID {
		log.Printf("ELECTION: I'm the leader")
		iamLeader = true
		election = false
		*reply = 1
		NotifyLeader(msg.From)
	}else if msg.Candidate.ID > CurrNode.ID {
		election = true
		log.Printf("ELECTION: Node %d wants to be candidate. Sending message to next node", msg.Candidate.ID)
		Election(msg.Candidate)
	}else{
		election = true
		log.Printf("ELECTION: Node %d has lower ID then mine. Running for election :)", msg.Candidate.ID)
		Election(CurrNode)
	}
	
	
	return nil
}


func (p *Peer) NewLeader(leader Node, reply *int) error {
	iamLeader = false
	Leader = leader
	log.Printf("ELECTION FINISHED: New leader is node %d", Leader.ID) 
	
	return nil

}

//////////////////////////////

func NotifyLeader(except Node){
	var next *Node = CurrNode.Next
	var reply int

	if next == nil{
		return
	}
	
	
	for next.ID != CurrNode.ID && next.ID != except.ID {
	
		if IsAlive(CurrNode.Next.Addr) < 0 {
			log.Fatal("Node %d not responding, passing to next...", next.ID)
			next = next.Next
			continue 
		}
		
		log.Printf("ELECTION FINISHED: Informing node %d I'm the leader", next.ID)
		client, err := rpc.DialHTTP("tcp", CurrNode.Next.Addr)
		if err != nil {
			log.Fatal("Dial error", next.ID)
		}
	
		err = client.Call("Peer.NewLeader", CurrNode, &reply)
		if err != nil {
			log.Fatal("Error while calling RPC:", err)
		}
		
		next = next.Next
		client.Close()
	
	}
	
	return

}


func Election(leader Node){
	var next *Node = CurrNode.Next
	var reply int
	var msg ElectionMsg

	if next == nil{
		iamLeader = true
		Leader = CurrNode
		log.Printf("ELECTION: No one in the network. I'am the leader")
		return
	}
	
	
	for next.ID != CurrNode.ID {
		if IsAlive(CurrNode.Next.Addr) < 0 {
			log.Fatal("Node %d not responding, passing to next...", next.ID)
			next = next.Next
			continue 
		}
		
		msg.Candidate = leader
		msg.From = CurrNode
		
		if election == false {
			log.Printf("STARTING ELECTION: Sending message to node %d", next.ID)
		} else {
			log.Printf("ELECTION: Sending message to node %d", next.ID)
		}
		
		client, err := rpc.DialHTTP("tcp", CurrNode.Next.Addr)
		if err != nil {
			log.Fatal("Dial error", next.ID)
		}
		
		err = client.Call("Peer.ElectionLeader", msg, &reply)
		if err != nil {
			log.Fatal("Error while calling RPC:", err)
		}
		
		if(reply == 1){
			iamLeader = false
			Leader = *next
			log.Printf("ELECTION FINISHED: New leader is %d", Leader.ID)
		}
		
		client.Close()
		election = true
		return
	}
	
	if next.ID == CurrNode.ID {
		log.Printf("ELECTION: No node is working in the network. Waiting for leader recovery")
		return
	}


}



func CheckLeaderAlive(){
	var ret int
	if Leader == (Node{}) {
		Election(CurrNode)
		
	}
	
	for {
		if iamLeader == false && election == false {
			ret = IsAlive(Leader.Addr)
			if ret < 0 {
				log.Printf("Leader is not responding")
				Election(CurrNode)
			}
		}
	}
}

func IsAlive(addr string) int {
	ret :=0
	client, err := net.DialTimeout("tcp", addr, 5*time.Second )
	if err != nil {
		ret = -1
	
	}
	client.Close()
	
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


func GetPeers() {

	serviceRegistry := "localhost:1234"
	var reply Neighbours
	
	//connection to service registry 
	log.Printf("Connecting to service registry")
	client, err := rpc.DialHTTP("tcp", serviceRegistry)
	if err != nil {
		log.Fatal("Error while connecting to registry server:", err)
	}
	
	err = client.Call("ServiceRegistry.AddNode", &CurrNode, &reply)
	if err != nil {
		log.Fatal("Error while calling RPC:", err)
	}
	
	client.Close()
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

}

func main() {

	peer := new(Peer)
	rpc.Register(peer)
	rpc.HandleHTTP()
	// Listen for incoming TCP packets on specified port
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("Error while starting RPC server:", err)
	}
	
	CurrNode.ID=lis.Addr().(*net.TCPAddr).Port
	CurrNode.Addr = fmt.Sprintf("localhost:%d", CurrNode.ID)
	CurrNode.Next = nil
	
	log.Printf("RPC server listens on port %s", CurrNode.Addr)

	GetPeers()
	
	go CheckLeaderAlive()
	
	if CurrNode.Prev != nil {
		NotifyPrev()
	}


	http.Serve(lis, nil) 
	
}
