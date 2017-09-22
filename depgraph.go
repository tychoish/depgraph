package depgraph

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Graph contains a fully materialized graph including all nodes
// (files, symbols, libraries, artifacts) along with its denormalized
// relationships, as well as the nornmalized relationships.
type Graph struct {
	Edges   []Edge `json:"edges"`
	Nodes   []Node `json:"nodes"`
	BuildID string `json:"id,omitempty"`
}

// Node represents a single item in the graph, either a symbol, file,
// artifact or library. The Relationship attribute contains some
// denormalized information about a node's relationships to other nodes.
type Node struct {
	Name          string `json:"id" bson:"name"`
	GraphID       int    `bson:"index" json:"index"`
	Relationships struct {
		DependentLibraries []string `json:"_dependent_libs" bson:"dependent_libs"`
		Libraries          []string `json:"_libs" bson:"libs"`
		Files              []string `json:"_files" bson:"files"`
		DependentFiles     []string `json:"_dependent_files" bson:"dependent_files"`
		Type               NodeType `json:"type" bson:"type"`
	} `json:"node" bson:"relationships"`
}

// NodeRelationship represents a single edge in the graph,
type NodeRelationship struct {
	GraphID int    `bson:"index" json:"index"`
	Name    string `json:"id" bson:"name"`
}

// Represents a single group of edges in the (directed) graph, which contains
// which kind of relationship a list of "to" nodes and the originating node.
type Edge struct {
	Type     EdgeType           `json:"type" bson:"type"`
	FromNode NodeRelationship   `bson:"from_node" json:"from_node"`
	ToNodes  []NodeRelationship `bson:"to_node" json:"to_node"`
}

// New parses a graph and returns the graph structure. New takes a
// build id and a path to the graph source. The path may either be a
// URL which it will download into the current directory, or the location
// of a local file.
func New(build, path string) (*Graph, error) {
	var data []byte
	var err error

	if strings.HasPrefix(path, "http") {
		data, err = cacheDownload(300*time.Hour, path, filepath.Base(path), false)
		if err != nil {
			return nil, errors.Wrap(err, "problem downloading file")
		}
	} else if _, err = os.Stat(path); os.IsNotExist(err) {
		return nil, errors.Errorf("could not find file %s", path)
	} else {
		data, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}

	g := &Graph{}

	if err = json.Unmarshal(data, g); err != nil {
		return nil, errors.Wrap(err, "problem reading json")
	}

	g.BuildID = build

	return g, nil
}
