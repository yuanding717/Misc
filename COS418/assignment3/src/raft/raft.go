package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"bytes"
	"encoding/gob"
	"labrpc"
	"math"
	"math/rand"
	"sync"
	"time"
)

const COMMON_TIMEOUT_BASE = 50
const ELECTION_TIMEOUT_BASE = 150
const ELECTION_TIMEOUT_INTERVAL = 150
const (
	FOLLOWER = iota
	CANDIDATE
	LEADER
)

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make().
//
type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex
	peers     []*labrpc.ClientEnd
	persister *Persister
	me        int // index into peers[]

	// TODO Your data here.
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	// persistent state
	currentTerm int
	votedFor    int // peer index
	log         []Entry

	// volatile state
	commitIndex int
	lastApplied int

	// volatile state on leaders
	nextIndex  []int
	matchIndex []int

	// util field
	applyCh           chan ApplyMsg
	state             int
	electionStartTime time.Time
	electionTimeout   time.Duration
	isKilled          bool
}

type Entry struct {
	Term    int
	Command interface{}
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	// var term int
	// var isleader bool
	// TODO Your code here.
	return rf.currentTerm, rf.state == LEADER
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// TODO Your code here.
	// Example:
	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	e.Encode(rf.log)
	data := w.Bytes()
	rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	// TODO Your code here.
	// Example:
	r := bytes.NewBuffer(data)
	d := gob.NewDecoder(r)
	d.Decode(&rf.currentTerm)
	d.Decode(&rf.votedFor)
	d.Decode(&rf.log)
	DPrintf("Server %v read state term: %v, votedFor: %v, log: %v", rf.me, rf.currentTerm, rf.votedFor, rf.log)
}

//
// example RequestVote RPC arguments structure.
//
type RequestVoteArgs struct {
	// TODO Your data here.
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

//
// example RequestVote RPC reply structure.
//
type RequestVoteReply struct {
	// TODO Your data here.
	Term        int
	VoteGranted bool
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) {
	// TODO Your code here.
	rf.resetElectionTimer()

	reply.Term = rf.currentTerm
	reply.VoteGranted = false
	if (args.Term > rf.currentTerm || args.Term == rf.currentTerm && rf.state != LEADER) &&
		(rf.votedFor == -1 || rf.votedFor == args.CandidateId) &&
		(rf.log[len(rf.log)-1].Term < args.LastLogTerm ||
			(rf.log[len(rf.log)-1].Term == args.LastLogTerm && len(rf.log)-1 <= args.LastLogIndex)) {
		reply.VoteGranted = true
	}
	// DPrintf("Server %v replying requestVote from %v, granted: %v, votedFor: %v",
	// rf.me, args.CandidateId, reply.VoteGranted, rf.votedFor)
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// returns true if labrpc says the RPC was delivered.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Entry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.resetElectionTimer()
	reply.Term = rf.currentTerm
	if args.Term > rf.currentTerm {
		rf.updateTerm(args.Term)
	}

	if args.Term < rf.currentTerm || args.PrevLogIndex >= len(rf.log) ||
		rf.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		reply.Success = false
		prevTerm := -1
		if args.PrevLogIndex < len(rf.log) {
			prevTerm = rf.log[args.PrevLogIndex].Term
		}
		DPrintf("SERVER %v Failed AppendEntries from %v, args.Term %v, current term = %v, args.PrevLogIndex %v, len(rf.log) %v, args.PrevTerm %v, log PrevTerm: %v",
			rf.me, args.LeaderId, args.Term, rf.currentTerm, args.PrevLogIndex, len(rf.log), args.PrevLogTerm, prevTerm)
		return
	}
	if len(args.Entries) > 0 {
		if args.PrevLogIndex+1 < len(rf.log) {
			rf.log = rf.log[:args.PrevLogIndex+1]
		}
		rf.log = append(rf.log, args.Entries...)
		rf.persist()
	}
	if args.LeaderCommit > rf.commitIndex {
		if args.LeaderCommit < len(rf.log)-1 {
			rf.commitIndex = args.LeaderCommit
		} else {
			rf.commitIndex = len(rf.log) - 1
		}
	}

	for rf.commitIndex > rf.lastApplied {
		rf.lastApplied++
		rf.applyCh <- ApplyMsg{rf.lastApplied, rf.log[rf.lastApplied].Command, false, nil}
	}

	reply.Success = true
}

