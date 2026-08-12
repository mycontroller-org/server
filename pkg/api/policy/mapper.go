package policy

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
)

// RequestAccess maps an HTTP request to action + resource for authorization.
// name is empty for collection list/create; filled for path id when available.
type RequestAccess struct {
	Action   string
	Resource string // kind or kind:name
	Kind     string
	Name     string // path id or empty (may be UUID - fine for coarse check; handlers may re-check)
	Skip     bool   // if true, authorization middleware skips (auth routes already public)
}

// MapRequest converts HTTP method + path to RBAC action/resource.
func MapRequest(r *http.Request) RequestAccess {
	path := r.URL.Path
	method := r.Method

	// normalize
	path = strings.TrimPrefix(path, "/")
	if !strings.HasPrefix(path, "api/") {
		return RequestAccess{Skip: true}
	}
	path = strings.TrimPrefix(path, "api/")

	// special non-restricted already handled by auth middleware
	if path == "status" || path == "version" || strings.HasPrefix(path, "user/login") ||
		strings.HasPrefix(path, "oauth/") || strings.HasPrefix(path, "plugin/gateway") {
		return RequestAccess{Skip: true}
	}

	// /api/user/profile - own profile (exact path; ids can start with "profile")
	if strings.TrimSuffix(path, "/") == "user/profile" {
		return RequestAccess{Action: policyTY.ActionGet, Kind: policyTY.ResourceUser, Resource: policyTY.ResourceUser, Name: ""}
	}

	// action endpoints
	if strings.HasPrefix(path, "action") {
		return mapActionRoutes(path, method, r)
	}

	// gateway sleeping queue - "clear" mutates state even though it is a GET
	if strings.HasPrefix(path, "gateway-sleeping-queue") {
		action := policyTY.ActionGet
		if strings.Contains(path, "clear") {
			action = policyTY.ActionAction
		}
		return RequestAccess{
			Action:   action,
			Kind:     policyTY.ResourceGateway,
			Resource: policyTY.ResourceGateway,
		}
	}

	// backup / restore
	if strings.HasPrefix(path, "backup") || strings.HasPrefix(path, "restore") {
		action := policyTY.ActionGet
		switch method {
		case http.MethodPost, http.MethodGet:
			if strings.Contains(path, "run") {
				action = policyTY.ActionAction
			} else if method == http.MethodDelete {
				action = policyTY.ActionDelete
			} else if method == http.MethodGet {
				action = policyTY.ActionList
			}
		case http.MethodDelete:
			action = policyTY.ActionDelete
		}
		return RequestAccess{Action: action, Kind: policyTY.ResourceBackup, Resource: policyTY.ResourceBackup}
	}

	// settings - sub resources are named so a read-only grant on the ui settings
	// does not also expose backup location credentials or allow a jwt secret reset
	if strings.HasPrefix(path, "settings") {
		return mapSettingsRoutes(path, method)
	}

	// metric
	if strings.HasPrefix(path, "metric") {
		return RequestAccess{Action: policyTY.ActionGet, Kind: policyTY.ResourceMetric, Resource: policyTY.ResourceMetric}
	}

	// quickid
	if strings.HasPrefix(path, "quickid") {
		return RequestAccess{Action: policyTY.ActionGet, Kind: policyTY.ResourceQuickID, Resource: policyTY.ResourceQuickID}
	}

	// server status
	if strings.HasPrefix(path, "server/status") {
		return RequestAccess{Action: policyTY.ActionGet, Kind: policyTY.ResourceStatus, Resource: policyTY.ResourceStatus}
	}

	// generic /api/{resource}[/{id}|/enable|/disable|/reload|/create|/update|/upload/...]
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		return RequestAccess{Skip: true}
	}

	kind := normalizeKind(parts[0])
	action := methodToAction(method)
	name := ""

	if len(parts) >= 2 {
		sub := parts[1]
		switch sub {
		case "enable":
			action = policyTY.ActionEnable
		case "disable":
			action = policyTY.ActionDisable
		case "reload":
			action = policyTY.ActionReload
		case "create":
			action = policyTY.ActionCreate
		case "update":
			action = policyTY.ActionUpdate
		case "upload":
			action = policyTY.ActionUpdate
			if len(parts) >= 3 {
				name = parts[2]
			}
		default:
			// path id
			if method == http.MethodGet {
				action = policyTY.ActionGet
			}
			name = sub
			// mux vars preferred
			if vars := mux.Vars(r); vars != nil {
				if id, ok := vars["id"]; ok && id != "" {
					name = id
				}
			}
		}
	} else {
		// collection
		switch method {
		case http.MethodGet:
			action = policyTY.ActionList
		case http.MethodPost:
			action = policyTY.ActionUpdate // create-or-update style APIs
		case http.MethodDelete:
			action = policyTY.ActionDelete
		}
	}

	resource := FormatResource(kind, name)
	return RequestAccess{Action: action, Kind: kind, Name: name, Resource: resource}
}

