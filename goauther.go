package gosession

import (
	"net/http"
	"strings"
)

const (
	PageNotFound  = "PAGE_NOT_FOUND"
	PageProtected = "PAGE_PROTECTED"
	RoleAdmin     = "ROLE_ADMIN"
	RoleManager   = "ROLE_MANAGER"
	RoleMember    = "ROLE_MEMBER"
	And           = ", "
)

var ga GoAuther

type GoAuther struct {
	mux           *http.ServeMux
	router        map[string]http.HandlerFunc
	auth          map[string]GSAuth
	pageNotFound  http.HandlerFunc
	pageProtected http.HandlerFunc
}

type GSAuth struct {
	url  string
	role []string
}

func InitGOauthter() *http.ServeMux {
	mux := http.NewServeMux()
	ga = GoAuther{mux: mux, router: make(map[string]http.HandlerFunc), auth: make(map[string]GSAuth),}
	mux.HandleFunc("/", authers)
	return mux
}

func authers(w http.ResponseWriter, r *http.Request) {
	handlerFunc, ok := ga.router[r.URL.Path]
	if ok {
		if ControlAuthorization(w, r) {
			handlerFunc(w, r)
		} else {
			if ga.pageProtected != nil {
				ga.pageProtected(w, r)
			}
		}
	} else {
		if ga.pageNotFound != nil {
			ga.pageNotFound(w, r)
		}
	}
}

//User's which is logged, role check with path's role. If user role can appropriate, method will return back as true.
func ControlAuthorization(w http.ResponseWriter, r *http.Request) bool {
	auth, ok := ga.auth[r.URL.String()]
	if ok {
		user := GetUser(w, r)
		for _, role := range auth.role {
			if strings.Contains(user.Role, role) {
				return true;
			}
		}
		return false
	}
	return true;
}

//If paths settled in this way, it can be viewed by everyone.
func SetRouter(url string, handlerFunc http.HandlerFunc) {
	switch url {
	case PageNotFound:
		ga.pageNotFound = handlerFunc
	case PageProtected:
		ga.pageProtected = handlerFunc
	default:
	}
	ga.router[url] = handlerFunc
}

//Access is allowed if the logged in user's role information matches the role information of the url path.
//If paths are settled in this way, user can be seen if has appropriate role..
func SetRouterWithRole(url string, handlerFunc http.HandlerFunc, role []string) {
	gsa := &GSAuth{
		url:  url,
		role: role,
	}
	ga.auth[url] = *gsa
	ga.router[url] = handlerFunc
}
