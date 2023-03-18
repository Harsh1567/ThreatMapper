package reporters_graph

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/deepfence/golang_deepfence_sdk/utils/directory"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j/dbtype"
)

type ThreatGraphReporter struct {
	driver neo4j.Driver
}

func NewThreatGraphReporter(ctx context.Context) (*ThreatGraphReporter, error) {
	driver, err := directory.Neo4jClient(ctx)

	if err != nil {
		return nil, err
	}

	nc := &ThreatGraphReporter{
		driver: driver,
	}

	return nc, nil
}

const (
	CLOUD_AWS     = "aws"
	CLOUD_AZURE   = "azure"
	CLOUD_GCP     = "gcp"
	CLOUD_PRIVATE = "others"
)

var CLOUD_ALL = [...]string{CLOUD_AWS, CLOUD_AZURE, CLOUD_GCP, CLOUD_PRIVATE}

func (tc *ThreatGraphReporter) GetThreatGraph() (ThreatGraph, error) {
	aggreg, err := tc.GetRawThreatGraph()
	if err != nil {
		return ThreatGraph{}, err
	}

	all := ThreatGraph{}
	for _, cp := range CLOUD_ALL {
		resources := []ThreatNodeInfo{}
		node_info := aggreg[cp].getNodeInfos()
		depths := aggreg[cp].nodes_depth
		if _, has := depths[1]; !has {
			goto end
		}
		for _, root := range depths[1] {
			visited := map[int64]struct{}{}
			attack_paths := build_attack_paths(aggreg[cp], root, visited)
			paths := [][]string{}
			for _, Attack_path := range attack_paths {
				path := []string{}
				for i := range Attack_path {
					index := int64(len(Attack_path)-1) - int64(i)
					path = append(path, node_info[index].Id)
				}
				paths = append(paths, append([]string{"The Internet"}, path...))
				entry := ThreatNodeInfo{
					Label:                 node_info[int64(len(Attack_path)-1)].Label,
					Id:                    node_info[int64(len(Attack_path)-1)].Id,
					Nodes:                 node_info[int64(len(Attack_path)-1)].Nodes,
					Vulnerability_count:   node_info[int64(len(Attack_path)-1)].Vulnerability_count,
					Secrets_count:         node_info[int64(len(Attack_path)-1)].Secrets_count,
					Compliance_count:      node_info[int64(len(Attack_path)-1)].Compliance_count,
					CloudCompliance_count: node_info[int64(len(Attack_path)-1)].CloudCompliance_count,
					Count:                 node_info[int64(len(Attack_path)-1)].Count,
					Node_type:             node_info[int64(len(Attack_path)-1)].Node_type,
					Attack_path:           paths,
				}
				resources = append(resources, entry)
			}
		}
	end:
		all[cp] = ProviderThreatGraph{
			Resources:             resources,
			Compliance_count:      0,
			Secrets_count:         0,
			Vulnerability_count:   0,
			CloudCompliance_count: 0,
		}
	}

	return all, nil
}

func build_attack_paths(paths AttackPaths, root int64, visited map[int64]struct{}) [][]int64 {
	if _, has := visited[root]; has {
		return [][]int64{}
	}
	visited[root] = struct{}{}
	if _, has := paths.nodes_data[root]; !has {
		return [][]int64{}
	}
	if _, has := paths.nodes_tree[root]; !has {
		return [][]int64{{root}}
	}
	res := [][]int64{}
	for _, edge := range paths.nodes_tree[root] {
		edge_paths := build_attack_paths(paths, edge, visited)
		for _, edge_path := range edge_paths {
			res = append(res, append([]int64{root}, edge_path...))
		}
	}
	if len(res) == 0 {
		return [][]int64{{root}}
	}
	return res
}