// mapSettingsRoutes names the settings sub resources using their storage keys, so a
// policy can grant "settings:system_settings" (what the console needs) without
// granting "settings:system_backup_locations" (may hold credentials) or the jwt
// secret reset, which is a GET but invalidates every session.
func mapSettingsRoutes(path, method string) RequestAccess {
	settings := func(action, name string) RequestAccess {
		return RequestAccess{
			Action:   action,
			Kind:     policyTY.ResourceSettings,
			Name:     name,
			Resource: FormatResource(policyTY.ResourceSettings, name),
		}
	}

	switch {
	case strings.HasPrefix(path, "settings/system/jwtsecret/reset"):
		return settings(policyTY.ActionUpdate, settingsTY.KeySystemDynamicSecrets)
	case strings.HasPrefix(path, "settings/backuplocations"):
		return settings(policyTY.ActionGet, settingsTY.KeySystemBackupLocations)
	case strings.HasPrefix(path, "settings/system"):
		return settings(policyTY.ActionGet, settingsTY.KeySystemSettings)
	}

	action := policyTY.ActionGet
	if method == http.MethodPost {
		action = policyTY.ActionUpdate
	}
	// collection level: the body names the settings document (checked separately)
	return RequestAccess{Action: action, Kind: policyTY.ResourceSettings, Resource: policyTY.ResourceSettings}
}

func mapActionRoutes(path, method string, r *http.Request) RequestAccess {
	// /api/action/node, /api/action/gateway, /api/action
	if strings.HasPrefix(path, "action/node") {
		ids := r.URL.Query()["id"]
		name := ""
		if len(ids) > 0 {
			name = ids[0]
		}
		return RequestAccess{
			Action:   policyTY.ActionAction,
			Kind:     policyTY.ResourceNode,
			Name:     name,
			Resource: FormatResource(policyTY.ResourceNode, name),
		}
	}
	if strings.HasPrefix(path, "action/gateway") {
		ids := r.URL.Query()["id"]
		name := ""
		if len(ids) > 0 {
			name = ids[0]
		}
		return RequestAccess{
			Action:   policyTY.ActionAction,
			Kind:     policyTY.ResourceGateway,
			Name:     name,
			Resource: FormatResource(policyTY.ResourceGateway, name),
		}
	}
	// generic resource action via quick id in query
	res := r.URL.Query().Get("resource")
	if res != "" {
		// quick id like field:gw.n.s.f
		if i := strings.Index(res, ":"); i > 0 {
			kind := normalizeKind(res[:i])
			return RequestAccess{
				Action:   policyTY.ActionAction,
				Kind:     kind,
				Name:     res[i+1:],
				Resource: FormatResource(kind, res[i+1:]),
			}
		}
	}
	return RequestAccess{
		Action:   policyTY.ActionAction,
		Kind:     policyTY.ResourceAction,
		Resource: policyTY.ResourceAction,
	}
}

func methodToAction(method string) string {
	switch method {
	case http.MethodGet:
		return policyTY.ActionGet
	case http.MethodPost:
		return policyTY.ActionUpdate
	case http.MethodDelete:
		return policyTY.ActionDelete
	case http.MethodPut, http.MethodPatch:
		return policyTY.ActionUpdate
	default:
		return policyTY.ActionGet
	}
}

func normalizeKind(segment string) string {
	return policyTY.NormalizeKind(segment)
}
