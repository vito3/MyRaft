package georaft

import (
	"log"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sync"
	"google.golang.org/grpc/reflection"
	RPC "../grpc/georaft"
	"fmt"
	"time"
	"../labgob"
	"bytes"
)

/* 	
type Log struct {
    Term    int32         "term when entry was received by leader"
    Command interface{} "command for state machine,"
} 

type ApplyMsg struct {
    CommandValid bool
    Command      interface{}
    CommandIndex int
}

*/

type Secretary struct {
    mu        sync.Mutex          // Lock to protect shared access to this peer's state

	me int32
    //Persistent state on all servers:(Updated on stable storage before responding to RPCs)
    currentTerm int32    "latest term server has seen (initialized to 0 increases monotonically)"
    log         []Log  "log entries;(first index is 1)"

    //Volatile state on all servers:
    commitIndex int32    "index of highest log entry known to be committed (initialized to 0, increases monotonically)"
    lastApplied int32    "index of highest log entry applied to state machine (initialized to 0, increases monotonically)"

    //Volatile state on leaders：(Reinitialized after election)
    nextIndex   []int32  "for each server,index of the next log entry to send to that server"
    matchIndex  []int32  "for each server,index of highest log entry known to be replicated on server(initialized to 0, im)"

	address string
	members []string

}




//Helper function
func send(ch chan bool) {
    select {
    case <-ch: //if already set, consume it then resent to avoid block
    default:
    }
    ch <- true
}

func (se *Secretary) getPrevLogIdx(i int) int32 {
    return se.nextIndex[i] - 1
}

func (se *Secretary) getPrevLogTerm(i int) int32 {
    prevLogIdx := se.getPrevLogIdx(i)
    if prevLogIdx < 0 {
        return -1
    }
    return se.log[prevLogIdx].Term
}

func (se *Secretary) getLastLogIdx() int32 {
    return int32(len(se.log) - 1)
}

func (se *Secretary) getLastLogTerm() int32 {
    idx := se.getLastLogIdx()
    if idx < 0 {
        return -1
    }
    return se.log[idx].Term
}


 
func (se *Secretary)L2SAppendEntries(ctx context.Context, args *RPC.L2SAppendEntriesArgs) (*RPC.L2SAppendEntriesReply, error) {

	r := bytes.NewBuffer(args.Log)
    d := labgob.NewDecoder(r)
	var log []Log 
	d.Decode(&log) 	
    
	fmt.Println("AppendEntries CALL",log )

	reply := &RPC.L2SAppendEntriesReply{}
	
	return reply, nil
}


func (se *Secretary) RegisterServer(address string)  {
	// Register Server 
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	RPC.RegisterSecretaryServer(s, se )
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}



func (se *Secretary) startAppendLog() {
	
	fmt.Println("startAppendLog ")
	
	//idx := 0
	//appendLog := append(make([]Log,0),rf.log[rf.nextIndex[idx]:]...)
	appendLog := make([]Log,3)
	appendLog[0].Term = 0
	appendLog[0].Command =  "sdfdsfsdfdsfdsfdsfdsfsfsfsfsdfdsfdsfdsfdsfsfdsfdsfs"
	appendLog[1].Term = 1
	appendLog[1].Command =  "sdfdsfsdfdsfdsfdsfdsfsfsfsfsdfdsfdsfdsfdsfsfdsfdsfs" 
	appendLog[2].Term = 2
	appendLog[2].Command =  "sdfdsfsdfdsfdsfdsfdsfsfsfsfsdfdsfdsfdsfdsfsfdsfdsfs"
	
	w := new(bytes.Buffer)
    e := labgob.NewEncoder(w)
    e.Encode(appendLog)

    data := w.Bytes()

	args := RPC.AppendEntriesArgs{
		Term: se.currentTerm,
		LeaderId: se.me,
		PrevLogIndex: 1,  //se.getPrevLogIdx(idx),
		PrevLogTerm: 1, //se.getPrevLogTerm(idx),
		Log: data,
		LeaderCommit: se.commitIndex,
	}
	for i := 0; i < len(se.members); i++{
		if (se.address == se.members[i]) {
			continue	

		}
		fmt.Println("CALL ADDRESS: ", se.members[i])
		go se.S2OsendAppendEntries(se.members[i], &args)
	}

	
}




func (se *Secretary) init (add string) {


    se.currentTerm = 0
    se.log = make([]Log,1) //(first index is 1)

    se.commitIndex = 0
    se.lastApplied = 0
	se.address = add;	

	fmt.Println("se.address ", se.address)

	go  se.RegisterServer(se.address)

	go func ()  {
		for {
			se.startAppendLog()
			time.Sleep(time.Second )
	
		}
	}()



	

	
}



func (se *Secretary) S2FsendAppendEntries(address string , args  *RPC.AppendEntriesArgs){

	fmt.Println("StartAppendEntries")

	// Initialize Client
	conn, err := grpc.Dial( address , grpc.WithInsecure(),grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := RPC.NewObserverClient(conn)


	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	//args := &RPC.L2SAppendEntriesArgs{}
	r, err := client.AppendEntries(ctx,args)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Append reply: %s", r)
	//fmt.Println("Append name is ABC")
}




func (se *Secretary) S2OsendAppendEntries(address string , args  *RPC.AppendEntriesArgs){

	fmt.Println("StartAppendEntries")

	// Initialize Client
	conn, err := grpc.Dial( address , grpc.WithInsecure(),grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := RPC.NewObserverClient(conn)


	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	//args := &RPC.L2SAppendEntriesArgs{}
	r, err := client.AppendEntries(ctx,args)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Append reply: %s", r)
	//fmt.Println("Append name is ABC")
}





func MakeSecretary(Args []string) *Secretary {
	se := &Secretary{}


  	if (len(Args) > 0){
		se.members = make([]string, len(Args) )
		for i:= 0; i < len(Args) ; i++{
			se.members[i] = Args[i]
			fmt.Printf(se.members[i])
		}
	}  

	//fmt.Println()
	se.init(se.members[0])
	return se



} 


