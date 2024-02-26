package http_server

import (
	mycache "TDKCache/cache"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func registerHandlers() *httprouter.Router {
	router := httprouter.New()

	router.GET("/TDKCache/Get", getGroupKeyHandler)

	return router
}

func getGroupKeyHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	values := r.URL.Query()

	groupName := values.Get("group")
	if groupName == "" {
		hsLogger.Error("lack of necessary param [group]")
		SendErrorResponse(w, ErrorURLParamsParseFailed)
		return
	}

	key := values.Get("key")
	if key == "" {
		hsLogger.Error("lack of necessary param [key]")
		SendErrorResponse(w, ErrorURLParamsParseFailed)
		return
	}

	group := mycache.GetGroup(groupName)
	if group == nil {
		hsLogger.Error("no such group: %s", groupName)
		SendErrorResponse(w, ErrorGroupUnexists)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		hsLogger.Error("Internal error: %v", err)
		SendErrorResponse(w, ErrorInternalFaults)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

func (p *HTTPPool) ListenAndServe() error {
	hsLogger.Info("TDKCache is running at %s", p.addr)
	return http.ListenAndServe(p.addr, p.router)
}
