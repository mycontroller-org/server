package policy

import (
	"fmt"
	"net/http"
	"strings"

	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
)

// AuthorizeSleepingQueueRequest checks the gateway/node named in the query.
// A kind-level grant on "gateway" must not be enough to read or clear another
// gateway's queue.
func (a *API) AuthorizeSleepingQueueRequest(subject Subject, r *http.Request) error {
	gatewayID := strings.TrimSpace(r.URL.Query().Get("gatewayId"))
	if gatewayID == "" {
		return ErrAccessDenied
	}
	action := policyTY.ActionGet
	if strings.Contains(strings.TrimSuffix(r.URL.Path, "/"), "/clear") {
		action = policyTY.ActionAction
	}
	nodeID := strings.TrimSpace(r.URL.Query().Get("nodeId"))
	if nodeID != "" {
		resource := FormatResource(policyTY.ResourceNode, gatewayID+"."+nodeID)
		if err := a.Allowed(subject, action, resource); err != nil {
			return fmt.Errorf("%s on %s: %w", action, resource, err)
		}
		return nil
	}
	resource := FormatResource(policyTY.ResourceGateway, gatewayID)
	if err := a.Allowed(subject, action, resource); err != nil {
		return fmt.Errorf("%s on %s: %w", action, resource, err)
	}
	return nil
}
