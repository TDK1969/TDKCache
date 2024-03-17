package main

import (
	"TDKCache/api"
	mycache "TDKCache/cache"
	"TDKCache/peers"
	"TDKCache/peers/rpc"
	"flag"
	"fmt"
	"time"
)

var (
	db = map[string]string{
		"Tom":  "630",
		"Tom1": "123",
		"Jack": "589",
		"Sam":  "567",
	}
	s peers.PeerServer
)

func createGroup() *mycache.Group {
	return mycache.NewGroup("scores", 2<<10, mycache.GetterFunc(
		func(key string) ([]byte, error) {
			// 模拟慢查询
			time.Sleep(2 * time.Second)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func main() {
	var serverPort int
	var apiPort int
	flag.IntVar(&serverPort, "port", 58500, "Cache port")
	flag.IntVar(&apiPort, "api", -1, "Frontend API port")
	flag.Parse()

	g := createGroup()

	if apiPort != -1 {
		// 开启api服务
		apiAddr := fmt.Sprintf(":%d", apiPort)
		p := api.NewAPIPool(apiAddr)
		go p.ListenAndServe()
	}

	//s = http_server.NewHTTPPool(addrMap[port])
	s = rpc.NewRPCServer(serverPort)
	s.Start(g)
}
