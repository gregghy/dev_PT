# Requirements Document

## Introduction

This feature enables the Tor Pluggable Transport to operate in two modes: with NetShaper differential privacy shaping enabled, or without shaping for direct QUIC connections. This provides flexibility for users to choose between enhanced privacy (with shaping) and better performance (without shaping), and allows for easy comparison and testing of both modes.

## Requirements

### Requirement 1: Mode Selection Interface

**User Story:** As a Tor user, I want to choose whether to use traffic shaping or not, so that I can balance privacy and performance based on my needs.

#### Acceptance Criteria

1. WHEN the PT client is started with a `--no-shaping` flag THEN the system SHALL bypass NetShaper and use direct QUIC connections
2. WHEN the PT client is started without the `--no-shaping` flag THEN the system SHALL use NetShaper traffic shaping by default
3. WHEN the PT bridge is started with a `--no-shaping` flag THEN the system SHALL bypass NetShaper and accept direct QUIC connections
4. WHEN the PT bridge is started without the `--no-shaping` flag THEN the system SHALL use NetShaper traffic shaping by default
5. WHEN an invalid mode is specified THEN the system SHALL display an error message and exit gracefully

### Requirement 2: Environment Variable Configuration

**User Story:** As a system administrator, I want to configure the shaping mode via environment variables, so that I can control the behavior without modifying command line arguments.

#### Acceptance Criteria

1. WHEN the environment variable `NETSHAPER_DISABLE=true` is set THEN the system SHALL disable traffic shaping regardless of command line flags
2. WHEN the environment variable `NETSHAPER_DISABLE=false` is set THEN the system SHALL enable traffic shaping (default behavior)
3. WHEN both environment variable and command line flag are specified THEN the command line flag SHALL take precedence
4. WHEN neither environment variable nor command line flag is specified THEN the system SHALL default to shaping enabled

### Requirement 3: Tor Configuration Compatibility

**User Story:** As a Tor user, I want to use the same PT binary for both modes, so that I don't need to manage multiple binaries or configurations.

#### Acceptance Criteria

1. WHEN using the same PT binary in torrc THEN the system SHALL support both shaping and non-shaping modes based on configuration
2. WHEN the PT is configured in torrc with shaping disabled THEN the system SHALL function identically to the original PT implementation
3. WHEN the PT is configured in torrc with shaping enabled THEN the system SHALL provide differential privacy protection
4. WHEN switching between modes THEN the system SHALL maintain compatibility with existing Tor bridge configurations

### Requirement 4: Performance Mode Detection

**User Story:** As a developer, I want the system to automatically detect when performance testing is needed, so that I can easily benchmark both modes.

#### Acceptance Criteria

1. WHEN the `--benchmark` flag is provided THEN the system SHALL run both shaping and non-shaping modes sequentially for comparison
2. WHEN benchmark mode is active THEN the system SHALL log performance metrics for both modes
3. WHEN benchmark mode completes THEN the system SHALL output a comparison report
4. WHEN benchmark mode is interrupted THEN the system SHALL clean up both mode processes gracefully

### Requirement 5: Runtime Mode Switching

**User Story:** As a power user, I want to switch between shaping modes without restarting the PT, so that I can adapt to changing network conditions.

#### Acceptance Criteria

1. WHEN a SIGUSR1 signal is sent to the PT process THEN the system SHALL toggle between shaping and non-shaping modes
2. WHEN mode switching occurs THEN the system SHALL maintain existing connections during the transition
3. WHEN mode switching fails THEN the system SHALL revert to the previous mode and log the error
4. WHEN mode switching succeeds THEN the system SHALL log the new mode status

### Requirement 6: Configuration Validation

**User Story:** As a system administrator, I want the system to validate that NetShaper components are available before enabling shaping, so that I get clear error messages if the setup is incomplete.

#### Acceptance Criteria

1. WHEN shaping mode is requested AND NetShaper binaries are missing THEN the system SHALL fall back to non-shaping mode with a warning
2. WHEN shaping mode is requested AND certificates are missing THEN the system SHALL fall back to non-shaping mode with a warning
3. WHEN non-shaping mode is requested THEN the system SHALL NOT check for NetShaper dependencies
4. WHEN dependency validation fails THEN the system SHALL provide clear instructions for resolving the issue

### Requirement 7: Logging and Monitoring

**User Story:** As a network administrator, I want clear logging to understand which mode is active and how it's performing, so that I can monitor and troubleshoot the system.

#### Acceptance Criteria

1. WHEN the PT starts THEN the system SHALL log the active mode (shaping enabled/disabled)
2. WHEN mode switching occurs THEN the system SHALL log the transition with timestamps
3. WHEN performance differs significantly between modes THEN the system SHALL log performance warnings
4. WHEN NetShaper components fail THEN the system SHALL log detailed error information for troubleshooting

### Requirement 8: Backward Compatibility

**User Story:** As an existing PT user, I want my current configurations to continue working, so that I can upgrade without breaking my setup.

#### Acceptance Criteria

1. WHEN using existing PT configurations THEN the system SHALL default to shaping enabled (new behavior)
2. WHEN the original PT binaries are replaced THEN the system SHALL maintain the same command line interface
3. WHEN existing torrc configurations are used THEN the system SHALL work without modification
4. WHEN users want the original behavior THEN they SHALL be able to disable shaping with a simple flag

### Requirement 9: Resource Management

**User Story:** As a system administrator, I want the system to efficiently manage resources in both modes, so that I can optimize system performance.

#### Acceptance Criteria

1. WHEN non-shaping mode is active THEN the system SHALL NOT start NetShaper processes
2. WHEN shaping mode is active THEN the system SHALL start only the necessary NetShaper components
3. WHEN switching modes THEN the system SHALL properly clean up unused processes and resources
4. WHEN the system shuts down THEN all processes SHALL be terminated gracefully regardless of mode

### Requirement 10: Testing and Validation

**User Story:** As a developer, I want comprehensive testing for both modes, so that I can ensure reliability and correctness.

#### Acceptance Criteria

1. WHEN integration tests run THEN the system SHALL test both shaping and non-shaping modes
2. WHEN performance tests run THEN the system SHALL measure and compare both modes
3. WHEN connectivity tests run THEN the system SHALL verify that both modes can establish Tor connections
4. WHEN stress tests run THEN the system SHALL validate stability under load for both modes