func (tc *ThreatGraphReporter) GetRawThreatGraph() (map[string]AttackPaths, error) {
	session, err := tc.driver.Session(neo4j.AccessModeRead)

	if err != nil {
		return nil, err
	}
	defer session.Close()

	tx, err := session.BeginTransaction()
	if err != nil {
		return nil, err
	}
	defer tx.Close()

	all := map[string]AttackPaths{}
	for _, cloud_provider := range CLOUD_ALL {
		var res neo4j.Result
		if cloud_provider != CLOUD_PRIVATE {
			if res, err = tx.Run(`
				CALL apoc.nodes.group(['CloudResource','Node'], ['node_type', 'depth', 'cloud_provider'],
				[{`+"`*`"+`: 'count', sum_cve: 'sum', sum_secrets: 'sum', sum_compliance: 'sum', sum_cloud_compliance: 'sum',
				node_id:'collect', num_cve: 'collect', num_secrets:'collect', num_compliance:'collect', num_cloud_compliance: 'collect'},
				{`+"`*`"+`: 'count'}], {selfRels: false})
				YIELD node, relationships
				WHERE apoc.any.property(node, 'depth') IS NOT NULL
				AND apoc.any.property(node, 'cloud_provider') = '`+cloud_provider+`'
				RETURN node, relationships
				`, map[string]interface{}{}); err != nil {
			}
		} else {
			if res, err = tx.Run(`
				CALL apoc.nodes.group(['Node'], ['node_type', 'depth', 'cloud_provider'],
				[{`+"`*`"+`: 'count', sum_cve: 'sum', sum_secrets: 'sum', sum_compliance: 'sum', sum_cloud_compliance: 'sum',
				node_id:'collect', num_cve: 'collect', num_secrets:'collect', num_compliance:'collect', num_cloud_compliance:'collect'},
				{`+"`*`"+`: 'count'}], {selfRels: false})
				YIELD node, relationships
				WHERE apoc.any.property(node, 'depth') IS NOT NULL
				AND NOT apoc.any.property(node, 'cloud_provider') IN ['aws', 'gcp', 'azure']
				AND apoc.any.property(node, 'cloud_provider') <> 'internet'
				RETURN node, relationships
				`, map[string]interface{}{}); err != nil {
			}
		}

		if err != nil {
			return nil, err
		}

		records, err := res.Collect()
		if err != nil {
			return nil, err
		}

		nodes_tree := map[int64][]int64{}
		nodes_data := map[int64]AttackPathData{}
		nodes_depth := map[int64][]int64{}
		for _, record := range records {
			record_node, _ := record.Get("node")
			record_relationships, _ := record.Get("relationships")
			node := record_node.(dbtype.Node)
			node_datum := record2struct(node)
			nodes_data[node.Id] = node_datum

			for _, rel_node := range record_relationships.([]interface{}) {
				rel := rel_node.(dbtype.Relationship)
				nodes_tree[node.Id] = append(nodes_tree[node.Id], rel.EndId)

			}
			nodes_depth[node_datum.depth] = append(nodes_depth[node_datum.depth], node.Id)
		}

		all[cloud_provider] = AttackPaths{
			nodes_tree:  nodes_tree,
			nodes_data:  nodes_data,
			nodes_depth: nodes_depth,
		}
	}

	return all, nil
}

type AttackPaths struct {
	nodes_tree  map[int64][]int64
	nodes_data  map[int64]AttackPathData
	nodes_depth map[int64][]int64
}

