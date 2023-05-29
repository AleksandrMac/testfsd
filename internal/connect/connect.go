package connect

import (
	"net/http"

	"github.com/AleksandrMac/testfsd/internal/datastores"
	"github.com/AleksandrMac/testfsd/internal/queue"
	"github.com/go-chi/chi/v5"
)

// RegisterRouterConnect group for /*
func RegisterRouterConnect(router *chi.Mux, q queue.Queuer, ds datastores.CommonDatastore) {
	router.Route("/connect", func(r chi.Router) {
		c := NewConnector(q, ds)
		// operationAPI := NewOperationAPI(datastores.NewOperationDatastoreExternal(config.Default.Datastore))

		r.Get("/ws", c.WS)
	})
}

// Connector api controller of produces
type Connector interface {
	WS(http.ResponseWriter, *http.Request)
}
