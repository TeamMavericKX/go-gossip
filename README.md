# go-gossip

An implementation of the SWIM (Scalable Weakly-consistent Infection-style Process Group Membership) protocol in Go.

## Project Plan

This project has been developed in several phases:

1.  **Phase 1: Core SWIM Protocol Implementation (Completed)**
    *   Implemented basic node membership (joining and leaving the cluster).
    *   Implemented the gossip mechanism for disseminating membership updates.
    *   Implemented failure detection and node removal. (To be implemented soon)

2.  **Phase 2: Advanced Features (Completed)**
    *   Added support for disseminating custom data (payloads) along with membership information.
    *   Implemented anti-entropy mechanisms to ensure eventual consistency.
    *   Added encryption for secure communication between nodes.

3.  **Phase 3: Testing and Documentation (In Progress)**
    *   Wrote comprehensive unit and integration tests. (Unit tests for core components completed)
    *   Create detailed documentation, including High-Level Design (HLD), Low-Level Design (LLD), and architecture diagrams.

## High-Level Design (HLD)

The `go-gossip` system aims to provide a robust and scalable membership management solution for distributed services using the SWIM protocol. It operates on a peer-to-peer model where each node in the cluster maintains a local view of the entire cluster's membership. This membership list includes information about other nodes, such as their network address, current state (alive, suspected, or dead), and optional custom payloads.

Key aspects of the HLD:

*   **Decentralized:** No central authority or leader node; all nodes are peers. This enhances fault tolerance and scalability.
*   **Weakly Consistent:** Membership information propagates through a gossip-style mechanism, leading to eventual consistency across the cluster.
*   **Failure Detection:** Nodes periodically "ping" a subset of other nodes to detect failures. Suspected nodes are then confirmed dead if they remain unresponsive.
*   **Anti-Entropy:** Mechanisms are in place to actively exchange membership lists between nodes, ensuring that discrepancies are eventually resolved and all nodes converge on a consistent view of the cluster.
*   **Secure Communication:** All inter-node communication is encrypted to prevent eavesdropping and tampering.

## Low-Level Design (LLD)

The `go-gossip` library is structured around the following core components:

*   **Node:** (pkg/gossip/node.go)
    *   Represents a single member within the gossip cluster.
    *   Contains its network `Addr` (e.g., IP:Port), its current `State` (Alive, Suspected, Dead), `LastUpdated` timestamp for consistency checks, and a generic `Payload` field for custom application-specific data.
    *   Implements `json.Marshaler` and `json.Unmarshaler` for `net.Addr` serialization.

*   **MembershipList:** (pkg/gossip/membership.go)
    *   A thread-safe data structure (`sync.RWMutex`) that stores the `Node` objects for all known members of the cluster.
    *   Provides methods for `Add`ing new nodes, `Get`ting a node by address, `All` for retrieving all known nodes, and crucially, `Merge` for incorporating membership updates from other nodes.
    *   The `addOrUpdate` internal method handles the core merging logic, prioritizing newer information based on `LastUpdated` timestamps and preserving non-empty payloads.

*   **Gossiper:** (pkg/gossip/gossiper.go)
    *   The central orchestrator of the SWIM protocol.
    *   Manages the `MembershipList` for its local view of the cluster.
    *   Utilizes a `Transport` interface for all network communication.
    *   Runs two primary goroutines:
        *   `pingLoop`: Periodically sends lightweight "Ping" messages to random nodes to actively detect failures.
        *   `syncLoop`: Periodically sends comprehensive "Sync" messages (containing its entire `MembershipList`) to random nodes to resolve inconsistencies (anti-entropy).
    *   Includes `SetPayload` to dynamically update the local node's custom data and `Members` to expose the current cluster view.
    *   Uses a `sync.WaitGroup` for graceful shutdown of its internal goroutines.

*   **Transport Interface:** (pkg/gossip/transport.go)
    *   Defines the contract for network communication, abstracting away the underlying transport mechanism.
    *   Specifies `Write` (send data to address), `Read` (receive data channel), and `Stop` (terminate transport) methods.

*   **UDPTransport:** (pkg/gossip/udp_transport.go)
    *   A concrete implementation of the `Transport` interface using UDP datagrams.
    *   Handles UDP socket creation, listening for incoming messages, and sending outgoing messages.
    *   Includes a `readLoop` goroutine that continuously reads from the UDP socket and dispatches messages to a channel, ensuring thread-safe buffer handling.