func record2struct(node dbtype.Node) AttackPathData {

	record := node.Props
	Node_type, _ := record["node_type"]
	depth, _ := record["depth"]
	cloud_provider, _ := record["cloud_provider"]
	sum_sum_cve_, _ := record["sum_sum_cve"]
	sum_sum_secrets_, _ := record["sum_sum_secrets"]
	sum_sum_compliance_, _ := record["sum_sum_compliance"]
	sum_sum_cloud_compliance_, _ := record["sum_sum_cloud_compliance"]
	node_count, _ := record["count_*"]
	collect_node_id_, _ := record["collect_node_id"]
	collect_num_cve_, _ := record["collect_num_cve"]
	collect_num_secrets_, _ := record["collect_num_secrets"]
	collect_num_compliance_, _ := record["collect_num_compliance"]
	collect_num_cloud_compliance_, _ := record["collect_num_cloud_compliance"]

	collect_node_id := []string{}
	for _, v := range collect_node_id_.([]interface{}) {
		collect_node_id = append(collect_node_id, v.(string))
	}

	collect_num_cve := []int64{}
	sum_sum_cve := int64(0)
	if collect_num_cve_ != nil {
		for _, v := range collect_num_cve_.([]interface{}) {
			collect_num_cve = append(collect_num_cve, v.(int64))
		}

		sum_sum_cve, _ = sum_sum_cve_.(int64)
	}

	collect_num_secrets := []int64{}
	sum_sum_secrets := int64(0)
	if collect_num_secrets_ != nil {
		for _, v := range collect_num_secrets_.([]interface{}) {
			collect_num_secrets = append(collect_num_secrets, v.(int64))
		}

		sum_sum_secrets = sum_sum_secrets_.(int64)
	}

	collect_num_compliance := []int64{}
	sum_sum_compliance := int64(0)
	if collect_num_compliance_ != nil {
		for _, v := range collect_num_compliance_.([]interface{}) {
			collect_num_compliance = append(collect_num_compliance, v.(int64))
		}
		sum_sum_compliance = sum_sum_compliance_.(int64)
	}

	collect_num_cloud_compliance := []int64{}
	sum_sum_cloud_compliance := int64(0)
	if collect_num_cloud_compliance_ != nil {
		for _, v := range collect_num_cloud_compliance_.([]interface{}) {
			collect_num_cloud_compliance = append(collect_num_cloud_compliance, v.(int64))
		}
		sum_sum_cloud_compliance = sum_sum_cloud_compliance_.(int64)
	}

	return AttackPathData{
		Node_type:                    Node_type.(string),
		cloud_provider:               cloud_provider.(string),
		depth:                        depth.(int64),
		sum_sum_cve:                  sum_sum_cve,
		sum_sum_secrets:              sum_sum_secrets,
		sum_sum_compliance:           sum_sum_compliance,
		sum_sum_cloud_compliance:     sum_sum_cloud_compliance,
		node_count:                   node_count.(int64),
		collect_node_id:              collect_node_id,
		collect_num_cve:              collect_num_cve,
		collect_num_secrets:          collect_num_secrets,
		collect_num_compliance:       collect_num_compliance,
		collect_num_cloud_compliance: collect_num_cloud_compliance,
	}
}

type AttackPathData struct {
	identity                     int64
	Node_type                    string
	cloud_provider               string
	depth                        int64
	sum_sum_cve                  int64
	sum_sum_secrets              int64
	sum_sum_compliance           int64
	sum_sum_cloud_compliance     int64
	node_count                   int64
	collect_node_id              []string
	collect_num_cve              []int64
	collect_num_secrets          []int64
	collect_num_compliance       []int64
	collect_num_cloud_compliance []int64
}

func getThreatNodeId(apd AttackPathData) string {
	h := sha256.New()
	v := []string{}
	for i := range apd.collect_node_id {
		v = append(v, apd.collect_node_id[i])
	}
	sort.Strings(v)

	for _, s := range v {
		h.Write([]byte(s))
	}

	return hex.EncodeToString(h.Sum(nil))
}

