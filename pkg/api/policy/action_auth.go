package policy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	webHandlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
)

// maxActionBodyBytes caps the body we buffer while authorizing action requests.
const maxActionBodyBytes = 1 << 20 // 1 MiB

// AuthorizeActionRequest enforces access for the /api/action* endpoints.
//
// These endpoints take their targets from the query string or the request body,
// so a bare "action" resource in a policy must not be enough - every target is
// checked individually (like quickid), otherwise /api/action becomes a write
// channel into every field/node in the system.
//
//	GET  /api/action/node?id=<uuid>&id=<uuid>   -> action on node:<gatewayId.nodeId> for each id
//	GET  /api/action/gateway?id=<id>            -> action on gateway:<id> for each id
//	GET  /api/action?resource=<quickId>         -> action on the quick id target
//	POST /api/action  [{resource: <quickId>}]   -> action on every quick id in the body
//
// Any denied target fails the whole request.
func (a *API) AuthorizeActionRequest(subject Subject, r *http.Request) error {
	path := strings.TrimSuffix(r.URL.Path, "/")

	switch {
	case strings.HasPrefix(path, "/api/action/node"):
		return a.allowedActionOnIDs(subject, policyTY.ResourceNode, r.URL.Query()["id"])
	case strings.HasPrefix(path, "/api/action/gateway"):
		return a.allowedActionOnIDs(subject, policyTY.ResourceGateway, r.URL.Query()["id"])
	}

	quickIDs := r.URL.Query()[keyResourceParam]
	if r.Method == http.MethodPost {
		bodyQuickIDs, err := a.actionQuickIDsFromBody(r)
		if err != nil {
			return err
		}
		quickIDs = append(quickIDs, bodyQuickIDs...)
	}

	if len(quickIDs) == 0 {
		// no identifiable target: require the generic action capability
		return a.Allowed(subject, policyTY.ActionAction, policyTY.ResourceAction)
	}

	checked := 0
	for _, quickID := range quickIDs {
		quickID = strings.TrimSpace(quickID)
		if quickID == "" {
			continue
		}
		resource, err := ResourceFromQuickID(quickID)
		if err != nil {
			return err
		}
		if err := a.Allowed(subject, policyTY.ActionAction, resource); err != nil {
			return fmt.Errorf("action on %s: %w", quickID, err)
		}
		checked++
	}
	if checked == 0 {
		return a.Allowed(subject, policyTY.ActionAction, policyTY.ResourceAction)
	}
	return nil
}

// keyResourceParam matches routes/action.go
const keyResourceParam = "resource"

// allowedActionOnIDs checks the action verb against each target id. Ids on these
// routes are storage ids, so they are resolved to business names first
// (node -> gatewayId.nodeId) to match the names used in policies.
func (a *API) allowedActionOnIDs(subject Subject, kind string, ids []string) error {
	if len(ids) == 0 {
		return a.Allowed(subject, policyTY.ActionAction, kind)
	}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		name := id
		if biz, err := a.ResolveBusinessName(kind, id); err == nil && biz != "" {
			name = biz
		}
		if err := a.Allowed(subject, policyTY.ActionAction, FormatResource(kind, name)); err != nil {
			return fmt.Errorf("action on %s: %w", FormatResource(kind, name), err)
		}
	}
	return nil
}

// actionQuickIDsFromBody reads the POST /api/action payload and returns the
// quick ids it targets. The body is restored for the handler.
func (a *API) actionQuickIDsFromBody(r *http.Request) ([]string, error) {
	if r.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxActionBodyBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxActionBodyBytes {
		return nil, fmt.Errorf("action payload too large")
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	if len(body) == 0 {
		return nil, nil
	}

	actions := make([]webHandlerTY.ActionConfig, 0)
	if err := json.Unmarshal(body, &actions); err != nil {
		// malformed payload: nothing to target, handler will report the parse error.
		// Fail closed here by requiring the generic action capability.
		return nil, nil
	}
	quickIDs := make([]string, 0, len(actions))
	for _, axn := range actions {
		quickIDs = append(quickIDs, axn.Resource)
	}
	return quickIDs, nil
}
