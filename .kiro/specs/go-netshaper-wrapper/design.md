# Design Document

## Overview

The Go NetShaper wrapper provides a simple Go interface to the existing C++ NetShaper executables. Rather than creating complex CGO bindings, this design takes a simpler approach by wrapping the existing C++ binaries as external processes and providing a Go API to configure and control them.

The wrapper exposes two main components:
- **Peer1Wrapper**: Manages the client middlebox (peer1 executable)
- **Peer2Wrapper**: Manages the server middlebox (peer2 executable)

This approach prioritizes simplicity, maintainability, and quick implementation while preserving all the functionality of the original C++ system.

## Architecture

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Go Client     │    │  Go NetShaper   │    │   Go Server     │
│   Application   │◄──►│    Wrapper      │◄──►│   Application   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │ Process Manager │
                       └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │  C++ NetShaper  │
                       │   Executables   │
                       │  (peer1/peer2)  │
                       └─────────────────┘
```

### Component Architecture

The wrapper consists of two main layers:

1. **Go API Layer**: Provides simple Go interfaces for configuration and process management
2. **Process Management Layer**: Handles starting/stopping C++ executables and managing their lifecycle

### Process Management Strategy

- **Configuration**: Generate JSON config files for C++ executables
- **Process lifecycle**: Use `os/exec` to start/stop peer1 and peer2 processes
- **Communication**: Use standard TCP/QUIC networking (no special IPC needed)
- **Cleanup**: Proper process termination and resource cleanup

## Components and Interfaces

### Core Go Types

```go
// Peer1Config represents configuration for the client middlebox
type Peer1Config struct {
    LogLevel               string  `json:"logLevel"`
    MaxClients             int     `json:"maxClients"`
    AppName                string  `json:"appName"`
    QueueSize              int     `json:"queueSize"`
    UnshapedServer         UnshapedServerConfig `json:"unshapedServer"`
    ShapedClient           ShapedClientConfig   `json:"shapedClient"`
}

// Peer2Config represents configuration for the server middlebox  
type Peer2Config struct {
    LogLevel               string  `json:"logLevel"`
    MaxPeers               int     `json:"maxPeers"`
    MaxStreamsPerPeer      int     `json:"maxStreamsPerPeer"`
    AppName                string  `json:"appName"`
    QueueSize              int     `json:"queueSize"`
    ShapedServer           ShapedServerConfig   `json:"shapedServer"`
    UnshapedClient         UnshapedClientConfig `json:"unshapedClient"`
}

// Supporting config structs match the C++ structure exactly
type ShapedClientConfig struct {
    Peer2Addr              string  `json:"peer2Addr"`
    Peer2Port              int     `json:"peer2Port"`
    NoiseMultiplier        float64 `json:"noiseMultiplier"`
    Sensitivity            float64 `json:"sensitivity"`
    // ... other fields
}
```

### Main Interfaces

```go
// Peer1Wrapper manages the client middlebox process
type Peer1Wrapper struct {
    config  Peer1Config
    process *os.Process
    cmd     *exec.Cmd
}

// Peer2Wrapper manages the server middlebox process
type Peer2Wrapper struct {
    config  Peer2Config
    process *os.Process
    cmd     *exec.Cmd
}

// Methods for both wrappers
func (p *Peer1Wrapper) Start() error
func (p *Peer1Wrapper) Stop() error
func (p *Peer1Wrapper) IsRunning() bool

func (p *Peer2Wrapper) Start() error
func (p *Peer2Wrapper) Stop() error
func (p *Peer2Wrapper) IsRunning() bool

// Factory functions
func NewPeer1Wrapper(config Peer1Config) *Peer1Wrapper
func NewPeer2Wrapper(config Peer2Config) *Peer2Wrapper
```

## Data Models

### Configuration Mapping

The Go configuration structs map exactly to the C++ JSON configuration format:

- `Peer1Config` → `peer1_config.json` (used by peer1 executable)
- `Peer2Config` → `peer2_config.json` (used by peer2 executable)

### Data Flow

1. **Configuration**: Go structs → JSON files → C++ executables
2. **Process Management**: Go wrapper starts/stops C++ processes
3. **Network Communication**: Applications connect directly to the running NetShaper processes via TCP/QUIC

### File Structure

```
netshaper-wrapper/
├── pkg/
│   └── netshaper/
│       ├── peer1.go          # Peer1Wrapper implementation
│       ├── peer2.go          # Peer2Wrapper implementation
│       ├── config.go         # Configuration structs
│       └── utils.go          # Helper functions
├── cmd/
│   └── example/
│       └── main.go           # Usage example
└── binaries/
    ├── peer1                 # C++ peer1 executable
    └── peer2                 # C++ peer2 executable
```

## Error Handling

### Error Categories

1. **Configuration Errors**: Invalid JSON, missing executables, invalid parameters
2. **Process Errors**: Failed to start/stop processes, process crashes
3. **File System Errors**: Cannot write config files, missing binaries

### Error Handling Strategy

```go
// Simple error types
var (
    ErrProcessNotRunning = errors.New("netshaper process is not running")
    ErrProcessAlreadyRunning = errors.New("netshaper process is already running")
    ErrInvalidConfig = errors.New("invalid configuration")
    ErrBinaryNotFound = errors.New("netshaper binary not found")
)

// Error wrapping for context
func (p *Peer1Wrapper) Start() error {
    if p.IsRunning() {
        return ErrProcessAlreadyRunning
    }
    
    if err := p.writeConfigFile(); err != nil {
        return fmt.Errorf("failed to write config: %w", err)
    }
    
    // ... start process
}
```

### Recovery Mechanisms

- **Process crashes**: Can be detected and restarted
- **Configuration errors**: Validated before process start
- **Resource cleanup**: Proper process termination and temp file cleanup

## Testing Strategy

### Unit Testing

- **Configuration tests**: Validate JSON generation and config struct marshaling
- **Process management tests**: Test start/stop functionality with mock processes
- **Error handling tests**: Verify proper error propagation and handling

### Integration Testing

- **End-to-end tests**: Start real peer1/peer2 processes and verify they communicate
- **Configuration validation**: Test with various config combinations
- **Process lifecycle tests**: Start, stop, restart scenarios

### Test Infrastructure

```go
// Test utilities
func TestPeer1Wrapper_Start(t *testing.T) {
    config := Peer1Config{
        LogLevel: "WARNING",
        MaxClients: 10,
        // ... test config
    }
    
    wrapper := NewPeer1Wrapper(config)
    err := wrapper.Start()
    assert.NoError(t, err)
    
    defer wrapper.Stop()
    assert.True(t, wrapper.IsRunning())
}

// Helper for integration tests
func SetupTestEnvironment(t *testing.T) (*Peer1Wrapper, *Peer2Wrapper, func()) {
    // Setup test configs and return cleanup function
}
```

### Testing Scenarios

1. **Basic functionality**: Start/stop processes, config file generation
2. **Configuration validation**: Invalid configs, missing binaries
3. **Error conditions**: Process failures, file system errors
4. **Process lifecycle**: Multiple start/stop cycles
5. **Resource cleanup**: Temp files, process termination

### Build Requirements

- **Binary dependencies**: Tests require compiled peer1/peer2 executables
- **Platform testing**: Linux focus (matching C++ implementation)
- **Simple CI**: Standard Go testing without complex dependencies