*   **SecureTransport:** (pkg/gossip/secure_transport.go)
    *   A decorator (wrapper) for any `Transport` implementation, providing authenticated encryption.
    *   Uses AES-GCM for secure communication, ensuring confidentiality and integrity.
    *   Encrypts data before `Write`ing it to the underlying transport and decrypts data received from the underlying transport in its `Read` method.
    *   Requires a symmetric `key` for encryption/decryption.

*   **Message:** (pkg/gossip/message.go)
    *   Defines the structure for inter-node communication, including `MessageType` (Ping, Sync) and a `Payload` (the marshaled membership list for `Sync` messages, or empty for `Ping`).
    *   Provides `Encode` and `Decode` methods for JSON serialization/deserialization.

## Architecture

The `go-gossip` architecture is entirely peer-to-peer. Each node running the `Gossiper` service is an independent entity that communicates directly with other nodes in the cluster.

```
+----------------+       +----------------+       +----------------+
|     Node A     |       |     Node B     |       |     Node C     |
| (Gossiper A)   |       | (Gossiper B)   |       | (Gossiper C)   |
+-------+--------+       +-------+--------+       +-------+--------+
        |                        |                        |
        | [SecureTransport]      | [SecureTransport]      | [SecureTransport]
        |                        |                        |
        | (Encrypted UDP)        | (Encrypted UDP)        | (Encrypted UDP)
        |                        |                        |
+-------v--------+       +-------v--------+       +-------v--------+
|   Membership   |<------>|   Membership   |<------>|   Membership   |
|     List A     |        |     List B     |        |     List C     |
+-------+--------+       +-------+--------+       +-------+--------+
        ^                        ^                        ^
        |                        |                        |
        |  (Ping/Sync Messages via Gossip)                |
        +-------------------------------------------------+
```

**Key Architectural Principles:**

*   **Symmetry:** All nodes are identical in functionality; there are no special roles like leaders or primary nodes.
*   **Scalability:** The gossip protocol's probabilistic nature allows it to scale effectively to large clusters without centralized bottlenecks.
*   **Resilience:** The decentralized nature means the system can tolerate the failure of multiple nodes without impacting the overall cluster membership management.
*   **Extensibility:** The use of a generic `Payload` in the `Node` structure and `Message` allows users to disseminate arbitrary application-specific data across the cluster. The `Transport` interface also provides flexibility to swap out underlying communication mechanisms (e.g., TCP, custom protocols) if needed.
*   **Security:** The `SecureTransport` layer ensures that all communication within the gossip network is confidential and tamper-proof.

**Message Flow:**

1.  **Initialization:** A new `Gossiper` is created with its local address, a list of initial peer addresses, and a configured `Transport` (e.g., `SecureTransport` wrapping `UDPTransport`).
2.  **Self-Reporting:** The `Gossiper` immediately adds itself to its `MembershipList`.
3.  **Peer Seeding:** The initial peer addresses are added to the `MembershipList` with `Alive` status and current timestamps.
4.  **Gossip Loops:**
    *   **Ping Loop (Failure Detection):** Periodically, a node randomly selects another node from its `MembershipList` (excluding itself) and sends a lightweight "Ping" message. The lack of a response within a timeout period would lead to a suspicion of failure (failure detection not yet fully implemented, but framework is there).
    *   **Sync Loop (Anti-Entropy):** Periodically, a node randomly selects another node from its `MembershipList` (excluding itself) and sends a "Sync" message. This message contains the sender's entire current `MembershipList`.
5.  **Message Handling:**
    *   When a `Gossiper` receives a message, it decodes it.
    *   If it's a "Ping" message, it currently does nothing (could be extended to send an "Ack" or trigger a "Sync Request").
    *   If it's a "Sync" message, it deserializes the incoming `MembershipList` and `Merge`s it with its local `MembershipList`.
6.  **Merging Logic:** The `MembershipList.Merge` method iterates through the incoming nodes. For each node, it compares its `LastUpdated` timestamp with the locally stored version. If the incoming node is newer, the local entry is updated (state, timestamp). If the incoming node also has a non-empty payload, the local node's payload is updated; otherwise, the local payload is preserved.
7.  **Payload Dissemination:** Custom data (payloads) attached to nodes are propagated through the `Sync` messages and updated in `MembershipList` during the merging process, ensuring all nodes eventually reflect the latest state of each member's payload.

This robust system ensures that all healthy nodes in the cluster eventually converge on a consistent and up-to-date view of the cluster membership, including any custom metadata.