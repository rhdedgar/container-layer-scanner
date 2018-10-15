package main

import (
	"flag"
	"log"

	clscmd "github.com/rhdedgar/container-layer-scanner/pkg/cmd"
	mainscan "github.com/rhdedgar/container-layer-scanner/pkg/mainscan"
)

func main() {
	scannerOptions := clscmd.NewDefaultContainerLayerScannerOptions()

	flag.StringVar(&scannerOptions.ScanDir, "scan-dir", scannerOptions.ScanDir, "Docker container to inspect (cannot be used with the image option)")
	flag.StringVar(&scannerOptions.ScanResultsDir, "scan-results-dir", scannerOptions.ScanResultsDir, "The directory that will contain the results of the scan")
	flag.StringVar(&scannerOptions.ClamSocket, "clam-socket", scannerOptions.ClamSocket, "Location of clamav socket file ")
	flag.StringVar(&scannerOptions.PostResultURL, "post-results-url", scannerOptions.PostResultURL, "After scan finish, HTTP POST the results to this URL")
	flag.StringVar(&scannerOptions.OutFile, "out-file", scannerOptions.OutFile, "Write the results of the scan to a local file)")

	flag.Parse()

	if err := scannerOptions.Validate(); err != nil {
		log.Fatalf("Error: %v", err)
	}

	scanner := mainscan.NewDefaultContainerLayerScanner(*scannerOptions)
	if err := scanner.ClamScanner(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
