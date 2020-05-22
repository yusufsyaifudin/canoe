package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"ysf/canoe/pkg/httpclient"
)

type Client struct {
	conf       Config
	httpClient httpclient.HttpRequester
	srvInfo    []*serverInfo
	leader     *serverInfo
}

type dataStat struct {
	State string
}

type respStat struct {
	Data dataStat `json:"data"`
}

type serverInfo struct {
	respStat   respStat
	raftServer RaftServer
}

// sync will do join or remove the server
func (c *Client) sync() {
	ctx := context.Background()
	correlationID := fmt.Sprintf("%d", time.Now().UnixNano())

	// now connect all remaining server to leader (even if they currently act as leader now)
	for _, raftServer := range c.srvInfo {
		if c.leader.raftServer == raftServer.raftServer {
			continue
		}

		joinAddr := fmt.Sprintf("%s%s", c.leader.raftServer.HttpAddress, c.conf.PathJoin)

		body, _ := json.Marshal(map[string]string{
			"node_id":      raftServer.raftServer.NodeID,
			"raft_address": raftServer.raftServer.RaftAddress,
		})

		respHttpJoin, err := c.httpClient.Post(ctx, correlationID, joinAddr, http.Header{
			"Content-Type": []string{"application/json"},
		}, body)

		if err != nil {
			continue
		}

		if respHttpJoin.Raw.StatusCode != http.StatusOK {
			continue
		}
	}
}

func (c Client) SendCommand(data []byte) {
	ctx := context.Background()
	correlationID := fmt.Sprintf("%d", time.Now().UnixNano())

	storeAddr := fmt.Sprintf("%s%s", c.leader.raftServer.HttpAddress, "/store")

	respHttpStore, err := c.httpClient.Post(ctx, correlationID, storeAddr, http.Header{
		"Content-Type": []string{"application/json"},
	}, data)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(respHttpStore.Raw)
	return
}

func (c *Client) leaderElection() {
	ctx := context.Background()
	correlationID := fmt.Sprintf("%d", time.Now().UnixNano())

	var srvInfo = make([]*serverInfo, 0)
	for _, raftServer := range c.conf.RaftServers {
		// get which one the leader
		statAddr := fmt.Sprintf("%s%s", raftServer.HttpAddress, c.conf.PathStat)

		respHttpStat, err := c.httpClient.Get(ctx, correlationID, statAddr, http.Header{
			"Content-Type": []string{"application/json"},
		})

		if err != nil {
			continue
		}

		var data = respStat{}
		err = respHttpStat.To(ctx, &data)
		if err != nil {
			continue
		}

		srvInfo = append(srvInfo, &serverInfo{
			respStat:   data,
			raftServer: raftServer,
		})
	}

	// choose the first leader server as leader
	var leader *serverInfo
	for _, srv := range srvInfo {
		if leader != nil {
			break
		}

		if srv.respStat.Data.State == "Leader" {
			leader = srv
		}
	}

	if leader == nil {
		leader = &serverInfo{}
	}

	c.srvInfo = srvInfo
	c.leader = leader
}

func New(conf Config) *Client {
	var netTransport = &http.Transport{
		TLSHandshakeTimeout: 5 * time.Second,
	}

	var netClient = &http.Client{
		Timeout:   2 * time.Second,
		Transport: netTransport,
	}

	httpRequester := httpclient.DefaultClient(netClient)

	c := &Client{
		conf:       conf,
		httpClient: httpRequester,
	}

	c.leaderElection()
	c.sync()

	return c
}
