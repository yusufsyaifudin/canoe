# CANOE

It is the KV storage server on top of BadgerDB using raft consensus.

First, run each server using `config.yaml`, then do *manual* join to cluster.
For example, we pick localhost:2222 (raft server localhost:1111 as stated in config.yaml) as the leader server, then do this:

```curl
curl --location --request POST 'localhost:2222/raft/join' \
--header 'Content-Type: application/json' \
--data-raw '{
	"node_id": "node_2", 
	"raft_address": "127.0.0.1:1112"
}'
```

```curl
curl --location --request POST 'localhost:2222/raft/join' \
--header 'Content-Type: application/json' \
--data-raw '{
	"node_id": "node_3", 
	"raft_address": "127.0.0.1:1113"
}'
```

Then ensure that the leader is in localhost:2222 by accessing to http://localhost:2222/raft/stats
You can also do to port 2223 and 2224.

Now, you know that the leader is in port 2222, then you can store and get value using command:

```
curl --location --request GET 'localhost:2222/store/foo'
```

```
curl --location --request POST 'localhost:2222/store' \
--header 'Content-Type: application/json' \
--data-raw '{
	"key": "foo",
	"value": "bar"
}'
```

## Client

Instead joining manually to cluster, we can pick the first server as the leader and connect the rest of server as follower.

Look at the directory `client/example` to see how we can build the raft client. It just get /stats of every server 
and move the request to the leader. This is because Apply command it raft only can be done in Leader server.

