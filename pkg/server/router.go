package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rancher/k3s/pkg/daemons/config"
	"k8s.io/apimachinery/pkg/util/json"
)

type CACertsGetter func() (string, error)

func router(serverConfig *config.Control, tunnel http.Handler, cacertsGetter CACertsGetter) http.Handler {
	authed := mux.NewRouter()
	authed.Use(authMiddleware(serverConfig))
	authed.NotFoundHandler = serverConfig.Runtime.Handler
	authed.Path("/v1-k3s/connect").Handler(tunnel)
	authed.Path("/v1-k3s/node.crt").Handler(nodeCrt(serverConfig))
	authed.Path("/v1-k3s/node.key").Handler(nodeKey(serverConfig))
	authed.Path("/v1-k3s/config").Handler(configHandler(serverConfig))

	router := mux.NewRouter()
	router.NotFoundHandler = authed
	router.Path("/cacerts").Handler(cacerts(cacertsGetter))

	return router
}

func cacerts(getter CACertsGetter) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		content, err := getter()
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(err.Error()))
		}
		resp.Header().Set("content-type", "text/plain")
		resp.Write([]byte(content))
	})
}

func nodeCrt(server *config.Control) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.TLS == nil {
			resp.WriteHeader(http.StatusNotFound)
			return
		}
		http.ServeFile(resp, req, server.Runtime.NodeCert)
	})
}

func nodeKey(server *config.Control) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.TLS == nil {
			resp.WriteHeader(http.StatusNotFound)
			return
		}
		http.ServeFile(resp, req, server.Runtime.NodeKey)
	})
}

func configHandler(server *config.Control) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.TLS == nil {
			resp.WriteHeader(http.StatusNotFound)
			return
		}
		resp.Header().Set("content-type", "application/json")
		json.NewEncoder(resp).Encode(server)
	})
}