package http_server

import (
	mycache "TDKCache/cache"
	"TDKCache/peers/protobuf/pb"
	"TDKCache/service/http_resp"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"google.golang.org/protobuf/proto"
)

func registerHandlers() *httprouter.Router {
	router := httprouter.New()
	router.GET("/TDKCache/PBGet", pbGetGroupKeyHandler)
	return router
}

func pbGetGroupKeyHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	values := r.URL.Query()

	groupName := values.Get("group")
	if groupName == "" {
		hsLogger.Error("lack of necessary param [group]")

		http_resp.SendErrorResponse(w, http_resp.ErrorURLParamsParseFailed)
		return
	}

	key := values.Get("key")
	if key == "" {
		hsLogger.Error("lack of necessary param [key]")
		http_resp.SendErrorResponse(w, http_resp.ErrorURLParamsParseFailed)
		return
	}

	group := mycache.GetGroup(groupName)
	if group == nil {
		hsLogger.Error("no such group: %s", groupName)
		http_resp.SendErrorResponse(w, http_resp.ErrorGroupUnexists)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		hsLogger.Error("Internal error: %v", err)
		http_resp.SendErrorResponse(w, http_resp.ErrorInternalFaults)
		return
	}

	// 将得到的view编码为protobuf响应
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		hsLogger.Error("Encoding response error: %v", err)
		http_resp.SendErrorResponse(w, http_resp.ErrorInternalFaults)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

func (p *HTTPPool) ListenAndServe() error {
	hsLogger.Info("TDKCache is running at %s", p.addr)
	return http.ListenAndServe(p.addr, p.router)
}
