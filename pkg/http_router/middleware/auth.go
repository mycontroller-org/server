package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	policyAPI "github.com/mycontroller-org/server/v2/pkg/api/policy"
	"github.com/mycontroller-org/server/v2/pkg/types"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	"github.com/mycontroller-org/server/v2/pkg/types/user"
	handlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	"github.com/mycontroller-org/server/v2/pkg/version"
	"go.uber.org/zap"
)

const (
	contextKey types.ContextKey = "api_context_route"
)

var (
	// middler check vars
	verifyPrefixes = []string{
		"/api/",                                // all api
		handlerTY.SecureShareDirWebHandlerPath, // web file secure share api
	}
	// Unauthenticated endpoints, matched exactly. A prefix match here would also
	// open anything that merely starts with one of these paths - e.g.
	// /api/user/{id} for an id starting with "registration".
	nonRestrictedAPIs = []string{
		"/api/status",            // reports mycontroller server status
		"/api/user/registration", // register new user. TODO: this api not used. verify and remove this
		"/api/user/login",        // login api
		"/api/oauth/login",       // oauth login api
		"/api/oauth/token",       // oauth token api
		"/api/plugin/gateway",    // gateway plugin api
	}
	// Unauthenticated path trees, matched by prefix (they serve directories).
	nonRestrictedPrefixes = []string{
		handlerTY.InsecureShareDirWebHandlerPath, // web file insecure share api
	}

	accessControlMu sync.RWMutex
	accessControl   *policyAPI.API
)

// SetAccessControl wires the policy API used by auth middleware (user disabled, RBAC).
// Call once during server HTTP setup.
func SetAccessControl(api *policyAPI.API) {
	accessControlMu.Lock()
	accessControl = api
	accessControlMu.Unlock()
}

func getAccessControl() *policyAPI.API {
	accessControlMu.RLock()
	defer accessControlMu.RUnlock()
	return accessControl
}

// struct used in api request
type McApiContext struct {
	Tenant         string `json:"tenant" yaml:"tenant"`
	UserID         string `json:"userId" yaml:"userId"`
	ServiceTokenID string `json:"serviceTokenId" yaml:"serviceTokenId"`
}