func (rf *Raft) doAppendEntries() {
	inprocess := make([]bool, len(rf.peers))
	for {
		wp := false
		for i, _ := range rf.peers {
			if rf.state != LEADER {
				DPrintf("Server %v is no longer the leader", rf.me)
				return
			}
			if inprocess[i] {
				continue
			}
			inprocess[i] = true

			rf.mu.Lock()
			if len(rf.log) < rf.nextIndex[i] {
				DPrintf("LEADER %v: Exception  - nextIndex %v of server %v is beyond log length",
					rf.me, rf.nextIndex[i], i)
				rf.adjustForConsistency(i)
			}
			args := AppendEntriesArgs{
				rf.currentTerm,
				rf.me,
				rf.nextIndex[i] - 1,
				rf.log[rf.nextIndex[i]-1].Term,
				make([]Entry, len(rf.log)-rf.nextIndex[i]),
				rf.commitIndex,
			}
			if len(rf.log) > rf.nextIndex[i] {
				wp = true
				copy(args.Entries, rf.log[rf.nextIndex[i]:len(rf.log)])
				DPrintf("LEADER %v: Requesting appendEntries to %v, len log: %v, rf.nextIndex: %v, PrevLogIndex: %v, PrevLogTerm: %v, new entries: %v",
					rf.me, i, len(rf.log), rf.nextIndex[i], args.PrevLogIndex, args.PrevLogTerm, len(args.Entries))
			}
			rf.mu.Unlock()

			go func(server int, args AppendEntriesArgs) {
				reply := AppendEntriesReply{}
				ok := rf.sendAppendEntries(server, args, &reply)
				if !ok {
					DPrintf("LEADER %v failed to connect to server %v", rf.me, server)
				} else if reply.Success {
					if len(args.Entries) > 0 {
						rf.mu.Lock()
						rf.nextIndex[server] += len(args.Entries)
						rf.matchIndex[server] = rf.nextIndex[server] - 1

						count := make(map[int]int)
						for _, j := range rf.matchIndex {
							if j >= len(rf.log) {
								DPrintf("LEADER %v: Exception - matchIndex %v of server %v is beyond log length",
									rf.me, rf.matchIndex[j], j)
							} else {
								if rf.log[j].Term == rf.currentTerm && j > rf.commitIndex {
									count[j]++
								}
							}
						}
						for k, v := range count {
							if v > len(rf.peers)/2 && k > rf.commitIndex {
								rf.commitIndex = k
							}
						}
						DPrintf("LEADER %v: STATS nextIndex: %v, matchIndex: %v, commitIndex: %v, log: %v, term: %v",
							rf.me, rf.nextIndex, rf.matchIndex, rf.commitIndex, rf.log, rf.currentTerm)

						rf.mu.Unlock()
					}
				} else if reply.Term > rf.currentTerm { // fail for leader change
					rf.updateTerm(reply.Term)
				} else {
					prevIndex := rf.nextIndex[server]
					rf.adjustForConsistency(server)
					DPrintf("LEADER %v: AppendEntries to %v failed for log inconsistency with nextIndex %v, fallback to %v",
						rf.me, server, prevIndex, rf.nextIndex[server])
				}
				inprocess[server] = false
			}(i, args)
		}
		if !wp {
			time.Sleep(time.Duration(rand.Int() % COMMON_TIMEOUT_BASE * int(math.Pow10(6))))
		}

	}
}

func (rf *Raft) adjustForConsistency(server int) {
	failedTerm := rf.log[rf.nextIndex[server]-1].Term
	rf.mu.Lock()
	for j := rf.nextIndex[server] - 1; j >= 0; j-- {
		if rf.log[j].Term != failedTerm {
			rf.nextIndex[server] = j + 1
			break
		}
	}
	rf.mu.Unlock()
}

