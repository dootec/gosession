package gosession

import (
	"github.com/satori/go.uuid"
	"net/http"
	"sync"
	"time"
)

// Reminding for Cookie:
//    err == <nil> : Cookie had been already created.
//    err != <nil> : There is no Cookie which is created.
//
// Reminding for Role:
//    User.Role must be like this "ROLE_ADMIN,ROLE_PRESIDENT,ROLE_BOARD_MEMBER"

const ( //todo: You can change internal parameters except "PageNotFound" and "PageProtected" for your application
	SessionTime        = 1800 //30min (second), It's time for valid session which is every single consumer/session
	ControlSessionTime = 45   //(second), It's time for trigger/controlling session.
)

var gs GoSession

type GoSession struct {
	sync.Mutex
	lastControlSessionTime time.Time
	users                  map[string]GSUser
	sessions               map[string]GSSession
}

type GSUser struct {
	UserName string      `json: "username"`
	Role     string      `json: "role"`
	Gsu      interface{} `json: "gsu"`
}

type GSSession struct {
	userName     string
	lastActivity time.Time
}

func InitGOsession() {
	gs = GoSession{lastControlSessionTime: time.Now(), users: make(map[string]GSUser), sessions: make(map[string]GSSession),}
}

//Creates a cookie called session and assign it to the user object.
func StartSession(w http.ResponseWriter, r *http.Request, u GSUser) {
	uID, _ := uuid.NewV4()
	cookie, err := r.Cookie("session")
	if err == nil {
		//Value of cookie which named session, is checking
		_, err := uuid.FromString(cookie.Value)
		if err != nil {
			//invalid uuid(Cookie/Session Value) is detected by *Satori* and value is changed
			delMaps(cookie.Value, u)
			cookie.Value = uID.String()
			cookie.MaxAge = SessionTime
			http.SetCookie(w, cookie)
		}
		//System already have a session now. Checking harmony between uuid and RAM(dbUsers, dbSessions)
		if !checkDatabases(cookie.Value) {
			//RAM is cleared, now system have a cookie but RAM cleared by 'checkDatabases'(internal command)
			//fmt.Println("içerideyiz", uID.String())
			cookie.Value = uID.String()
			cookie.MaxAge = SessionTime
			http.SetCookie(w, cookie)
			addMaps(cookie.Value, u)
			//OK, everything is fine. Session value(uuid) is valid and RAM and Session are pointed by each other.
		} else {
			//OK, everything is fine. Session value(uuid) is valid and RAM and Session are pointed by each other.
		}
	} else {
		//System has no any cookie which named session and everything is created from A to Z
		//In this command, RAM isn't check because 1 session can point 1 user object.
		//but
		//1 user object can pointed more than one session(uuid)
		//
		//Why?: User have mobile, desktop, tablet devices which can login by.
		createSessionCookie(w, r, uID.String())
		addMaps(uID.String(), u)
		//OK, everything is fine. Session value(uuid) is valid and RAM and Session are pointed by each other.
	}
}

func StopSession(w http.ResponseWriter, r *http.Request) {
	cookie, ok := DeleteCookie(w, r, "session")
	if ok {
		s, _ := gs.sessions[cookie.Value]
		delMaps(cookie.Value, gs.users[s.userName])
	}
}

func StopInActiveSessions(w http.ResponseWriter, r *http.Request) {
	if time.Now().Sub(gs.lastControlSessionTime) > (time.Second * time.Duration(ControlSessionTime)) { //It's temporary time for low cpu cost(resist to looping)
		gs.lastControlSessionTime = time.Now()
		for uuid, session := range gs.sessions {
			if time.Now().Sub(session.lastActivity) > (time.Second * time.Duration(SessionTime)) { //Condition control all "LastActivity" of session object. If duration greater than "SessionTime", system deletes RAMs which are pointed by cookie value
				user := gs.users[session.userName]
				delMaps(uuid, user)
				//Also We can use DeleteCookie(...) but cookie isn't valid now. In this case no need to delete session.
			}
		}
	}
}

