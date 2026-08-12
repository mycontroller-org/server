package http_handler

import (
	"net/http"
	"sync"

	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// ListQueryScope returns extra storage filters to AND into list queries (RBAC resource scope).
// Wired from HTTP setup once access control is ready.
type ListQueryScope func(r *http.Request, kind string) (filters []storageTY.Filter, err error)

var (
	listScopeMu sync.RWMutex
	listScope   ListQueryScope
)

// SetListQueryScope registers RBAC list scoping at query level (call at server HTTP init).
func SetListQueryScope(fn ListQueryScope) {
	listScopeMu.Lock()
	listScope = fn
	listScopeMu.Unlock()
}

func applyListQueryScope(r *http.Request, kind string, filters []storageTY.Filter) ([]storageTY.Filter, error) {
	listScopeMu.RLock()
	fn := listScope
	listScopeMu.RUnlock()
	if fn == nil || kind == "" {
		return filters, nil
	}
	extra, err := fn(r, kind)
	if err != nil {
		return filters, err
	}
	if len(extra) == 0 {
		return filters, nil
	}
	return append(filters, extra...), nil
}
