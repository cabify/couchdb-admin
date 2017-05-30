package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/cabify/couchdb-admin/http_utils"
	"github.com/kr/pretty"
)

type Cluster struct {
	NodesInfo Nodes
}

type Nodes struct {
	AllNodes     []string `json:"all_nodes"`
	ClusterNodes []string `json:"cluster_nodes"`
}

func LoadCluster(server string) *Cluster {
	cluster := &Cluster{}
	cluster.refreshNodesInfo(server)
	return cluster
}

func (c *Cluster) refreshNodesInfo(server string) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:5984/_membership", server), nil)
	if err != nil {
		log.Fatal(err)
	}

	var info Nodes
	http_utils.RunRequest(req, &info)
	fmt.Printf("%# v", pretty.Formatter(c))
	c.NodesInfo = info

	fmt.Println("Current cluster layout")
	fmt.Printf("%# v", pretty.Formatter(c.NodesInfo))
}

func (cluster *Cluster) knowsNode(node string) bool {
	for _, n := range cluster.NodesInfo.ClusterNodes {
		if node == n {
			return true
		}
	}
	return false
}

func getLastRevForNode(server, node string) string {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:5986/_nodes/%s", server, node), nil)
	if err != nil {
		log.Fatal(err)
	}

	var nodeDetails = struct {
		Rev string `json:"_rev"`
	}{}

	http_utils.RunRequest(req, &nodeDetails)
	return nodeDetails.Rev
}

func (cluster *Cluster) AddNode(server, nodeAddr string) error {
	node := fmt.Sprintf("couchdb@%s", nodeAddr)

	body := make(map[string]string)
	if cluster.knowsNode(node) {
		body["_rev"] = getLastRevForNode(server, node)
	}

	body_bytes, err := json.Marshal(body)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5986/_nodes/%s", server, node), bytes.NewReader(body_bytes))

	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	http_utils.RunRequest(req, nil)

	cluster.refreshNodesInfo(server)

	return nil
}
