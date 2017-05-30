package cluster

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/kr/pretty"
)

type Cluster struct {
	AllNodes     []string `json:"all_nodes"`
	ClusterNodes []string `json:"cluster_nodes"`
}

func LoadCluster(server string) (cluster *Cluster) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:5984/_membership", server), nil)
	if err != nil {
		log.Fatal(err)
	}
	runRequest(req, cluster)

	fmt.Println("Current cluster layout")
	fmt.Printf("%# v", pretty.Formatter(cluster))

	return
}

func (cluster *Cluster) knowsNode(node string) bool {
	for n := range cluster.ClusterNodes {
		if node == n {
			return true
		}
	}
	return false
}

func (cluster *Cluster) addNode(server, nodeAddr string) error {
	node := fmt.Sprintf("couchdb@%s", nodeAddr)

	if cluster.knowsNode(node) {
		cluster.Rejoin(node)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5986/_nodes/%s", server, node), strings.NewReader("{}"))

	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	runRequest(req, nil)

	describeCluster()

	return nil
}

func describeCluster() (cluster ClusterLayout) {
}
