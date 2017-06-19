package couchdb_admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/cabify/couchdb-admin/httpUtils"
	"github.com/kr/pretty"
)

type Cluster struct {
	NodesInfo Nodes
}

type Nodes struct {
	AllNodes     []string `json:"all_nodes"`
	ClusterNodes []string `json:"cluster_nodes"`
}

func LoadCluster(ahr *httpUtils.AuthenticatedHttpRequester) *Cluster {
	cluster := &Cluster{}
	cluster.refreshNodesInfo(ahr)
	return cluster
}

func (c *Cluster) refreshNodesInfo(ahr *httpUtils.AuthenticatedHttpRequester) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:5984/_membership", ahr.GetServer()), nil)
	if err != nil {
		log.Fatal(err)
	}

	var info Nodes
	ahr.RunRequest(req, &info)
	pretty.Println(c)
	c.NodesInfo = info

	fmt.Println("Current cluster layout")
	pretty.Println(c.NodesInfo)
}

func (cluster *Cluster) knowsNode(node string) bool {
	for _, n := range cluster.NodesInfo.ClusterNodes {
		if node == n {
			return true
		}
	}
	return false
}

func getLastRevForNode(node string, ahr *httpUtils.AuthenticatedHttpRequester) string {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:5986/_nodes/%s", ahr.GetServer(), node), nil)
	if err != nil {
		log.Fatal(err)
	}

	var nodeDetails = struct {
		Rev string `json:"_rev"`
	}{}

	ahr.RunRequest(req, &nodeDetails)
	return nodeDetails.Rev
}

func (cluster *Cluster) AddNode(nodeAddr string, ahr *httpUtils.AuthenticatedHttpRequester) error {
	node := fmt.Sprintf("couchdb@%s", nodeAddr)

	if cluster.IsNodeUpAndJoined(node) {
		return fmt.Errorf("Node: %s is already part of the cluster!", node)
	}

	body := make(map[string]string)
	if cluster.knowsNode(node) {
		body["_rev"] = getLastRevForNode(node, ahr)
	}

	body_bytes, err := json.Marshal(body)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5986/_nodes/%s", ahr.GetServer(), node), bytes.NewReader(body_bytes))

	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	ahr.RunRequest(req, nil)

	cluster.refreshNodesInfo(ahr)

	return nil
}

func (cluster *Cluster) IsNodeUpAndJoined(node string) bool {
	for _, n := range cluster.NodesInfo.AllNodes {
		if node == n {
			return cluster.knowsNode(node)
		}
	}
	return false
}
