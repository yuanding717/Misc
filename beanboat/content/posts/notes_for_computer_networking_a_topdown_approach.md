+++
title = "notes_for_computer_networking_a_topdown_approach"
description = "Notes for <Computer Networking a Topdown Approach>"
tags = [
    "computer_network",
]
date = "2018-10-05"
categories = [
    "notes_for_introduction_to_computer_security",
    "book_notes",
]
menu = "main"
+++

## Concepts
* **Protocol**: A protocol defines the format and the order of messages exchanged between two or more communicating entities, as well as the actions taken on the transmission and/or receipt of a message or other event.
 * hardware-implemented protocols in two physically connected computers control the flow of bits on the “wire” between the two network interface cards
 * congestion-control protocols in end systems control the rate at which packets are transmitted between sender and receiver
 * protocols in routers determine a packet’s path from source to destination.
 <!--more-->

* **Access Networks**: the network that physically connects an end system to the first router on a path from the end system to any other distant end system.
 * Types: digital subscriber line (DSL), cable, fiber to the home(FTTH), satellite link, local area network(LAN)(including Ethernet and WiFi)

## The Network Core
### Packet Switching
* Definition: A packet switch takes a packet arriving on one of its incoming communication links and forwards that packet on one of its outgoing communication links
* Types: router and link layer switches
* **store-and-forward transmission**: packet switch must receive the entire packet before it can begin to transmit the first bit of the packet onto the outbound link.
 * packets suffer output buffer queuing delays
 * packet loss: an arriving packet may find the buffer is completely full with other packets waiting for transmission, either the arriving packet or one of the alreday-queued packets will be dropped.
* Forwarding table in Internet: each router has a forwarding table that maps destination address to that router's outbound links
* Advantages:
 * offers better sharing of transmission capacity than circuit switching
 * simpler, more efficient, and less costly to implement than circuit switching

### Circuit Switching
* Definition: the resources needed along a path (buffers, link transmission rate) to provide for communication between the end systems are reserved for the duration of the communication session between the end systems
