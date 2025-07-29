# Requirements Document

## Introduction

This feature involves refactoring the existing C++ peer applications (`peer_1` and `peer_2`) from a client/server relay architecture to a true proxy architecture. The current implementation creates new connections for each data transmission, which is incompatible with Tor's Pluggable Transport system that requires persistent tunnel connections. The refactoring will enable proper integration into the Tor PT chain: `Tor Client -> netshaper-client -> peer_1 -> peer_2 -> netshaper-bridge -> Tor Network`.

## Requirements

### Requirement 1

**User Story:** As a Tor PT system integrator, I want peer_1 to maintain persistent connections instead of creating new connections for each data transmission, so that it can function as a proper proxy in the Tor communication chain.

#### Acceptance Criteria

1. WHEN peer_1's ShapedClient receives a new session THEN it SHALL establish a single persistent connection to peer_2
2. WHEN peer_1's ShapedClient needs to forward data THEN it SHALL use the existing persistent connection rather than creating a new one
3. WHEN a session ends THEN peer_1's ShapedClient SHALL properly close the persistent connection
4. WHEN peer_1's UnshapedServer accepts a new client connection THEN it SHALL trigger the establishment of a single outbound connection via ShapedClient
5. IF the persistent connection fails THEN the system SHALL handle the error gracefully and attempt reconnection if appropriate

### Requirement 2

**User Story:** As a Tor PT system integrator, I want peer_2 to maintain persistent connections instead of creating new connections for each data transmission, so that it can function as a proper proxy in the Tor communication chain.

#### Acceptance Criteria

1. WHEN peer_2's UnshapedClient receives a SYN signal THEN it SHALL establish a single persistent connection to netshaper-bridge
2. WHEN peer_2's UnshapedClient needs to forward data THEN it SHALL use the existing persistent connection rather than creating a new TCP::Client object
3. WHEN a session ends THEN peer_2's UnshapedClient SHALL properly close the persistent connection
4. WHEN peer_2's ShapedServer accepts a new client connection THEN it SHALL trigger the establishment of a single outbound connection via UnshapedClient
5. IF the persistent connection fails THEN the system SHALL handle the error gracefully and attempt reconnection if appropriate

### Requirement 3

**User Story:** As a system administrator, I want the refactored peer applications to properly manage connection lifecycles, so that there are no resource leaks or connection conflicts.

#### Acceptance Criteria

1. WHEN a peer application starts THEN it SHALL initialize connection management structures
2. WHEN a persistent connection is established THEN it SHALL be stored in appropriate member variables
3. WHEN data is transmitted THEN it SHALL reuse existing connections without creating new socket descriptors
4. WHEN a session terminates THEN all associated connections SHALL be properly closed and resources freed
5. WHEN the application shuts down THEN all active connections SHALL be gracefully terminated

### Requirement 4

**User Story:** As a developer, I want the refactored code to maintain clear separation between connection establishment and data transmission, so that the code is maintainable and follows good architectural practices.

#### Acceptance Criteria

1. WHEN implementing connection management THEN connection establishment logic SHALL be separated into dedicated methods
2. WHEN implementing data forwarding THEN it SHALL use pre-established connections stored in member variables
3. WHEN implementing connection cleanup THEN it SHALL be handled by dedicated cleanup methods
4. WHEN reviewing the code THEN the connection lifecycle SHALL be clearly defined and documented
5. WHEN extending the functionality THEN the separation of concerns SHALL make modifications straightforward

### Requirement 5

**User Story:** As a Tor PT system user, I want the refactored peer applications to successfully integrate with the existing Tor PT infrastructure, so that the complete communication chain functions correctly.

#### Acceptance Criteria

1. WHEN the Tor client initiates a connection THEN the traffic SHALL flow successfully through the entire chain: Tor Client -> netshaper-client -> peer_1 -> peer_2 -> netshaper-bridge -> Tor Network
2. WHEN data is transmitted through the chain THEN it SHALL maintain data integrity and proper sequencing
3. WHEN the Tor bridge bootstraps THEN the Tor client SHALL be able to establish and maintain connections
4. WHEN multiple concurrent connections are active THEN each SHALL maintain its own persistent tunnel without interference
5. WHEN connection errors occur THEN they SHALL be handled without causing system-wide failures

### Requirement 6

**User Story:** As a system integrator, I want the refactored applications to handle edge cases and error conditions gracefully, so that the system remains robust in production environments.

#### Acceptance Criteria

1. WHEN a persistent connection is lost unexpectedly THEN the system SHALL detect the failure and handle it appropriately
2. WHEN network conditions cause connection delays THEN the system SHALL handle timeouts gracefully
3. WHEN the destination service is temporarily unavailable THEN the system SHALL provide appropriate error handling
4. WHEN system resources are constrained THEN the connection management SHALL not exhaust available resources
5. WHEN debugging connection issues THEN the system SHALL provide sufficient logging and diagnostic information