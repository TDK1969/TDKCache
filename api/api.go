package api

import (
	mycache "TDKCache/cache"
	"TDKCache/service/http_resp"
	"TDKCache/service/log"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var logger *log.LogEntry

type APIPool struct {
	addr   string
	router *httprouter.Router
}

func NewAPIPool(addr string) *APIPool {
	logger = log.NewLogger("API", fmt.Sprintf("Server <%s>", addr))
	return &APIPool{
		addr:   addr,
		router: registerHandlers(),
	}
}

func registerHandlers() *httprouter.Router {
	router := httprouter.New()

	router.GET("/TDKCache/Get", getGroupKeyHandler)
	router.GET("/TDKCache/Del", deleteGroupKeyHandler)
	return router
}

func getGroupKeyHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	values := r.URL.Query()

	groupName := values.Get("group")
	if groupName == "" {
		logger.Error("lack of necessary param [group]")
		http_resp.SendErrorResponse(w, http_resp.ErrorURLParamsParseFailed)
		return
	}

	key := values.Get("key")
	if key == "" {
		logger.Error("lack of necessary param [key]")
		http_resp.SendErrorResponse(w, http_resp.ErrorURLParamsParseFailed)
		return
	}

	group := mycache.GetGroup(groupName)
	if group == nil {
		logger.Error("no such group: %s", groupName)
		http_resp.SendErrorResponse(w, http_resp.ErrorGroupUnexists)
		return
	}

	logger.Info("%s GET -> get [group] %s | [key] %s", r.RemoteAddr, groupName, key)

	view, err := group.Get(key)
	if err != nil {
		logger.Error("Internal error: %v", err)
		http_resp.SendErrorResponse(w, http_resp.ErrorInternalFaults)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

func deleteGroupKeyHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	values := r.URL.Query()

	groupName := values.Get("group")
	if groupName == "" {
		logger.Error("lack of necessary param [group]")
		http_resp.SendErrorResponse(w, http_resp.ErrorURLParamsParseFailed)
		return
	}

	key := values.Get("key")
	if key == "" {
		logger.Error("lack of necessary param [key]")
		http_resp.SendErrorResponse(w, http_resp.ErrorURLParamsParseFailed)
		return
	}

	group := mycache.GetGroup(groupName)
	if group == nil {
		logger.Error("no such group: %s", groupName)
		http_resp.SendErrorResponse(w, http_resp.ErrorGroupUnexists)
		return
	}

	logger.Info("%s GET -> delete [group] %s | [key] %s", r.RemoteAddr, groupName, key)

	err := group.Delete(key)
	if err != nil {
		logger.Error("Internal error: %v", err)
		http_resp.SendErrorResponse(w, http_resp.ErrorInternalFaults)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write([]byte("ok"))
}

func (p *APIPool) ListenAndServe() error {
	logger.Info("API Server is running at %s", p.addr)
	return http.ListenAndServe(p.addr, p.router)
}
