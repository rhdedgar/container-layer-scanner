package scanner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rhdedgar/container-layer-scanner/pkg/api"
	"github.com/rhdedgar/container-layer-scanner/pkg/clamav"
	"github.com/rhdedgar/container-layer-scanner/pkg/cmd"
	"github.com/rhdedgar/container-layer-scanner/pkg/containerspec"
)

var osMkdir = os.Mkdir
var ioutilTempDir = ioutil.TempDir

type containerMeta struct {
	Container *containerspec.Container
	Image     *containerspec.Image
}

// ContainerLayerScanner is the interface for all image containerLayerScanners.
type ContainerLayerScanner interface {
	// Inspect inspects and serves the image based on the ContainerLayerScannerOptions.
	ClamScanner() error
}

// scanOutputs is a struct to hold all the scan outputs that needs to be served
type scanOutputs struct {
	ScanReport     []byte
	HtmlScanReport []byte
	ScanResults    api.ScanResult
}

// defaultContainerLayerScanner is the default implementation of ContainerLayerScanner.
type defaultContainerLayerScanner struct {
	opts cmd.ContainerLayerScannerOptions
	meta api.InspectorMetadata
	// an optional image server that will server content for inspection.

	scanOutputs scanOutputs
}

// NewInspectorMetadata returns a new InspectorMetadata out of *containerspec.Image
// The OpenSCAP status will be NotRequested
func NewInspectorMetadata(imageMetadata *containerspec.Image) api.InspectorMetadata {
	return api.InspectorMetadata{
		Image: *imageMetadata,
		OpenSCAP: &api.OpenSCAPMetadata{
			Status:           api.StatusNotRequested,
			ErrorMessage:     "",
			ContentTimeStamp: string(time.Now().Format(time.RFC850)),
		},
	}
}

// NewDefaultContainerLayerScanner provides a new default scanner.
func NewDefaultContainerLayerScanner(opts cmd.ContainerLayerScannerOptions) ContainerLayerScanner {
	containerLayerScanner := &defaultContainerLayerScanner{
		opts: opts,
		meta: NewInspectorMetadata(&containerspec.Image{}),
	}

	containerLayerScanner.scanOutputs.ScanResults = api.ScanResult{
		APIVersion: api.DefaultResultsAPIVersion,
		Results:    []api.Result{},
	}

	return containerLayerScanner
}

// Inspect inspects and serves the image based on the ImageInspectorOptions.
func (i *defaultContainerLayerScanner) ClamScanner() error {
	err := i.acquireAndScan()
	if err != nil {
		i.meta.ImageAcquireError = err.Error()
	}

	return err
}

// AcquireAndScan acquires and scans the image based on the ContainerLayerScannerOptions.
func (i *defaultContainerLayerScanner) acquireAndScan() error {
	var (
		scanner api.Scanner
		err     error
		//source  string

		filterFn api.FilesFilter
	)

	ctx := context.Background()

	//if len(i.opts.ScanDir) != 0 {
	//	source = i.opts.ScanDir
	//}

	scanner, err = clamav.NewScanner(i.opts.ClamSocket)
	if err != nil {
		return fmt.Errorf("failed to initialize clamav scanner: %v", err)
	}
	results, _, err := scanner.Scan(ctx, i.opts.ScanDir, &i.meta.Image, filterFn)
	if err != nil {
		log.Printf("DEBUG: Unable to scan directory %q with ClamAV: %v", i.opts.ScanDir, err)
		return err
	}
	i.scanOutputs.ScanResults.Results = append(i.scanOutputs.ScanResults.Results, results...)

	if len(i.opts.PostResultURL) > 0 {
		if err := i.postResults(i.scanOutputs.ScanResults); err != nil {
			log.Printf("Error posting results: %v", err)
			return err
		}
	}

	return nil
}

func (i *defaultContainerLayerScanner) postResults(scanResults api.ScanResult) error {
	url := i.opts.PostResultURL
	log.Printf("Posting results to %q ...", url)
	resultJSON, err := json.Marshal(scanResults)
	if err != nil {
		return err
	}
	client := http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader(resultJSON))
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	log.Printf("DEBUG: Success: %v", resp)
	return nil
}
