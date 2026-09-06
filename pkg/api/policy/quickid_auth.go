package policy

import (
	"fmt"
	"net/http"
	"strings"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	quickIdUL "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
)

// AuthorizeQuickIDRequest enforces access for GET /api/quickid?id=...
// Bare "quickid" in a policy only opens the API; each requested id is checked
// against the underlying resource (field:/node:/…), including device-tree cascade.
// Any denied id fails the whole request (403).
func (a *API) AuthorizeQuickIDRequest(subject Subject, r *http.Request) error {
	ids := r.URL.Query()["id"]
	if len(ids) == 0 {
		// no targets: require generic quickid access
		return a.Allowed(subject, policyTY.ActionGet, policyTY.ResourceQuickID)
	}

	// Optional coarse gate: if the principal has neither quickid nor any device access,
	// fail fast. Prefer per-id checks so "quickid" alone cannot leak denied fields.
	for _, quickID := range ids {
		quickID = strings.TrimSpace(quickID)
		if quickID == "" {
			continue
		}
		res, err := ResourceFromQuickID(quickID)
		if err != nil {
			return err
		}
		if err := a.Allowed(subject, policyTY.ActionGet, res); err != nil {
			return fmt.Errorf("quickid %s: %w", quickID, err)
		}
	}
	return nil
}

// ResourceFromQuickID maps a quick id (e.g. field:gw.n.s.f) to a policy resource string.
func ResourceFromQuickID(quickID string) (string, error) {
	rt, kv, err := quickIdUL.EntityKeyValueMap(quickID)
	if err != nil {
		return "", err
	}

	switch strings.ToLower(rt) {
	case quickIdUL.QuickIdGateway:
		return FormatResource(policyTY.ResourceGateway, kv[types.KeyGatewayID]), nil
	case quickIdUL.QuickIdNode:
		return FormatResource(policyTY.ResourceNode, joinIDs(kv[types.KeyGatewayID], kv[types.KeyNodeID])), nil
	case quickIdUL.QuickIdSource:
		return FormatResource(policyTY.ResourceSource, joinIDs(kv[types.KeyGatewayID], kv[types.KeyNodeID], kv[types.KeySourceID])), nil
	case quickIdUL.QuickIdField:
		return FormatResource(policyTY.ResourceField, joinIDs(
			kv[types.KeyGatewayID], kv[types.KeyNodeID], kv[types.KeySourceID], kv[types.KeyFieldID],
		)), nil
	case quickIdUL.QuickIdTask:
		return FormatResource(policyTY.ResourceTask, firstNonEmpty(kv[types.KeyID], kv["id"])), nil
	case quickIdUL.QuickIdSchedule:
		return FormatResource(policyTY.ResourceSchedule, firstNonEmpty(kv[types.KeyID], kv["id"])), nil
	case quickIdUL.QuickIdHandler:
		return FormatResource(policyTY.ResourceHandler, firstNonEmpty(kv[types.KeyID], kv["id"])), nil
	case quickIdUL.QuickIdFirmware:
		return FormatResource(policyTY.ResourceFirmware, firstNonEmpty(kv[types.KeyID], kv["id"])), nil
	case quickIdUL.QuickIdDataRepository, "datarepository":
		return FormatResource(policyTY.ResourceDataRepository, firstNonEmpty(kv[types.KeyID], kv["id"])), nil
	case quickIdUL.QuickIdForwardPayload, "forwardpayload":
		return FormatResource(policyTY.ResourceForwardPayload, firstNonEmpty(kv[types.KeyID], kv["id"])), nil
	default:
		// unknown: fall back to kind:rest so custom kinds still gate somehow
		if i := strings.Index(quickID, ":"); i > 0 {
			return FormatResource(normalizeKind(quickID[:i]), quickID[i+1:]), nil
		}
		return "", fmt.Errorf("unsupported quick id type: %s", rt)
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
