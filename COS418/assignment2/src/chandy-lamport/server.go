package chandy_lamport

import "log"

// The main participant of the distributed snapshot protocol.
// Servers exchange token messages and marker messages among each other.
// Token messages represent the transfer of tokens from one server to another.
// Marker messages represent the progress of the snapshot process. The bulk of
// the distributed protocol is implemented in `HandlePacket` and `StartSnapshot`.
type Server struct {
	Id            string
	Tokens        int
	sim           *Simulator
	outboundLinks map[string]*Link // key = link.dest
	inboundLinks  map[string]*Link // key = link.src
	// TODO: ADD MORE FIELDS HERE
	SnapshotStates  *SyncMap
	markersReceived map[int]map[string]bool
}

// A unidirectional communication channel between two servers
// Each link contains an event queue (as opposed to a packet queue)
type Link struct {
	src    string
	dest   string
	events *Queue
}

func NewServer(id string, tokens int, sim *Simulator) *Server {
	return &Server{
		id,
		tokens,
		sim,
		make(map[string]*Link),
		make(map[string]*Link),

		NewSyncMap(),
		make(map[int]map[string]bool),
	}
}

// Add a unidirectional link to the destination server
func (server *Server) AddOutboundLink(dest *Server) {
	if server == dest {
		return
	}
	l := Link{server.Id, dest.Id, NewQueue()}
	server.outboundLinks[dest.Id] = &l
	dest.inboundLinks[server.Id] = &l
}

// Send a message on all of the server's outbound links
func (server *Server) SendToNeighbors(message interface{}) {
	for _, serverId := range getSortedKeys(server.outboundLinks) {
		link := server.outboundLinks[serverId]
		server.sim.logger.RecordEvent(
			server,
			SentMessageEvent{server.Id, link.dest, message})
		link.events.Push(SendMessageEvent{
			server.Id,
			link.dest,
			message,
			server.sim.GetReceiveTime()})
	}
}

// Send a number of tokens to a neighbor attached to this server
func (server *Server) SendTokens(numTokens int, dest string) {
	if server.Tokens < numTokens {
		log.Fatalf("Server %v attempted to send %v tokens when it only has %v\n",
			server.Id, numTokens, server.Tokens)
	}
	message := TokenMessage{numTokens}
	server.sim.logger.RecordEvent(server, SentMessageEvent{server.Id, dest, message})
	// Update local state before sending the tokens
	server.Tokens -= numTokens
	link, ok := server.outboundLinks[dest]
	if !ok {
		log.Fatalf("Unknown dest ID %v from server %v\n", dest, server.Id)
	}
	link.events.Push(SendMessageEvent{
		server.Id,
		dest,
		message,
		server.sim.GetReceiveTime()})
}

// Callback for when a message is received on this server.
// When the snapshot algorithm completes on this server, this function
// should notify the simulator by calling `sim.NotifySnapshotComplete`.
func (server *Server) HandlePacket(src string, message interface{}) {
	// TODO: IMPLEMENT ME
	/**
	  When you receive a marker message on an interface:
	      ● If you haven’t started the snapshot procedure yet, record your local state and send
	      marker messages on all outbound interfaces
	      ● Stop recording messages you receive on this interface
	      ● Start recording messages you receive on all other interfaces
	*/
	switch msg := message.(type) {
	case TokenMessage:
		server.Tokens += msg.numTokens
		ids := make([]int, 0)
		server.SnapshotStates.Range(func(key, value interface{}) bool {
			snapshotId := key.(int)
			// snap := value.(SnapshotState)
			if !server.markersReceived[snapshotId][src] {
				ids = append(ids, snapshotId)
				//snap.messages = append(snap.messages, &SnapshotMessage{src, server.Id, msg})
			}
			return true
		})
		snapshotMessage := &SnapshotMessage{src, server.Id, msg}
		for _, id := range ids {
			snapRaw, _ := server.SnapshotStates.Load(id)
			snap := snapRaw.(SnapshotState)
			snap.messages = append(snap.messages, snapshotMessage)
			server.SnapshotStates.Store(id, snap)
		}
	case MarkerMessage:
		_, ok := server.SnapshotStates.Load(msg.snapshotId)
		if !ok {
			server.StartSnapshot(msg.snapshotId)
		}
		server.markersReceived[msg.snapshotId][src] = true
		if len(server.markersReceived[msg.snapshotId]) == len(server.inboundLinks) {
			// defer delete(server.markersReceived, msg.snapshotId)
			// defer delete(server.SnapshotStates, msg.snapshotId)
			server.sim.NotifySnapshotComplete(server.Id, msg.snapshotId)
		}
	}
}

// Start the chandy-lamport snapshot algorithm on this server.
// This should be called only once per server.
func (server *Server) StartSnapshot(snapshotId int) {
	// TODO: IMPLEMENT ME
	// Store local state
	snap := SnapshotState{snapshotId, make(map[string]int), make([]*SnapshotMessage, 0)}
	snap.tokens[server.Id] = server.Tokens
	server.SnapshotStates.Store(snapshotId, snap)
	server.markersReceived[snapshotId] = make(map[string]bool)

	// Send marker messages on all outbound interfaces
	server.SendToNeighbors(MarkerMessage{snapshotId})
}
