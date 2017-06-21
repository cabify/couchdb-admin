package couchdb_admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cabify/couchdb-admin/httpUtils"
)

type Node struct {
	addr string
}

func NodeAt(addr string) (*Node, error) {
	n := &Node{
		addr: fmt.Sprintf("couchdb@%s", addr),
	}
	return n, nil
}

func (n *Node) IntoMaintenance(ahr *httpUtils.AuthenticatedHttpRequester) error {
	return n.setMaintenanceFlag(true, ahr)
}

func (n *Node) Addr() string {
	return n.addr
}

func (n *Node) DisableMaintenance(ahr *httpUtils.AuthenticatedHttpRequester) error {
	return n.setMaintenanceFlag(false, ahr)
}

func (n *Node) setMaintenanceFlag(value bool, ahr *httpUtils.AuthenticatedHttpRequester) error {
	return n.SetConfig("couchdb", "maintenance_mode", strconv.FormatBool(value), ahr)
}

func (n *Node) SetConfig(section, key, value string, ahr *httpUtils.AuthenticatedHttpRequester) error {
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5984/_node/%s/_config/%s/%s", ahr.Server(), n.addr, section, key),
		strings.NewReader(fmt.Sprintf("\"%s\"", value)))
	if err != nil {
		return err
	}

	return ahr.RunRequest(req, nil)
}
