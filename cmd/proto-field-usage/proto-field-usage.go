package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mkeeler/proto-analyze/pkg/dynamicregistry"
	"github.com/mkeeler/proto-analyze/pkg/fieldusage"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func main() {
	protoMessageName := ""
	protoFormat := ""
	protoSet := ""

	flag.StringVar(&protoMessageName, "message", "", "protoreflect.FullName of the top level message to unmarshal")
	flag.StringVar(&protoSet, "protoset", "", "file containing a protobuf FileDescriptorSet of types to load")
	flag.StringVar(&protoFormat, "format", "json", "format to unmarshal from the files: json or proto (default is json)")

	flag.Parse()

	if protoMessageName == "" {
		fmt.Fprintf(os.Stderr, "-message must be supplied to unmarshal the top level objects")
		os.Exit(-1)
	}

	registry := protoregistry.GlobalTypes
	if protoSet != "" {
		newRegistry, err := dynamicregistry.CreateRegistryFromFile(protoSet)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create the protobuf type registry from the provided protoset file")
			os.Exit(-1)
		}
		registry = newRegistry
	}

	var files []string

	// collect all the file paths to load
	for _, path := range flag.Args() {
		filepath.WalkDir(path, func(path string, dirEnt fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !dirEnt.IsDir() {
				files = append(files, path)
			}
			return nil
		})
	}

	collector := fieldusage.NewCollector(fieldusage.CollectorConfig{RemoveListIndexing: true})
	usage := collector.CollectUsageFromJSONFile
	switch strings.ToLower(protoFormat) {
	case "json":
		usage = collector.CollectUsageFromJSONFile
	default:
		usage = collector.CollectUsageFromProtoFile
	}
	for _, fpath := range files {
		err := usage(protoreflect.FullName(protoMessageName), fpath, registry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting field usage from %s: %v", fpath, err)
			os.Exit(-1)
		}
	}

	indentLevel := 0

	collector.Walk(func(pathElem string) {
		fmt.Printf("%s%v\n", strings.Repeat("   ", indentLevel), pathElem)
		indentLevel++
	}, func(pathElem string) {
		indentLevel--
	})
}
