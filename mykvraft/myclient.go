package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	KV "../grpc/mykv"
	//"strconv"
	"strings"
	//"google.golang.org/grpc/reflection"
	crand "crypto/rand"
	"math/big"
	"strconv"
	//Per "../persister"

	//"math/rand"
)
type Clerk struct {
	mu      sync.Mutex
	servers []string
	// You will have to modify this struct.
	leaderId int
	id int64
	seq int64
}



func makeSeed() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := crand.Int(crand.Reader, max)
	x := bigx.Int64()
	return x
}


func MakeClerk(servers []string) *Clerk {
	ck := new(Clerk)

	ck.servers = make([]string, len(servers))
	serversNum := len(servers)
	for i := 0; i < serversNum; i++ {
		ck.servers[i] = servers[i] + "1"
	}

	ck.id = makeSeed()
	ck.seq = 0
	return ck
}

func (ck *Clerk) Get(key string) string {
	args := &KV.GetArgs{Key:key}
	id := ck.leaderId
	for {
		//fmt.Println(id)
		reply, ok := ck.getValue(ck.servers[id], args)
		//fmt.Println(reply.IsLeader)
		if (ok && reply.IsLeader){
			ck.leaderId = id
			return reply.Value
		}else{
			//fmt.Println("can not connect ", ck.servers[id], "or it's not leader")
		}
		id = (id + 1) % len(ck.servers) 

	} 
}

func (ck *Clerk) getValue(address string , args  *KV.GetArgs) (*KV.GetReply, bool){
	// Initialize Client
	conn, err := grpc.Dial( address , grpc.WithInsecure() ) //,grpc.WithBlock())
	if err != nil {
		log.Printf("did not connect: %v", err)
		return nil, false
	}
	defer conn.Close()
	client := KV.NewKVClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 1)
	defer cancel()
	reply, err := client.Get(ctx,args)
	if err != nil {
		return  nil, false
		log.Printf(" getValue could not greet: %v", err)
	}
	return reply, true
}



func (ck *Clerk) Put(pid string, key string, value string) bool {
	//fmt.Println("###### ", pid, "Enter Client Put() ######")
	args := &KV.PutAppendArgs{Key:key,Value:value,Op:"Put", Id:ck.id, Seq:ck.seq }
	//fmt.Printf("PUT - ck.leadepid:%d ck_id:%d seq:%d\n", ck.leaderId, ck.id, ck.seq)
	fmt.Println("PUT-", pid, "Info-", args)
	id := ck.leaderId //初识为0
	//fmt.Println(pid, args.Seq, id) // , 0, 0
	ck.seq++
	for {
		reply, ok := ck.putAppendValue(pid, ck.servers[id], args)
		if ok && reply.IsLeader {
			ck.leaderId = id
			fmt.Println("PUT-", pid, "successfully find leader ", id)
			return true
		}else{
			if !ok {
				fmt.Println("PUT-", pid, "putAppendValue() return false")
			} else { //reply为空，不可直接取 reply.IsLeader
				fmt.Println("PUT-", pid, "find wrong leader")
			}
		}
		id = (id + 1) % len(ck.servers) 
	} 
}

func (ck *Clerk) Append(aid string, key string, value string) bool {
	fmt.Println("###### ", aid, "Enter Client Append() ######")
	args := &KV.PutAppendArgs{Key:key,Value:value,Op:"Append", Id:ck.id, Seq:ck.seq }
	fmt.Println("Append-", aid, "Info-", args)
	id := ck.leaderId
	for {
		reply, ok := ck.putAppendValue(aid, ck.servers[id], args)
		if ok && reply.IsLeader {
			ck.leaderId = id
			fmt.Println("Append-", aid, "successfully find leader ", id)
			return true
		}else{
			if !ok {
				fmt.Println("Append-", aid, "putAppendValue() return false")
			} else {
				fmt.Println("Append-", aid, "find wrong leader")
			}
		}
		id = (id + 1) % len(ck.servers) 
	} 
}


func (ck *Clerk) putAppendValue(rid string , address string , args  *KV.PutAppendArgs) (*KV.PutAppendReply, bool){
	//fmt.Println("###### ", rid, "Enter Client putAppendValue() ######")
	conn, err := grpc.Dial( address , grpc.WithInsecure() )//,grpc.WithBlock())
	if err != nil {
		//fmt.Println("PutAppend-", rid, "Dial fail: ", err)
		return  nil, false
	}
	defer conn.Close()
	client := KV.NewKVClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 1)
	defer cancel()
	//reply, err := client.PutAppend(context.Background(), args)
	reply, err := client.PutAppend(ctx, args)
	if err != nil {
		fmt.Println("PutAppend-", rid, "PutAppend() fail", err)
		return nil, false
	}
	return reply, true
}

var count int32  = 0

//func request(num int, servers []string)  {
func (ck *Clerk) request(num int)  { //第num个client发起请求
	//fmt.Println("###### Request Client: ", num, "######")
	/*ck := Clerk{}
	ck.servers = make([]string, len(servers))
	for i:= 0; i < len(servers); i++ {
		ck.servers[i] = servers[i] + "1"
	}*/
	fmt.Println(num, ck.servers)

	requestNum := 2
 	for i := 0; i < requestNum ; i++ {
		//rand.Seed(time.Now().UnixNano())
		//key := "key" + strconv.Itoa(rand.Intn(100000))
		//value := "value"+ strconv.Itoa(rand.Intn(100000))
 		key := "key-" + strconv.Itoa(num) + "-" + strconv.Itoa(i)
		value := "value-" + strconv.Itoa(num) + "-" + strconv.Itoa(i)
		fmt.Println("Client ", num,  ", try to put: ", "[", key, "]", "- [", value, "]")
		rid := strconv.Itoa(num) + "-" + strconv.Itoa(i)
		ck.Put(rid, key, value)
		atomic.AddInt32(&count,1)
	}
}

func main()  {

	var ser = flag.String("servers", "", "Input Your follower")
	flag.Parse()
	servers := strings.Split( *ser, ",")

	clientNum := 1
	for i := 0; i < clientNum ; i++ {
		ck := MakeClerk(servers)
		go ck.request(i)
	}

	time.Sleep(time.Second*1200)

}