// MiddlewareAuthenticationVerification verifies user auth details
func MiddlewareAuthenticationVerification(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// allows 'options' method without authentication
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path
		isSecurePrefix := false

		// if signout requested, do signout
		if path == handlerTY.SIGNOUT_PATH {
			doSignOut(w, r)
			handlerUtils.WriteResponse(w, []byte("Signout success"))
			return
		}

		for _, verifyPrefix := range verifyPrefixes {
			if strings.HasPrefix(path, verifyPrefix) {
				isSecurePrefix = true
				break
			}
		}

		if isSecurePrefix {
			if isNonRestricted(path) {
				next.ServeHTTP(w, r)
				return
			}
			// authentication required
			if mcApiContext, err := IsValidToken(r); err == nil {
				// verify user still active (cached) and service token still valid
				if err := verifyPrincipalActive(mcApiContext); err != nil {
					w.Header().Set("Content-Type", "application/json")
					handlerUtils.PostErrorResponse(w, "401 Unauthorized", http.StatusUnauthorized)
					return
				}

				// authorization (RBAC) - lightweight, uses in-memory cache
				if err := authorizeRequest(mcApiContext, r); err != nil {
					w.Header().Set("Content-Type", "application/json")
					handlerUtils.PostErrorResponse(w, "403 Forbidden", http.StatusForbidden)
					return
				}

				ctx := context.WithValue(r.Context(), contextKey, mcApiContext)
				reqWithCtx := r.WithContext(ctx)
				next.ServeHTTP(w, reqWithCtx)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			handlerUtils.PostErrorResponse(w, "401 Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)

	})
}

// isNonRestricted reports whether the path may be served without authentication.
func isNonRestricted(path string) bool {
	trimmed := strings.TrimSuffix(path, "/")
	for _, aPath := range nonRestrictedAPIs {
		if trimmed == strings.TrimSuffix(aPath, "/") {
			return true
		}
	}
	for _, prefix := range nonRestrictedPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func verifyPrincipalActive(mc *McApiContext) error {
	ac := getAccessControl()
	if ac == nil {
		// SetAccessControl runs before any route is served; an unwired middleware
		// cannot authorize anything, so refuse rather than fall open
		return errors.New("access control not initialized")
	}
	if mc.UserID == "" {
		return errors.New("token has no user")
	}
	if _, err := ac.EnsureUserActive(mc.UserID); err != nil {
		return err
	}
	if err := ac.EnsureServiceTokenActive(mc.UserID, mc.ServiceTokenID); err != nil {
		return err
	}
	return nil
}

// isOwnProfilePath reports whether the path is the self-service profile endpoint.
func isOwnProfilePath(path string) bool {
	return strings.TrimSuffix(path, "/") == "/api/user/profile"
}

func authorizeRequest(mc *McApiContext, r *http.Request) error {
	ac := getAccessControl()
	if ac == nil {
		return errors.New("access control not initialized")
	}
	access := policyAPI.MapRequest(r)
	if access.Skip {
		return nil
	}
	// own profile always allowed for get/update of self.
	// exact path only - a prefix match would also skip authorization for
	// /api/user/{id} ids that happen to start with "profile"
	if isOwnProfilePath(r.URL.Path) {
		return nil
	}
	subject := policyAPI.Subject{UserID: mc.UserID, ServiceTokenID: mc.ServiceTokenID}

	// Metrics: enforce per target field/node/gateway (quick_id or body tags.id), not bare "metric"
	if access.Kind == policyTY.ResourceMetric {
		return ac.AuthorizeMetricRequest(subject, r)
	}

	// QuickID: bare "quickid" is not enough — each ?id= target is checked as field:/node:/…
	if access.Kind == policyTY.ResourceQuickID {
		return ac.AuthorizeQuickIDRequest(subject, r)
	}

	// Actions: targets come from the query string or the body, so bare "action"
	// is not enough — every target is checked as node:/gateway:/field:/…
	if strings.HasPrefix(strings.TrimSuffix(r.URL.Path, "/"), "/api/action") {
		return ac.AuthorizeActionRequest(subject, r)
	}

	// Sleeping queue targets are gatewayId/nodeId query params, not the path.
	if strings.HasPrefix(strings.TrimSuffix(r.URL.Path, "/"), "/api/gateway-sleeping-queue") {
		return ac.AuthorizeSleepingQueueRequest(subject, r)
	}

	// Path ids are often storage UUIDs; policies use business names (gatewayId.nodeId...).
	// Resolve entity so get/update/delete by UUID matches the same rules as list.
	if access.Name != "" {
		ac.ResolveResource(&access)
	}
	if err := ac.Allowed(subject, access.Action, access.Resource); err != nil {
		return err
	}

	// Collection endpoints carry their targets in the body (POST /api/gateway,
	// POST /api/gateway/enable, DELETE /api/gateway ...). The check above only
	// established that the principal may reach the endpoint; authorize each
	// object it actually writes.
	return ac.AuthorizeBodyTargets(subject, r, &access)
}

// steps to verify the authentication
// 1. Verify the token in header
// 2. Verify the token in cookie
func IsValidToken(r *http.Request) (*McApiContext, error) {
	token, claims, err := getJwtToken(r)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// verify the validity
	expiresAt := convertor.ToInteger(claims[handlerTY.KeyExpiresAt])
	if time.Now().Unix() >= expiresAt {
		return nil, errors.New("expired token")
	}

	// clear userID / svc token headers, might be injected from external
	r.Header.Del(handlerTY.HeaderUserID)
	r.Header.Del(handlerTY.HeaderServiceTokenID)

	userID := ""
	if v, ok := claims[handlerTY.KeyUserID]; ok {
		if id, ok := v.(string); ok {
			userID = id
			r.Header.Set(handlerTY.HeaderUserID, id)
		}
	}

	svcTokenID := ""
	if v, ok := claims[handlerTY.KeyServiceTokenID]; ok {
		if id, ok := v.(string); ok && id != "" {
			svcTokenID = id
			r.Header.Set(handlerTY.HeaderServiceTokenID, id)
		}
	}

	mcApiContext := McApiContext{
		Tenant:         "",
		UserID:         userID,
		ServiceTokenID: svcTokenID,
	}

	return &mcApiContext, nil
}

func getJwtToken(r *http.Request) (*jwt.Token, jwt.MapClaims, error) {
	tokenString := extractJwtToken(r)
	if tokenString == "" {
		return nil, nil, errors.New("token not supplied")
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		// Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return getJwtSecret(), nil
	})
	if err != nil {
		return nil, nil, err
	}
	return token, claims, err
}

// steps to extract the authentication token
// 1. verify the token in request
// 2. verify the token in header
// 3. verify the token in cookie
func extractJwtToken(r *http.Request) string {
	// 1. verify the token in request
	accessToken := r.URL.Query().Get(handlerTY.AccessToken)
	if accessToken != "" {
		return accessToken
	}

	// 2. verify the token in header
	accessToken = r.Header.Get(handlerTY.HeaderAuthorization)
	if accessToken != "" {
		// normally, Authorization: 'Bearer the_token_xxx'
		if !strings.Contains(accessToken, " ") {
			return accessToken
		}
		strArr := strings.Split(accessToken, " ")
		if len(strArr) == 2 {
			return strArr[1]
		}
	}

	// 3. verify the token in cookie
	authCookie, err := r.Cookie(handlerTY.AUTH_COOKIE_NAME)
	if err == nil {
		return authCookie.Value
	}
	return ""
}

// CreateToken creates a token for a user
func CreateToken(user user.User, expiresIn, svcTokenID string) (string, error) {
	if user.Disabled {
		return "", errors.New("user is disabled")
	}

	atClaims := jwt.MapClaims{}
	atClaims[handlerTY.KeyAuthorized] = true
	atClaims[handlerTY.KeyUserID] = user.ID
	atClaims[handlerTY.KeyFullName] = user.FullName
	atClaims[handlerTY.KeyServiceTokenID] = svcTokenID

	expiresInDuration := handlerTY.DefaultTokenExpiration

	if expiresIn != "" {
		expiresInReceived, err := time.ParseDuration(expiresIn)
		if err == nil {
			expiresInDuration = expiresInReceived
		} else {
			zap.L().Error("error on parse the duration", zap.String("expiration", expiresIn), zap.Error(err))
		}
	}

	atClaims[handlerTY.KeyExpiresAt] = time.Now().Add(expiresInDuration).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString(getJwtSecret())
	if err != nil {
		return "", err
	}
	return token, nil
}

// GetUserID returns the logged in user details
func GetUserID(r *http.Request) string {
	return r.Header.Get(handlerTY.HeaderUserID)
}

// GetServiceTokenID returns service token id from request (if login used a service token)
func GetServiceTokenID(r *http.Request) string {
	return r.Header.Get(handlerTY.HeaderServiceTokenID)
}

// GetAPIContext returns McApiContext from request context if present
func GetAPIContext(r *http.Request) *McApiContext {
	v := r.Context().Value(contextKey)
	if v == nil {
		return nil
	}
	if mc, ok := v.(*McApiContext); ok {
		return mc
	}
	return nil
}

// SubjectFromRequest builds the access control subject from the verified request
// context. Fails when the request did not pass authentication - the mc_userid
// header is only trusted after IsValidToken has rewritten it.
func SubjectFromRequest(r *http.Request) (policyAPI.Subject, error) {
	mc := GetAPIContext(r)
	if mc == nil || mc.UserID == "" {
		return policyAPI.Subject{}, errors.New("unauthenticated request")
	}
	return policyAPI.Subject{UserID: mc.UserID, ServiceTokenID: mc.ServiceTokenID}, nil
}

func getJwtSecret() []byte {
	jwtSeed := types.GetEnvString(types.ENV_JWT_SEED)
	if jwtSeed == "" {
		jwtSeed = version.Get().HostID
	}
	return []byte(fmt.Sprintf("%s_%s", types.GetEnvString(types.ENV_JWT_ACCESS_SECRET), jwtSeed))
}

// doSignout clears the cookies
func doSignOut(w http.ResponseWriter, r *http.Request) {
	// remove cookie and redirect to authentication page
	clearCookie := &http.Cookie{
		Name:   handlerTY.AUTH_COOKIE_NAME,
		Path:   "/",
		Domain: handlerUtils.ExtractHost(r.Host),
		MaxAge: -1,
		Value:  "",
	}
	http.SetCookie(w, clearCookie)
}
