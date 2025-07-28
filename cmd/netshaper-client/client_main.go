package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"git.torproject.org/pluggable-transports/goptlib.git"
	"github.com/quic-go/quic-go"
)

const (
	ptVersion     = "0.0.1"
	transportName = "netshaper"
)

// logErrorf prints an error message to the standard error stream.
func logErrorf(format string, v ...interface{}) {
	prefix := "ERROR: "
	fmt.Fprintf(os.Stderr, prefix+format+"\n", v...)
}

// handleConnection handles a single SOCKS request from Tor.
func handleConnection(conn *pt.SocksConn) {
	defer conn.Close()
	log.Printf("Handling connection for SOCKS target: %s", conn.Req.Target)

	// The server address is the target of the SOCKS request from Tor.
	// This is the bridge address from your torrc file.
	serverAddr := conn.Req.Target
	log.Printf("Dialing server address: '%s' (HEX: %x)", serverAddr, serverAddr)


	// Certificate configuration
	tlsConf := &tls.Config{
		InsecureSkipVerify: true, // OK for local testing with self-signed certs
		NextProtos:         []string{"netshaper-pt"},
	}

	log.Printf("Dialing QUIC server at %s", serverAddr)
	quicConn, err := quic.DialAddr(context.Background(), serverAddr, tlsConf, nil)
	if err != nil {
		logErrorf("Failed to dial QUIC server %s: %v", serverAddr, err)
		conn.Reject() // Explicitly reject the SOCKS request on failure
		return
	}
	defer quicConn.CloseWithError(0, "connection closed")

	// New bidirectional stream
	stream, err := quicConn.OpenStreamSync(context.Background())
	if err != nil {
		logErrorf("Failed to open QUIC stream: %v", err)
		conn.Reject()
		return
	}
	log.Printf("Successfully opened QUIC stream to %s", serverAddr)

	// Grant the SOCKS connection to Tor, allowing it to proceed.
	err = conn.Grant(&net.TCPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		logErrorf("Failed to grant SOCKS connection: %v", err)
		return
	}

	// Proxy data between the SOCKS connection and the QUIC stream.
	var wg sync.WaitGroup
	wg.Add(2)

	// SOCKS -> QUIC Stream
	go func() {
		defer wg.Done()
		defer stream.Close()
		if _, err := io.Copy(stream, conn); err != nil && !isUseOfClosedNetConnError(err) {
			logErrorf("Error copying data from SOCKS to QUIC: %v", err)
		}
	}()

	// QUIC Stream -> SOCKS
	go func() {
		defer wg.Done()
		if _, err := io.Copy(conn, stream); err != nil && !isUseOfClosedNetConnError(err) {
			logErrorf("Error copying data from QUIC to SOCKS: %v", err)
		}
	}()

	wg.Wait()
	log.Printf("Finished proxying for connection to %s", conn.Req.Target)
}

// acceptLoop accepts incoming SOCKS connections from the Tor client.
func acceptLoop(ln *pt.SocksListener) {
	log.Printf("NetShaper client: accepting SOCKS connections on %s", ln.Addr())
	for {
		conn, err := ln.AcceptSocks()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			logErrorf("Error accepting SOCKS connection: %v", err)
			break
		}
		log.Printf("Accepted SOCKS connection from %s", conn.RemoteAddr())
		go handleConnection(conn)
	}
	log.Printf("NetShaper client: accept loop finished.")
}

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)
	log.Printf("Starting NetShaper client version %s", ptVersion)

	// ClientSetup returns the transport methods Tor wants us to use.
	ptInfo, err := pt.ClientSetup(nil)
	if err != nil {
		log.Fatalf("Error setting up PT client: %v", err)
	}

	listeners := make([]net.Listener, 0)
	var wg sync.WaitGroup

	// THE FIX IS HERE: The field is called MethodNames, not Transports.
	for _, methodName := range ptInfo.MethodNames {
		if methodName != transportName {
			continue
		}

		// Listen on a random port for SOCKS connections from Tor.
		ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
		if err != nil {
			pt.CmethodError(methodName, err.Error())
			continue
		}
		log.Printf("NetShaper client: SOCKS listener ready on %s", ln.Addr())

		// Announce that our transport is ready.
		pt.Cmethod(methodName, "socks5", ln.Addr())

		// Start the accept loop for this listener.
		wg.Add(1)
		go func() {
			defer wg.Done()
			acceptLoop(ln)
		}()
		listeners = append(listeners, ln)
	}

	// Tell Tor we are done with all methods.
	pt.CmethodsDone()

	// Wait for a shutdown signal.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Received shutdown signal. Cleaning up.")

	for _, ln := range listeners {
		ln.Close()
	}
	wg.Wait()
	log.Println("NetShaper client finished.")
}

func isUseOfClosedNetConnError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "use of closed network connection") ||
		strings.Contains(err.Error(), "closing")
}
