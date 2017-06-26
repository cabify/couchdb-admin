package couchdb_admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/apex/log"
	"github.com/cabify/couchdb-admin/httpUtils"
)

type Cluster struct {
	NodesInfo Nodes
}

type Nodes struct {
	AllNodes     []string `json:"all_nodes"`
	ClusterNodes []string `json:"cluster_nodes"`
}

func LoadCluster(ahr *httpUtils.AuthenticatedHttpRequester) (*Cluster, error) {
	cluster := &Cluster{}
	if err := cluster.refreshNodesInfo(ahr); err != nil {
		return nil, err
	}
	return cluster, nil
}

func (c *Cluster) refreshNodesInfo(ahr *httpUtils.AuthenticatedHttpRequester) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:5984/_membership", ahr.Server()), nil)
	if err != nil {
		return err
	}

	var info Nodes
	if err := ahr.RunRequest(req, &info); err != nil {
		return err
	}
	c.NodesInfo = info

	return nil
}

func (cluster *Cluster) knowsNode(node string) bool {
	for _, n := range cluster.NodesInfo.ClusterNodes {
		if node == n {
			return true
		}
	}
	return false
}

func getLastRevForNode(node string, ahr *httpUtils.AuthenticatedHttpRequester) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:5986/_nodes/%s", ahr.Server(), node), nil)
	if err != nil {
		return "", err
	}

	var nodeDetails = struct {
		Rev string `json:"_rev"`
	}{}

	if err = ahr.RunRequest(req, &nodeDetails); err != nil {
		return "", err
	}
	return nodeDetails.Rev, nil
}

func (cluster *Cluster) AddNode(nodeAddr string, ahr *httpUtils.AuthenticatedHttpRequester) error {
	node := fmt.Sprintf("couchdb@%s", nodeAddr)

	if cluster.IsNodeUpAndJoined(node) {
		return fmt.Errorf("Node: %s is already part of the cluster!", node)
	}

	body := make(map[string]string)
	if cluster.knowsNode(node) {
		rev, err := getLastRevForNode(node, ahr)
		if err != nil {
			return err
		}
		body["_rev"] = rev
	}

	body_bytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5986/_nodes/%s", ahr.Server(), node), bytes.NewReader(body_bytes))
	// TODO if the node cannot be added a 201 is returned but the node won't appear fully joined in _membership. Handle this situation here.
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if err = ahr.RunRequest(req, nil); err != nil {
		return err
	}

	return cluster.refreshNodesInfo(ahr)
}

func (cluster *Cluster) IsNodeUpAndJoined(node string) bool {
	for _, n := range cluster.NodesInfo.AllNodes {
		if node == n {
			return cluster.knowsNode(node)
		}
	}
	return false
}

func (cluster *Cluster) RemoveNode(node *Node, ahr *httpUtils.AuthenticatedHttpRequester) error {
	var dbs []string
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:5984/_all_dbs", ahr.Server()), nil)
	if err != nil {
		return err
	}

	log.WithField("node", node.Addr()).Info("Checking that node does not own any shard...")
	if err = ahr.RunRequest(req, &dbs); err != nil {
		return err
	}

	for _, db_name := range dbs {
		log.WithFields(log.Fields{"node": node.Addr(), "db": db_name}).Debug("Checking database shards ownership...")
		db, err := LoadDB(db_name, ahr)
		if err != nil {
			return fmt.Errorf("Could not access the %s database", db_name)
		}
		if _, ok := db.config.ByNode[node.Addr()]; ok {
			return fmt.Errorf("Cannot remove %s because it is replicating db %s", node.Addr(), db_name)
		}
	}

	req, err = http.NewRequest("GET", fmt.Sprintf("http://%s:5986/_nodes/%s", ahr.Server(), node.Addr()), nil)
	if err != nil {
		return err
	}

	var nodeInfo struct {
		ID  string `json:"_id"`
		Rev string `json:"_rev"`
	}

	if err = ahr.RunRequest(req, &nodeInfo); err != nil {
		return err
	}

	req, err = http.NewRequest("DELETE", fmt.Sprintf("http://%s:5986/_nodes/%s?rev=%s", ahr.Server(), node.Addr(), nodeInfo.Rev), nil)
	if err != nil {
		return err
	}

	return ahr.RunRequest(req, nil)
}
