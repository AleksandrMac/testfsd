package connect

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterRouterConnect group for /*
func RegisterRouterConnect(router *chi.Mux) {
	router.Route("/connect", func(r chi.Router) {
		ws := NewConnectorWS()
		// operationAPI := NewOperationAPI(datastores.NewOperationDatastoreExternal(config.Default.Datastore))

		r.Get("/ws", ws.Listen)
	})
}

// Connector api controller of produces
type Connector interface {
	Listen(http.ResponseWriter, *http.Request)
}
