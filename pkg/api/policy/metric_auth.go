package policy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	quickIdUL "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
	mtsTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// AuthorizeMetricRequest enforces metric access for GET (quick_id) or POST (body queries).
// Resource patterns use kind "metric" with the same hierarchical names as fields:
//
//	metric:*                          all metrics
//	metric:home-gw.*                  all fields under gateway
//	metric:home-gw.living-room.*      all under node
//	metric:home-gw.n.s.*              all under source
//	metric:home-gw.n.s.temperature    one field
//
// Built-in policies that list bare "metric" still allow all (kind-only match).
func (a *API) AuthorizeMetricRequest(subject Subject, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		return a.authorizeMetricGET(subject, r)
	case http.MethodPost:
		return a.authorizeMetricPOST(subject, r)
	default:
		return a.Allowed(subject, policyTY.ActionGet, policyTY.ResourceMetric)
	}
}

func (a *API) authorizeMetricGET(subject Subject, r *http.Request) error {
	quickID := r.URL.Query().Get(QuickIDParam)
	if quickID == "" {
		// alternate casing used in some clients
		quickID = r.URL.Query().Get("quickId")
	}
	if quickID == "" {
		// no target: require unrestricted metric API
		return a.Allowed(subject, policyTY.ActionGet, policyTY.ResourceMetric)
	}
	path, err := a.fieldPathFromQuickID(quickID)
	if err != nil {
		return err
	}
	return a.AllowedMetricPath(subject, path)
}

func (a *API) authorizeMetricPOST(subject Subject, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	// restore body for the handler
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	if len(body) == 0 {
		return a.Allowed(subject, policyTY.ActionGet, policyTY.ResourceMetric)
	}

	queryConfig := &mtsTY.QueryConfig{}
	if err := json.Unmarshal(body, queryConfig); err != nil {
		// let handler report bad JSON; only require generic metric access
		return a.Allowed(subject, policyTY.ActionGet, policyTY.ResourceMetric)
	}

	if len(queryConfig.Individual) == 0 {
		return a.Allowed(subject, policyTY.ActionGet, policyTY.ResourceMetric)
	}

	for i := range queryConfig.Individual {
		q := queryConfig.Individual[i]
		path, err := a.fieldPathFromMetricQuery(&q)
		if err != nil {
			return err
		}
		if path == "" {
			// no identifiable field: require full metric access
			if err := a.Allowed(subject, policyTY.ActionGet, policyTY.ResourceMetric); err != nil {
				return fmt.Errorf("metric query %d: %w", i, err)
			}
			continue
		}
		if err := a.AllowedMetricPath(subject, path); err != nil {
			return fmt.Errorf("metric query %d (%s): %w", i, path, err)
		}
	}
	return nil
}

// AllowedMetricPath checks get on metric:<fieldPath> (hierarchical match).
func (a *API) AllowedMetricPath(subject Subject, fieldPath string) error {
	if fieldPath == "" {
		return a.Allowed(subject, policyTY.ActionGet, policyTY.ResourceMetric)
	}
	return a.Allowed(subject, policyTY.ActionGet, FormatResource(policyTY.ResourceMetric, fieldPath))
}

// QuickIDParam matches routes/metric.go
const QuickIDParam = "quick_id"

func (a *API) fieldPathFromQuickID(quickID string) (string, error) {
	rt, kvMap, err := quickIdUL.EntityKeyValueMap(quickID)
	if err != nil {
		return "", err
	}
	switch rt {
	case quickIdUL.QuickIdField:
		return joinIDs(
			kvMap[types.KeyGatewayID],
			kvMap[types.KeyNodeID],
			kvMap[types.KeySourceID],
			kvMap[types.KeyFieldID],
		), nil
	case quickIdUL.QuickIdSource:
		return joinIDs(kvMap[types.KeyGatewayID], kvMap[types.KeyNodeID], kvMap[types.KeySourceID]), nil
	case quickIdUL.QuickIdNode:
		return joinIDs(kvMap[types.KeyGatewayID], kvMap[types.KeyNodeID]), nil
	case quickIdUL.QuickIdGateway:
		return kvMap[types.KeyGatewayID], nil
	default:
		return "", fmt.Errorf("metric resource type not supported: %s", rt)
	}
}

func (a *API) fieldPathFromMetricQuery(q *mtsTY.Query) (string, error) {
	if q == nil || q.Tags == nil {
		return "", nil
	}
	// Prefer explicit hierarchy tags if present
	gw := firstTag(q.Tags, types.KeyGatewayID, "gatewayId", "GatewayID")
	node := firstTag(q.Tags, types.KeyNodeID, "nodeId", "NodeID")
	src := firstTag(q.Tags, types.KeySourceID, "sourceId", "SourceID")
	field := firstTag(q.Tags, types.KeyFieldID, "fieldId", "FieldID")
	if gw != "" && node != "" && src != "" && field != "" {
		return joinIDs(gw, node, src, field), nil
	}
	if gw != "" && node != "" && src != "" {
		return joinIDs(gw, node, src), nil
	}
	if gw != "" && node != "" {
		return joinIDs(gw, node), nil
	}
	if gw != "" && field == "" && node == "" {
		return gw, nil
	}

	// UI usually sends tags.id = field storage UUID
	id := firstTag(q.Tags, types.KeyID, "id", "ID")
	if id == "" {
		return "", nil
	}
	var e fieldTY.Field
	err := a.storage.FindOne(types.EntityField, &e, []storageTY.Filter{{Key: types.KeyID, Value: id}})
	if err != nil {
		return "", fmt.Errorf("metric field id %s: %w", id, err)
	}
	return BusinessName(policyTY.ResourceField, e), nil
}

func firstTag(tags map[string]string, keys ...string) string {
	for _, k := range keys {
		if v, ok := tags[k]; ok && v != "" {
			return v
		}
	}
	return ""
}
