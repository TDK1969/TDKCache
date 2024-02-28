package main

import (
	"TDKCache/api"
	mycache "TDKCache/cache"
	http_server "TDKCache/peers/http"
	"flag"
	"fmt"
	"time"
)

var db = map[string]string{
	"Tom":  "630",
	"Tom1": "123",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *mycache.Group {
	return mycache.NewGroup("scores", 2<<10, mycache.GetterFunc(
		func(key string) ([]byte, error) {
			time.Sleep(2 * time.Second)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// 启动CacheServer
func startCacheServer(addr string, addrs []string, g *mycache.Group) {
	p := http_server.NewHTTPPool(addr)
	p.Set(addrs...)
	g.RegisterPeers(p)
	p.ListenAndServe()
}

func main() {
	var port int
	var apiPort int
	flag.IntVar(&port, "port", 58500, "Cache port")
	flag.IntVar(&apiPort, "api", -1, "Frontend API port")
	flag.Parse()

	addrMap := map[int]string{
		58500: "localhost:58500",
		58501: "localhost:58501",
		58502: "localhost:58502",
	}
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	g := createGroup()

	if apiPort != -1 {
		// 开启api服务
		apiAddr := fmt.Sprintf("localhost:%d", apiPort)
		p := api.NewAPIPool(apiAddr)
		go p.ListenAndServe()
	}
	startCacheServer(addrMap[port], []string(addrs), g)

}
