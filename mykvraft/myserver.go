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

	fmt.Println("###### Enter Server PutAppend Handler ######")
	reply := &KV.PutAppendReply{}
	_ , reply.IsLeader = kv.rf.GetState()
	//reply.IsLeader = false
	if !reply.IsLeader{
		fmt.Println("PutAppend-Handler: wrong leader")
		return reply, nil
	}
	originOp := config.Op{args.Op, args.Key,args.Value, args.Id, args.Seq}
	fmt.Println("PutAppend-Handler: Op-", args)
	index, _, isLeader := kv.rf.Start(originOp)
	if !isLeader {
		fmt.Println("PutAppend-Handler: Leader Changed !")
		reply.IsLeader = false
		return reply, nil
	}
	//apply := <- kv.applyCh
	ch := kv.putIfAbsent(int(index))
	//op := <- ch
	op := beNotified(ch) //实际得到的？,ch可能为空，
	if equalOp(op, originOp) {
		reply.IsLeader = true
		kv.mu.Lock()
		reply.Success = true
		kv.mu.Unlock()
	}
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


func  equalOp(a config.Op, b config.Op) bool  {
	return (a.Option == b.Option && a.Key == b.Key &&a.Value == b.Value) 
}

func (kv *KVServer) putIfAbsent(idx int) chan config.Op{
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if _,ok := kv.chMap[idx]; !ok {
		kv.chMap[idx] = make(chan config.Op, 1)
	}
	return kv.chMap[idx]
}

func beNotified(ch chan config.Op) config.Op {
	select {
	case op := <- ch:
		return op
	case <- time.After(time.Second):
		return config.Op{}
	}
}


func (kv *KVServer) RegisterServer(address string)  {
	// Register Server 
	for{

		lis, err := net.Listen("tcp", address)
		//fmt.Println("myserver", address)
		if err != nil {
			fmt.Println("failed to listen:", err)
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

	//server.applyCh = make(chan int, 1) // 原代码只保留了此句
	//fmt.Println("server.applyCh:", server.applyCh)
	server.applyCh = make(chan config.ApplyMsg)
	go server.RegisterServer(address+"1") //50001用于外部cs间的gRPC
	server.rf = myraft.MakeRaft(address, members, persist, &server.mu, server.applyCh)

	server.chMap = make(map[int]chan config.Op)
	server.cid2Seq = make(map[int64]int64)
	server.db = make(map[string]string)
	server.persist  = persist

	go func() {
		for  {
			applyMsg := <- server.applyCh
			op := applyMsg.Command.(config.Op)
			server.mu.Lock()
			maxSeq,found := server.cid2Seq[op.Id]
			if !found || op.Seq>maxSeq {//未出现或者该客户端看到的最大序列号小于op的序列号
				if found {
					fmt.Println("Found! p Info: ", op.Id, maxSeq, op.Seq)
				} else {
					fmt.Println(op.Id, "Not found")
				}
				switch op.Option {
				case "Put":
					fmt.Println("Put Op ", op.Key, op.Value)
					server.db[op.Key] = op.Value
				case "Append":
					fmt.Println("Append Op ", op.Key, op.Value)
					server.db[op.Key] += op.Value
				}
				server.cid2Seq[op.Id] = op.Seq //每个客户机看到的最大的序列号
			}
			server.mu.Unlock()
			index := applyMsg.CommandIndex
			ch := server.putIfAbsent(int(index))
			ch <- op //通道
		}
	}()

	time.Sleep(time.Second*1200)

}