func (rf *Raft) sendAppendEntries(server int, args AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) doElection() {
	rf.resetElectionTimer()
	rf.state = CANDIDATE
	rf.votedFor = rf.me

	args := RequestVoteArgs{
		rf.currentTerm + 1,
		rf.me,
		len(rf.log) - 1,
		rf.log[len(rf.log)-1].Term}

	voteChan := make(chan RequestVoteReply, len(rf.peers))
	for i, _ := range rf.peers {
		go func(leader *Raft, args RequestVoteArgs, server int) {
			reply := RequestVoteReply{}
			ok := leader.sendRequestVote(server, args, &reply)
			if !ok {
				// DPrintf("Sending requestVote from server %v to server %v failed", leader.me, server)
			} else {
				voteChan <- reply
			}
		}(rf, args, i)
	}

	votes := 0
	votesGranted := 0
loop:
	for {
		select {
		case <-rf.getTimeoutChan(rf.electionTimeout):
			DPrintf("Server %v Timeout waiting for votes, got %v votes and %v granted",
				rf.me, votes, votesGranted)
			rf.checkAndUpdateElectionResult(votesGranted)
			break loop
		case reply := <-voteChan:
			if reply.Term > rf.currentTerm { // discover leader or new term
				DPrintf("SERVER %v discovers new leader or term, election stopped", rf.me)
				rf.updateTerm(reply.Term)
				break loop
			}
			votes++
			if reply.VoteGranted {
				votesGranted++
			}
			if votes == len(rf.peers) {
				rf.checkAndUpdateElectionResult(votesGranted)
				break loop
			}
		}
	}
}

func (rf *Raft) getTimeoutChan(timeout time.Duration) chan bool {
	timeoutChan := make(chan bool, 1)
	go func() {
		time.Sleep(timeout)
		timeoutChan <- true
	}()
	return timeoutChan
}

func (rf *Raft) resetElectionTimer() {
	rf.electionStartTime = time.Now()
	rf.electionTimeout = time.Duration((ELECTION_TIMEOUT_BASE + rand.Int()%ELECTION_TIMEOUT_INTERVAL) * int(math.Pow10(6)))
}

func (rf *Raft) updateTerm(term int) {
	rf.mu.Lock()
	rf.currentTerm = term
	rf.state = FOLLOWER
	rf.votedFor = -1
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))
	for i := 0; i < len(rf.nextIndex); i++ {
		rf.nextIndex[i] = len(rf.log)
		rf.matchIndex[i] = 0
	}
	rf.mu.Unlock()
	rf.persist()
}

func (rf *Raft) checkAndUpdateElectionResult(votesGranted int) {
	if votesGranted > len(rf.peers)/2 {
		rf.state = LEADER
		rf.currentTerm++
		rf.votedFor = -1
		rf.nextIndex = make([]int, len(rf.peers))
		rf.matchIndex = make([]int, len(rf.peers))
		for i := 0; i < len(rf.nextIndex); i++ {
			rf.nextIndex[i] = len(rf.log)
			rf.matchIndex[i] = 0
		}
		rf.matchIndex[rf.me] = len(rf.log) - 1
		DPrintf("SERVER %v is the leader for term %v, nextIndex: %v, matchIndex: %v",
			rf.me, rf.currentTerm, rf.nextIndex, rf.matchIndex)
		rf.persist()
	} else {
		rf.updateTerm(rf.currentTerm)
	}
}

func (rf *Raft) doWork() {
	for {
		if rf.isKilled {
			break
		}
		if rf.state == LEADER {
			rf.doAppendEntries()
		} else if time.Now().Sub(rf.electionStartTime) > rf.electionTimeout {
			rf.doElection()
		} else {
			time.Sleep(time.Duration(rand.Int() % COMMON_TIMEOUT_BASE * int(math.Pow10(6))))
		}
	}
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	index := len(rf.log)
	term := rf.currentTerm
	isLeader := rf.state == LEADER
	if isLeader {
		entry := Entry{rf.currentTerm, command}
		rf.log = append(rf.log, entry)
		rf.nextIndex[rf.me] = len(rf.log)
		rf.matchIndex[rf.me] = len(rf.log) - 1
		rf.persist()

		DPrintf("LEADER %v: Appending new log entry, command: %v, new len: %v, current term %v",
			rf.me, command, len(rf.log), rf.currentTerm)
	}

	return index, term, isLeader
}

//
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// TODO Your code here, if desired.
	rf.isKilled = true
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any.
// applyCh is a channel on which the -------
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// TODO Your initialization code here.
	rf.applyCh = applyCh
	rf.state = FOLLOWER
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.log = make([]Entry, 1)
	rf.log[0] = Entry{-1, nil}
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.resetElectionTimer()
	rf.isKilled = false

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	go rf.doWork()

	return rf
}
