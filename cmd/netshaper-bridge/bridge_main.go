package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
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

// handleConnection accepts a single QUIC connection and proxies data to the Tor ORPort.
func handleConnection(conn quic.Connection, orPort *net.TCPAddr) {
	defer conn.CloseWithError(0, "connection closed")
	log.Printf("Accepted QUIC connection from: %s", conn.RemoteAddr())

	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		logErrorf("Failed to accept stream from client %s: %v", conn.RemoteAddr(), err)
		return
	}
	defer stream.Close()
	log.Printf("Accepted stream from client %s", conn.RemoteAddr())

	torConn, err := net.DialTCP("tcp", nil, orPort)
	if err != nil {
		logErrorf("Failed to connect to ORPort %s: %v", orPort.String(), err)
		return
	}
	defer torConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	// QUIC -> Tor
	go func() {
		defer wg.Done()
		if _, err := io.Copy(torConn, stream); err != nil && !isUseOfClosedNetConnError(err) {
			logErrorf("Error copying data from QUIC to Tor: %v", err)
		}
		torConn.CloseWrite()
	}()

	// Tor -> QUIC
	go func() {
		defer wg.Done()
		if _, err := io.Copy(stream, torConn); err != nil && !isUseOfClosedNetConnError(err) {
			logErrorf("Error copying data from Tor to QUIC: %v", err)
		}
		stream.Close()
	}()

	wg.Wait()
}

// acceptLoop continuously accepts new connections from the QUIC listener.
func acceptLoop(ln *quic.Listener, orPort *net.TCPAddr) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("FATAL: Recovered from panic in acceptLoop: %v", r)
		}
	}()
	log.Printf("Accepting connections on %s", ln.Addr())
	for {
		conn, err := ln.Accept(context.Background())
		if err != nil {
			logErrorf("Failed to accept connection: %v", err)
			break
		}
		go handleConnection(conn, orPort)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)
	log.Printf("Starting NetShaper server version %s", ptVersion)

	certPath := flag.String("cert-path", ".", "Path to the directory containing cert.pem and key.pem")
	flag.Parse()

	// ServerSetup is still useful for parsing Bindaddrs
	ptInfo, err := pt.ServerSetup(nil)
	if err != nil {
		log.Fatalf("Error setting up PT server: %v", err)
	}

	// THE FIX IS HERE: Read the ORPort directly from the environment variable.
	// This bypasses any goptlib versioning issues.
	orPortStr := os.Getenv("TOR_PT_ORPORT")
	if orPortStr == "" {
		log.Fatalf("TOR_PT_ORPORT environment variable not set. Cannot proceed.")
	}
	orPort, err := net.ResolveTCPAddr("tcp", orPortStr)
	if err != nil {
		log.Fatalf("Failed to resolve ORPort address '%s': %v", orPortStr, err)
	}
	log.Printf("Successfully resolved ORPort: %s", orPort.String())


	certFile := filepath.Join(*certPath, "cert.pem")
	keyFile := filepath.Join(*certPath, "key.pem")

	tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load certificate and key from %s and %s: %v", certFile, keyFile, err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"netshaper-pt"},
	}

	var wg sync.WaitGroup
	listeners := make([]*quic.Listener, 0)

	for _, bindInfo := range ptInfo.Bindaddrs {
		if bindInfo.MethodName != transportName {
			continue
		}
		
		log.Printf("Attempting to bind QUIC listener to address: %s", bindInfo.Addr.String())
		ln, err := quic.ListenAddr(bindInfo.Addr.String(), tlsConfig, nil)
		if err != nil {
			log.Printf("ERROR: quic.ListenAddr failed: %v", err)
			pt.SmethodError(bindInfo.MethodName, fmt.Sprintf("failed to listen on QUIC: %s", err))
			continue
		}

		log.Printf("SUCCESS: Listening for %s QUIC connections on %s (Reported by listener: %s)", bindInfo.MethodName, bindInfo.Addr.String(), ln.Addr())
		pt.Smethod(bindInfo.MethodName, ln.Addr())

		wg.Add(1)
		go func(l *quic.Listener, p *net.TCPAddr) {
			defer wg.Done()
			acceptLoop(l, p)
		}(ln, orPort)
		listeners = append(listeners, ln)
	}

	pt.SmethodsDone()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Received shutdown signal. Cleaning up.")

	for _, ln := range listeners {
		ln.Close()
	}

	wg.Wait()
	log.Println("NetShaper server finished.")
}

func isUseOfClosedNetConnError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "use of closed network connection") ||
		strings.Contains(err.Error(), "closing")
}

