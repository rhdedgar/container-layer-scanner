package clamav

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/openshift/clam-scanner/pkg/clamav"
	"github.com/rhdedgar/container-layer-scanner/pkg/api"
)

// ScannerName is a string of the name of the scanner
const ScannerName = "clamav"

// ClamScanner is a structure of two vars
// Socket is the location of the clamav socket.
// clamd is a new clamav ClamdSession
type ClamScanner struct {
	Socket string

	clamd clamav.ClamdSession
}

var _ api.Scanner = &ClamScanner{}

// NewScanner initializes a new clamd session
func NewScanner(socket string) (api.Scanner, error) {
	clamSession, err := clamav.NewClamdSession(socket, true)
	if err != nil {
		fmt.Println("NewScanner error")
		return nil, err
	}
	return &ClamScanner{
		Socket: socket,
		clamd:  clamSession,
	}, nil
}

// Scan will scan the image
func (s *ClamScanner) Scan(ctx context.Context, path string, filter api.FilesFilter) ([]api.Result, interface{}, error) {
	scanResults := []api.Result{}
	// Useful for debugging
	scanStarted := time.Now()
	fmt.Println(scanResults)
	defer func() {
		log.Printf("clamav scan took %ds (%d problems found)", int64(time.Since(scanStarted).Seconds()), len(scanResults))
	}()
	if err := s.clamd.ScanPath(ctx, path, clamav.FilterFiles(filter)); err != nil {
		return nil, nil, err
	}
	s.clamd.WaitTillDone()
	defer s.clamd.Close()

	clamResults := s.clamd.GetResults()

	for _, r := range clamResults.Files {
		r := api.Result{
			Name:           ScannerName,
			ScannerVersion: "0.99.2",
			Timestamp:      scanStarted,
			Reference:      fmt.Sprintf("file://%s", strings.TrimPrefix(r.Filename, path)),
			Description:    r.Result,
		}
		scanResults = append(scanResults, r)
	}
	fmt.Println(scanResults)
	return scanResults, nil, nil
}

// Name returns the const ScannerName
func (s *ClamScanner) Name() string {
	return ScannerName
}