func (ap AttackPaths) getNodeInfos() map[int64]ThreatNodeInfo {
	res := map[int64]ThreatNodeInfo{}
	for _, v := range ap.nodes_data {
		var Label, Id string
		Id = getThreatNodeId(v)
		switch v.Node_type {
		case "host":
			Label = "Compute Instance"
		case "container":
			Label = "Container"
		case "internet":
			Label = "The Internet"
			Id = "The Internet"
		default:
			Label = "CloudResource"
		}
		Nodes := map[string]NodeInfo{}
		for i, Node_id := range v.collect_node_id {
			vuln_count := int64(0)
			if len(v.collect_num_cve) == len(v.collect_node_id) {
				vuln_count = v.collect_num_cve[i]
			}
			secrets_count := int64(0)
			if len(v.collect_num_secrets) == len(v.collect_node_id) {
				secrets_count = v.collect_num_secrets[i]
			}
			compliance_count := int64(0)
			if len(v.collect_num_compliance) == len(v.collect_node_id) {
				compliance_count = v.collect_num_compliance[i]
			}
			cloud_compliance_count := int64(0)
			if len(v.collect_num_cloud_compliance) == len(v.collect_node_id) {
				cloud_compliance_count = v.collect_num_cloud_compliance[i]
			}
			Nodes[Node_id] = NodeInfo{
				Node_id:                 v.collect_node_id[i],
				Image_name:              "",
				Name:                    Node_id,
				Vulnerability_count:     vuln_count,
				Vulnerability_scan_id:   "",
				Secrets_count:           secrets_count,
				Secrets_scan_id:         "",
				Compliance_count:        compliance_count,
				Compliance_scan_id:      "",
				CloudCompliance_count:   cloud_compliance_count,
				CloudCompliance_scan_id: "",
			}
		}
		res[v.identity] = ThreatNodeInfo{
			Label:                 Label,
			Id:                    Id,
			Nodes:                 Nodes,
			Vulnerability_count:   v.sum_sum_cve,
			Secrets_count:         v.sum_sum_secrets,
			Compliance_count:      v.sum_sum_compliance,
			CloudCompliance_count: v.sum_sum_cloud_compliance,
			Count:                 int64(len(v.collect_node_id)),
			Node_type:             v.Node_type,
			Attack_path:           [][]string{},
		}
	}
	return res
}

type ThreatGraph map[string]ProviderThreatGraph

type ProviderThreatGraph struct {
	Resources             []ThreatNodeInfo `json:"resources" required:"true"`
	Compliance_count      int64            `json:"compliance_count" required:"true"`
	Secrets_count         int64            `json:"secrets_count" required:"true"`
	Vulnerability_count   int64            `json:"vulnerability_count" required:"true"`
	CloudCompliance_count int64            `json:"cloud_compliance_count" required:"true"`
}

type ThreatNodeInfo struct {
	Label string              `json:"label" required:"true"`
	Id    string              `json:"id" required:"true"`
	Nodes map[string]NodeInfo `json:"nodes" required:"true"`

	Vulnerability_count   int64 `json:"vulnerability_count" required:"true"`
	Secrets_count         int64 `json:"secrets_count" required:"true"`
	Compliance_count      int64 `json:"compliance_count" required:"true"`
	CloudCompliance_count int64 `json:"cloud_compliance_count" required:"true"`
	Count                 int64 `json:"count" required:"true"`

	Node_type string `json:"node_type" required:"true"`

	Attack_path [][]string `json:"attack_path" required:"true"`
}

type NodeInfo struct {
	Node_id                 string `json:"node_id" required:"true"`
	Image_name              string `json:"image_name" required:"true"`
	Name                    string `json:"name" required:"true"`
	Vulnerability_count     int64  `json:"vulnerability_count" required:"true"`
	Vulnerability_scan_id   string `json:"vulnerability_scan_id" required:"true"`
	Secrets_count           int64  `json:"secrets_count" required:"true"`
	Secrets_scan_id         string `json:"secrets_scan_id" required:"true"`
	Compliance_count        int64  `json:"compliance_count" required:"true"`
	Compliance_scan_id      string `json:"compliance_scan_id" required:"true"`
	CloudCompliance_count   int64  `json:"cloud_compliance_count" required:"true"`
	CloudCompliance_scan_id string `json:"cloud_compliance_scan_id" required:"true"`
}
