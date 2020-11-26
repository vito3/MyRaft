package main

import (
	config "../config"
	KV "../grpc/mykv"
	"../myraft"
	Per "../persister"
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)




type KVServer struct {
	mu      sync.Mutex
	me      int
	rf      *myraft.Raft
	//applyCh chan int
	applyCh chan config.ApplyMsg
	dead    int32 // set by Kill()
	maxraftstate int // snapshot if log grows this big

	// Your definitions here.
	cid2Seq map[int64]int64 //clientId -> maxSeq
	db map[string]string // key -> value
	chMap map[int]chan config.Op // index -> op
	persist  *Per.Persister

}


func (kv *KVServer) PutAppend(ctx context.Context,args *KV.PutAppendArgs) ( *KV.PutAppendReply, error){

	//time.Sleep(time.Second)
	reply := &KV.PutAppendReply{}
	_ , reply.IsLeader = kv.rf.GetState()
	//reply.IsLeader = false;
	if !reply.IsLeader{
		return reply, nil
	}

	oringalOp := config.Op{args.Op, args.Key,args.Value, args.Id, args.Seq}
	index, _, isLeader := kv.rf.Start(oringalOp)
	if !isLeader {
		fmt.Println("Leader Changed !")
		reply.IsLeader = false
		return reply, nil
	}

	apply := <- kv.applyCh
	fmt.Println("apply ", apply)
	fmt.Println("index", index)

	return reply, nil

}



func (kv *KVServer) Get(ctx context.Context, args *KV.GetArgs) ( *KV.GetReply, error){
	
	reply := &KV.GetReply{}
	_ , reply.IsLeader = kv.rf.GetState()
	//reply.IsLeader = false;
	if !reply.IsLeader{
		return reply, nil
	}

	//fmt.Println()


	oringalOp := config.Op{"Get", args.Key,"" , 0, 0}
	_, _, isLeader  := kv.rf.Start(oringalOp)
	if !isLeader {
		return reply, nil
	}
	//fmt.Println(index)

	reply.IsLeader = true
	//kv.mu.Lock()
	//fmt.Println("Asdsada")
	reply.Value = string( kv.persist.Get(args.Key) )
	//fmt.Println("Asdsada")

	//kv.mu.Unlock()
	return reply, nil
}


func (kv *KVServer) equal(a config.Op, b config.Op) bool  {
	return (a.Option == b.Option && a.Key == b.Key &&a.Value == b.Value) 
}



func (kv *KVServer) RegisterServer(address string)  {
	// Register Server 
	for{

		lis, err := net.Listen("tcp", address)
		fmt.Println("myserver", address)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		KV.RegisterKVServer(s, kv )
		// Register reflection service on gRPC server.
		reflection.Register(s)
		if err := s.Serve(lis); err != nil {
			fmt.Println("failed to serve: %v", err)
		}
		
	}
	
}


func main()  {

	var add = flag.String("address", "", "Input Your address")
	var mems = flag.String("members", "", "Input Your follower")
	flag.Parse()

	server := KVServer{}

	// Local address	
	address := *add

	persist := &Per.Persister{}
	persist.Init("../db/"+address)

	// Members's address
	members := strings.Split( *mems, ",")
	fmt.Println("add+mem", address, members)
	
	//for i := 0; i <= int (address[ len(address) - 1] - '0'); i++{
	//server.applyCh = make(chan int, 1)
	//}

	//server.applyCh = make(chan int, 1) // 原代码只保留了此句
	//fmt.Println("server.applyCh:", server.applyCh)
	server.applyCh = make(chan config.ApplyMsg)
	go server.RegisterServer(address+"1")
	server.rf = myraft.MakeRaft(address , members ,persist, &server.mu, server.applyCh)
	server.chMap = make(map[int]chan config.Op)
	server.cid2Seq = make(map[int64]int64)
	server.db = make(map[string]string)
	server.persist  = persist




	time.Sleep(time.Second*1200)

}