// just create a cookie;
// true means cookie created succesfully;
// false means cookie isn't created
func CreateCookie(w http.ResponseWriter, r *http.Request, n string, v string) bool {
	cookie, err := r.Cookie(n)
	if err != nil {
		cookie = &http.Cookie{
			Name:  n,
			Value: v,
		}
		http.SetCookie(w, cookie)
		return true;
	}
	return false;
}

// just delete a cookie;
// true means cookie deleted successfully;
// false means cookie isn't deleted
func DeleteCookie(w http.ResponseWriter, r *http.Request, name string) (*http.Cookie, bool) {
	cookie, err := r.Cookie(name)
	if err == nil {
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)
		return cookie, true
	}
	return cookie, false
}

//This method similar to createCookie but there is differences.
//Properties of cookie are predefined which are Name and HttpOnly
//Name is defaultly "session
//HttpOnly is defaulty "true"
func createSessionCookie(w http.ResponseWriter, r *http.Request, value string) {
	cookie, err := r.Cookie("session")
	if err != nil {
		cookie = &http.Cookie{
			Name:     "session",
			MaxAge:   SessionTime,
			Value:    value,
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
	}
}

func addMaps(uID string, u GSUser) {
	gs.Lock()
	defer gs.Unlock()
	s := &GSSession{
		userName:     u.UserName,
		lastActivity: time.Now(),
	}
	gs.users[u.UserName] = u
	gs.sessions[uID] = *s
}

func delMaps(uID string, u GSUser) {
	gs.Lock()
	defer gs.Unlock()
	delete(gs.users, u.UserName)
	delete(gs.sessions, uID)
}

func checkDatabases(uID string) bool {
	if s, okSession := gs.sessions[uID]; okSession {
		var u GSUser;
		var okUser bool
		if u, okUser = gs.users[s.userName]; okUser {
			return true
		}
		//Scenario:
		//Unique id(uuid) which is key, is pointed by dbSession even though the uuid has no relation with user object.
		//This is wrong because all databases have to valid and check by each other.
		//User object must be pointed by uuid and uuid must be pointed too by user object.
		//
		//In this situation, What will we do?
		//User object and uuid must be consistent in each other. If object or uuid can't point, all data which is related must be deleted from ram.
		delMaps(uID, u)
	}
	return false
}

// if session, ram memory(dbUser, dbSession) are created and matched with together, method will return back as true
// true means user logged seccesful
// false means either cookie isn't created or cookie is created but ram memory isn't created so mechanism is not authenticated each other.
func AlreadyLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("session")
	if err != nil {
		return false
	}
	s := gs.sessions[cookie.Value]
	_, ok := gs.users[s.userName]
	return ok
}

// if session exist, method will return all details which are session status, session unique uuid and object of user.
// true means session name which is uuid match exact with ram memories which are dbUsers and dbSessions
// false means either session isn't created yet or session uuid and ram aren't matched
func GetSession(w http.ResponseWriter, r *http.Request) (string, GSUser, bool) {
	var u GSUser
	cookie, err := r.Cookie("session")
	if err == nil {
		if s, ok := gs.sessions[cookie.Value]; ok {
			u = gs.users[s.userName]
			return cookie.Value, u, true
		}
	}
	return "", u, false
}

// if cookie exist which name is session, method will return back session uuid
// nil-null-"" means cookie which name is session isn't created yet
func GetSessionValue(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("session")
	if err == nil {
		return cookie.Value
	}
	return ""
}

// If there is a cookie named session, the method returns its User object of username.
// nil-null-"" means cookie which name is session isn't created yet
func GetUserName(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("session")
	if err == nil {
		if s, ok := gs.sessions[cookie.Value]; ok {
			return s.userName
		}
	}
	return ""
}

// If there is a cookie named session, the method returns its User object.
// false means cookie which name is session isn't created yet
func GetUser(w http.ResponseWriter, r *http.Request) GSUser {
	var u GSUser
	cookie, err := r.Cookie("session")
	if err == nil {
		if s, ok := gs.sessions[cookie.Value]; ok {
			u = gs.users[s.userName]
			return u
		}
	}
	return u
}
