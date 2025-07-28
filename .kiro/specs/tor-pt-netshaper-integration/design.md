# Design Document

## Overview

This design integrates NetShaper traffic obfuscation into the existing Tor PT QUIC tunnel system using the simplest possible approach. The integration will wrap the QUIC traffic with NetShaper's differential privacy-based traffic shaping to make the connections less detectable by network analysis.

The design follows a transparent proxy approach where NetShaper processes sit between the PT components and the network, intercepting and obfuscating traffic without requiring major changes to the existing QUIC tunnel logic.

## Architecture

### High-Level Architecture

```
Client Side:
Tor Client → PT Client (SOCKS) → NetShaper Peer1 → Network

Bridge Side:
Network → NetShaper Peer2 → PT Bridge (QUIC) → Tor ORPort
```

### Component Interaction Flow

1. **Bridge Startup**: PT Bridge starts NetShaper Peer2, then binds QUIC listener to localhost
2. **Client Startup**: PT Client starts NetShaper Peer1, then connects through NetShaper to bridge
3. **Traffic Flow**: All QUIC traffic flows through NetShaper processes for obfuscation
4. **Shutdown**: PT processes gracefully terminate NetShaper processes

### NetShaper Integration Points

- **Peer2 (Bridge)**: Acts as obfuscated server, forwards to PT Bridge QUIC listener
- **Peer1 (Client)**: Acts as obfuscated client, receives from PT Client QUIC dialer

## Components and Interfaces

### NetShaper Manager

A new component that manages NetShaper process lifecycle:

```go
type NetShaperManager struct {
    peer1 *netshaper.Peer1Wrapper  // Client-side obfuscation
    peer2 *netshaper.Peer2Wrapper  // Bridge-side obfuscation
    config NetShaperConfig
    logger Logger
}

type NetShaperConfig struct {
    Enabled        bool
    BinaryPath     string
    WorkingDir     string
    Peer1Config    netshaper.Peer1Config
    Peer2Config    netshaper.Peer2Config
    StartupTimeout time.Duration
}
```

### Modified PT Bridge

The bridge will be modified to:
1. Start NetShaper Peer2 before binding QUIC listener
2. Bind QUIC listener to localhost (NetShaper forwards to this)
3. Configure NetShaper to listen on the original PT port
4. Gracefully shutdown NetShaper on exit

### Modified PT Client

The client will be modified to:
1. Start NetShaper Peer1 before making QUIC connections
2. Connect to NetShaper Peer1 instead of directly to bridge
3. Configure NetShaper to forward to the actual bridge address
4. Gracefully shutdown NetShaper on exit

## Data Models

### Configuration Structure

```go
// NetShaper configuration for PT integration
type PTNetShaperConfig struct {
    // Global settings
    Enabled        bool          `json:"enabled"`
    BinaryPath     string        `json:"binaryPath"`
    WorkingDir     string        `json:"workingDir"`
    StartupTimeout time.Duration `json:"startupTimeout"`
    
    // Bridge-specific settings
    Bridge struct {
        Enabled      bool   `json:"enabled"`
        ListenPort   uint16 `json:"listenPort"`   // Port NetShaper listens on
        ForwardPort  uint16 `json:"forwardPort"`  // Port PT Bridge listens on
        CertFile     string `json:"certFile"`
        KeyFile      string `json:"keyFile"`
    } `json:"bridge"`
    
    // Client-specific settings
    Client struct {
        Enabled     bool   `json:"enabled"`
        ListenPort  uint16 `json:"listenPort"`  // Port NetShaper listens on
        BridgeAddr  string `json:"bridgeAddr"` // NetShaper bridge address
        BridgePort  uint16 `json:"bridgePort"` // NetShaper bridge port
    } `json:"client"`
    
    // Privacy settings
    Privacy struct {
        NoiseMultiplier float64 `json:"noiseMultiplier"`
        Sensitivity     float64 `json:"sensitivity"`
        Strategy        string  `json:"strategy"` // "BURST" or "UNIFORM"
    } `json:"privacy"`
}
```

### Port Allocation Strategy

- **Original PT Bridge Port** (24433): NetShaper Peer2 listens here
- **PT Bridge Internal Port** (24434): PT Bridge QUIC listener (localhost only)
- **NetShaper Peer1 Port** (8000): Client-side NetShaper listener
- **NetShaper Peer2 Internal Port** (4567): Peer1 → Peer2 shaped connection

## Error Handling

### NetShaper Process Failures

1. **Startup Failures**: 
   - Log clear error messages with troubleshooting hints
   - Fall back to direct QUIC if configured
   - Provide binary installation guidance

2. **Runtime Failures**:
   - Monitor NetShaper process health
   - Attempt automatic restart with exponential backoff
   - Graceful degradation to direct QUIC if restart fails

3. **Configuration Errors**:
   - Validate NetShaper configs before process start
   - Provide detailed error messages for common misconfigurations
   - Use sensible defaults for optional parameters

### Resource Management

- Automatic cleanup of temporary configuration files
- Proper process termination with timeout handling
- Resource leak prevention through defer statements and cleanup handlers

## Testing Strategy

### Unit Tests

1. **NetShaper Manager Tests**:
   - Configuration validation
   - Process lifecycle management
   - Error handling scenarios

2. **Integration Tests**:
   - PT + NetShaper startup/shutdown sequences
   - Configuration file generation and cleanup
   - Process monitoring and restart logic

### Integration Tests

1. **End-to-End Traffic Tests**:
   - QUIC tunnel functionality with NetShaper enabled
   - Performance impact measurement
   - Traffic obfuscation verification

2. **Failure Scenario Tests**:
   - NetShaper binary missing
   - NetShaper process crashes
   - Configuration file corruption
   - Port conflicts

### Performance Tests

1. **Latency Impact**: Measure additional latency introduced by NetShaper
2. **Throughput Impact**: Verify QUIC tunnel performance with obfuscation
3. **Resource Usage**: Monitor CPU and memory overhead
4. **Startup Time**: Measure time to establish obfuscated connections

## Implementation Phases

### Phase 1: Core Integration
- Implement NetShaperManager component
- Add NetShaper process lifecycle to PT Bridge
- Add NetShaper process lifecycle to PT Client
- Basic configuration and error handling

### Phase 2: Configuration and Robustness
- Configuration file management
- Process monitoring and restart logic
- Comprehensive error handling and logging
- Resource cleanup and leak prevention

### Phase 3: Optimization and Testing
- Performance optimization
- Comprehensive test suite
- Documentation and troubleshooting guides
- Optional fallback mechanisms

## Security Considerations

### Certificate Management
- NetShaper requires TLS certificates for Peer1 ↔ Peer2 communication
- Reuse existing PT certificates or generate dedicated NetShaper certificates
- Ensure proper certificate validation in production environments

### Process Isolation
- NetShaper processes run as separate processes with limited privileges
- Temporary configuration files use secure permissions (0600)
- Process communication through localhost interfaces only

### Privacy Protection
- NetShaper differential privacy parameters tuned for Tor traffic patterns
- Configurable privacy/performance trade-offs
- No logging of sensitive traffic data in NetShaper processes