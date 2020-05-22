package main

import (
	"encoding/json"
	"ysf/canoe/client"
)

func main() {
	conf := client.Config{
		RaftServers: []client.RaftServer{
			{
				NodeID:      "node_1",
				RaftAddress: "localhost:1111",
				HttpAddress: "http://localhost:2222",
			},
			{
				NodeID:      "node_2",
				RaftAddress: "localhost:1112",
				HttpAddress: "http://localhost:2223",
			},
			{
				NodeID:      "node_3",
				RaftAddress: "localhost:1113",
				HttpAddress: "http://localhost:2224",
			},
		},
		PathStat:   "/raft/stats",
		PathJoin:   "/raft/join",
		PathRemove: "",
	}

	body, _ := json.Marshal(map[string]string{
		"key":   "foo",
		"value": "bar",
	})

	c := client.New(conf)
	c.SendCommand(body)
}
