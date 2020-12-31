package gorilla

import (
	"net/http"
	"xdf/common/log"

	"github.com/kataras/neffos"

	gorilla "github.com/gorilla/websocket"
)

// DefaultUpgrader is a gorilla/websocket Upgrader with all fields set to the default values.
var DefaultUpgrader = Upgrader(gorilla.Upgrader{
	// v12中的BUG
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
})

// Upgrader is a `neffos.Upgrader` type for the gorilla/websocket subprotocol implementation.
// Should be used on `New` to construct the neffos server.
func Upgrader(upgrader gorilla.Upgrader) neffos.Upgrader {
	return func(w http.ResponseWriter, r *http.Request) (neffos.Socket, error) {
		underline, err := upgrader.Upgrade(w, r, w.Header())
		if err != nil {
			return nil, err
		}
		s, err := newSocket(underline, r, false)
		if err != nil {
			log.Error(err)
		}
		return s, err
	}
}
