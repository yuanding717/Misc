+++
title = "notes_for_Princeton_COS418_Distributed_Systems"
description = "Notes for Princeton <COS418 Distributed Systems>"
tags = [
    "distributed_systems",
]
date = "2018-10-14"
categories = [
    "distributed_systems",
    "course",
]
menu = "main"
+++
## RPC
### Asynchronous RPC
Await RPC response in a separate thread
Multiple ways to implement this:
Pass a callback to RPC that will be invoked later
Use channels to communicate RPC reply back to main thread
<!--more-->

## Clock Synchronization
### Cristian’s algorithm
* Client sends a request packet, timestamped with its local clock T1
* Server timestamps its receipt of the request T2 with its local clock
* Server sends a response packet with its local clock T3 and T2
* Client locally timestamps its receipt of the server’s response T4
* Client samples round trip time &delta; = &delta;req + &delta;resp = (T4 − T1) − (T3 − T2)
* Assume: &delta;req ≈ &delta;resp, client sets clock <- T3 + &delta; * 1/2

### Berkeley Algorithm
* Assumes all machines have equally-accurate local clocks
* Obtains average from participating computers and synchronizes clocks to that average

### The Network Time Protocol (NTP)
#### System structure
* Servers and time sources are arranged in layers (strata)
 * Stratum 0: High-precision time sources themselves. E.g., atomic clocks, shortwave radio time receivers
 * Stratum 1: NTP servers directly connected to Stratum 0
 * Stratum 2: NTP servers that synchronize with Stratum 1 (Stratum 2 servers are clients of Stratum 1 servers)
 * Stratum 3: NTP servers that synchronize with Stratum 2 (Stratum 3 servers are clients of Stratum 2 servers)
* Users’ computers synchronize with Stratum 3 servers

#### NTP operation: Server selection
* Messages between an NTP client and server are exchanged in pairs: request and response, using Cristian’s algorithm
* For i-th message exchange with a particular server, calculate:
 * Clock offset &theta;i from client to server
 * Round trip time &delta;i between client and server
* Over last eight exchanges with server k, the client computes its dispersion &sigma;k = max{i}&delta;i − min{i}&theta;i; Client uses the server with minimum dispersion
* Then uses a best estimate of clock offset

### Logical Time: Lamport clocks
#### Idea
Disregard the precise clock time. Instead, capture just a “happens before” relationship between a pair of events. I.E. for every event, assign it a time value C(a) on which all processes agree. If a < b, then C(a) < C(b)

#### "happens-before" (->)
 * If same process and a occurs before b, then a -> b
 * If c is a message receipt of b, then b -> c
 * Unrelated events are concurrent, written as a || d

#### Algorithm:
 * Before executing an event b, Ci <- Ci + 1
 * Send the local clock in the message m
 * On process Pj receiving a message m, set Cj and receive event time C(c) <- 1 + max(Cj, C(m))
 * Break ties by appending the process number to each event:
   * Process Pi timestamps event e with Ci(e).i
   * C(a).i < C(b).j when: C(a) < C(b), or C(a) = C(b) and i < j

#### Total-ordered multicasting / state machine replication
* On receiving an update from client, broadcast to others (including yourself)
* On receiving or processing an update:
 * a) Add it to your local queue, if received update
 * b) Broadcast an acknowledgement message to every replica (including yourself), only from head of queue
* On receiving an acknowledgement, mark corresponding update acknowledged in your queue
* Remove and process updates everyone has ack’ed from head of queue

### Logical Time: Vector clocks
* Objective: Lamport clock timestamps do not capture causality. C(a) < C(b) =>  a -> b or a || b
* Definition: a Vector Clock (VC) is a vector of integers, one entry for each process in the entire distributed system. Label event e with VC(e) = [c1, c2 …, cn]. Each entry ck is a count of events in process k that causally precede e.
* Update rules:
 * Initially, all vectors are [0, 0, …, 0]
 * For each local event on process i, increment local entry ci
 * If process j receives message with vector [d1, d2, …, dn]:
   * Set each local entry ck = max{ck, dk}
   * Increment local entry cj
* Compare:
 * V(a) = V(b) when ak = bk for all k
 * V(a) < V(b) when ak ≤ bk for all k and V(a) ≠ V(b)
 * a || b if ai < bi and aj > bj, some i, j
* Capture causality:
 * V(a) < V(b) then there is a chain of events linked by Happens-Before (->) between a and b
 * V(a) || V(b) then there is no such chain of events between a and b

## Bayou’s take-away ideas
* Eventual consistency, eventually if updates stop, all replicas are the same
* Update functions for automatic applicationdriven conflict resolution
* Ordered update log is the real truth, not the DB
* Application of Lamport clocks for eventual consistency that respect causality

## Chandy-Lamport snapshot algorithm
* Intuition
 * Guarantee zero loss + zero duplication
 * If you haven’t snapshotted your local state yet, do NOT record future messages you receive
 * If you have snapshotted your local state, do record future messages you receive
* Key idea: Servers send marker messages to each other
* Marker messages:
 * mark the beginning of the snapshot process on the server
 * act as a barrier (stopper) for recording messages
* Assumptions
 * There are no failures and all messages arrive intact and only once
 * The communication channels are unidirectional and FIFO ordered
 * There is a communication path between any two processes in the system
 * Any process may initiate the snapshot algorithm
 * The snapshot algorithm does not interfere with the normal execution of the processes
 * Each process in the system records its local state and the state of its incoming channels
* Procedures
 * Starting the snapshot procedure on a server:
   * Record local state
   * Send marker messages on all outbound interfaces
 * When you receive a marker message on an interface:
   * If you haven’t started the snapshot procedure yet, record your local state and send marker messages on all outbound interfaces
   * Stop recording messages you receive on this interface
   * Start recording messages you receive on all other interfaces
 * Terminate when all servers have received marker messages on all interfaces
