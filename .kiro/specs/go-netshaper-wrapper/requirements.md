# Requirements Document

## Introduction

This feature involves creating a Go wrapper for the NetShaper C++ shaping functionalities. NetShaper is a network traffic shaping system that uses differential privacy to hide traffic patterns. The system consists of two main components: Peer1 (client middlebox) and Peer2 (server middlebox). The Go wrapper will provide a clean, idiomatic Go API to interact with the underlying C++ shaping components while maintaining the performance and privacy guarantees of the original implementation.

## Requirements

### Requirement 1

**User Story:** As a Go developer, I want to use NetShaper's traffic shaping capabilities from Go code, so that I can integrate privacy-preserving network shaping into Go applications.

#### Acceptance Criteria

1. WHEN a Go application imports the wrapper THEN it SHALL provide access to both Peer1 (ShapedClient) and Peer2 (ShapedServer) functionalities
2. WHEN the wrapper is initialized THEN it SHALL accept configuration parameters equivalent to the C++ config structs
3. WHEN the wrapper starts shaping THEN it SHALL maintain the same differential privacy guarantees as the original C++ implementation

### Requirement 2

**User Story:** As a Go developer, I want to configure NetShaper components through Go structs, so that I can easily customize shaping behavior without dealing with C++ configuration files.

#### Acceptance Criteria

1. WHEN configuring a ShapedClient THEN the wrapper SHALL accept Go structs with fields for peer2Addr, peer2Port, noiseMultiplier, sensitivity, and other C++ config parameters
2. WHEN configuring a ShapedServer THEN the wrapper SHALL accept Go structs with fields for serverCert, serverKey, listeningPort, and other server-specific parameters
3. WHEN invalid configuration is provided THEN the wrapper SHALL return appropriate Go errors

### Requirement 3

**User Story:** As a Go developer, I want to start and stop NetShaper components programmatically, so that I can control the lifecycle of traffic shaping within my application.

#### Acceptance Criteria

1. WHEN starting a ShapedClient THEN it SHALL establish connection to the specified peer2 address and begin traffic shaping
2. WHEN starting a ShapedServer THEN it SHALL begin listening on the specified port for shaped traffic
3. WHEN stopping components THEN they SHALL cleanly shut down and release resources
4. WHEN errors occur during startup THEN appropriate Go errors SHALL be returned

### Requirement 4

**User Story:** As a Go developer, I want to send and receive data through the shaped channels, so that my application traffic benefits from differential privacy protection.

#### Acceptance Criteria

1. WHEN sending data through ShapedClient THEN it SHALL be processed through the differential privacy shaping algorithm
2. WHEN receiving data through ShapedServer THEN it SHALL be delivered to the application after privacy processing
3. WHEN the shaping queues are full THEN the wrapper SHALL handle backpressure appropriately
4. WHEN connection status changes THEN the wrapper SHALL notify the Go application through callbacks or channels

### Requirement 5

**User Story:** As a Go developer, I want proper error handling and logging integration, so that I can debug issues and monitor the shaping system's health.

#### Acceptance Criteria

1. WHEN C++ components log messages THEN they SHALL be accessible through Go logging interfaces
2. WHEN errors occur in the C++ layer THEN they SHALL be converted to appropriate Go errors
3. WHEN the wrapper encounters resource constraints THEN it SHALL provide meaningful error messages
4. WHEN debugging is enabled THEN detailed operational information SHALL be available

### Requirement 6

**User Story:** As a Go developer, I want the wrapper to handle memory management automatically, so that I don't need to manually manage C++ object lifecycles.

#### Acceptance Criteria

1. WHEN Go objects are garbage collected THEN associated C++ resources SHALL be automatically cleaned up
2. WHEN the wrapper allocates C++ objects THEN it SHALL track them for proper deallocation
3. WHEN shared memory is used THEN it SHALL be properly managed across the Go-C++ boundary
4. WHEN the application exits THEN all C++ resources SHALL be cleanly released