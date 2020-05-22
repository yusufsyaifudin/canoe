package client

type Config struct {
	RaftServers []RaftServer `json:"raft_servers"`
	PathStat    string       `json:"path_stat"`
	PathJoin    string       `json:"path_join"`
	PathRemove  string       `json:"path_remove"`
}

type RaftServer struct {
	NodeID      string `json:"node_id"`
	RaftAddress string `json:"raft_address"`
	HttpAddress string `json:"http_address"`
}
