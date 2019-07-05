## GOsession

GOsession is a basic authenticator for web visitor. This cookie(session)
controller is used third-party dependencies which name is ```go.uuid```
for generating unique id for visitors. The second purpose of this
application is to redirect web requests with authentication. Its main purpose is a minimal library that is not complicated by the root methods but meets the needs. Helps you stay productive.

## Installation

Use the `go` command for installing GOsession and its dependency.

```
go get github.com/satori/go.uuid
go get github.com/dootec/gosession
```

## Requirements

* Third party library [satori/go.uuid](https://github.com/satori/go.uuid)

* UUID dependencies tested against `Go >= 1.6.`

## Quick Review

* GOsession integrate easily to your web project.
* Map structure which include `sync.Mutex` provide healthy concurrency.
* Cookie named session is created by `StartSession(...)` and deleted by
  `StopSession(...)`
* Tracking easily and delete unused sessions with `StopInActiveSessions(...)`
* If you want your visitors to be automatically redirected to path
  related to their roles, you use the `InitGOauthter()` command which
  starts related background job. And now routers can created by
  `SetRouterWithRole()` and `SetRouter()` methods. You use
  `SetRouterWithRole(...)` if you want users with specific roles to
  access it. You can use the `SetRouter(...)` command for the exact
  opposite.
* For detailed information about the user, you can use commands starting with `get`.
* GOauthter supports `*http.ServeMux` methods.

## Details

If the example is not clear enough and you are curious about the details, you can find a description of each command line in the `gosession.go` file. In this way, you can learn how the system works. Good luck! :)

## Example

```golang
package main

import (
	"encoding/json"
	"fmt"
	"github.com/dootec/gosession"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Person struct {
	id       int
	First    string
	Last     string
	Age      int
	Sex      string
	Interest []string
	Address  string
	Phone    string
}

func main() {
	gosession.InitGOsession()
	mux := gosession.InitGOauthter()

	gosession.SetRouter("/", index)
	gosession.SetRouter("/login", login)
	gosession.SetRouter("/logout", logout)

	gosession.SetRouterWithRole("/profile", profile, []string{gosession.RoleAdmin})

	gosession.SetRouter(gosession.PageNotFound, pageNotFound)
	gosession.SetRouter(gosession.PageProtected, pageProtected)

	gosession.SetRouter("/api/getdata", apiGetJson)
	gosession.SetRouter("/api/setdata", apiSetJson)

	http.ListenAndServe(":7070", mux)
}

func index(writer http.ResponseWriter, request *http.Request) {
	outputHtml(writer, `I N D E X - P A G E`)
}

func login(writer http.ResponseWriter, request *http.Request) {
	gosession.StartSession(writer, request, addNewUser(withMyValues()))
	outputHtml(writer, `L O G I N - P A G E`)
}

func logout(writer http.ResponseWriter, request *http.Request) {
	userName := gosession.GetUserName(writer, request)
	gosession.StopSession(writer, request)
	gosession.StopInActiveSessions(writer, request)
	outputHtml(writer, "Have a nice day, "+userName)
}

func profile(writer http.ResponseWriter, request *http.Request) {
	userName := gosession.GetUserName(writer, request)
	uuuid, user, _ := gosession.GetSession(writer, request)
	//user := gosession.GetUser(writer, request) //todo: --> It can be used instead of GetSession for GSUser object.

	p := user.Gsu.(Person)
	And := gosession.And

	_profilepage := "<br>P R O F I L E - P A G E<br>"
	_gsuser := "<br>[U S E R]: " + user.UserName + And + user.Role + And + p.First + And + p.Last + And + strconv.Itoa(p.Age) + And + p.Sex + And + strings.Join(p.Interest, "-") + And + p.Address + And + p.Phone + "<br>"
	_username := "<br>[U S E R N A M E]: " + userName + "<br>"
	_uuid := "<br>[U U I D]: " + uuuid

	outputHtml(writer, _profilepage+_gsuser+_username+_uuid)
}

func pageProtected(writer http.ResponseWriter, request *http.Request) {
	outputHtml(writer, "You must have the necessary privileges to view this page. You can't see this page right now.")
}

func pageNotFound(writer http.ResponseWriter, request *http.Request) {
	outputHtml(writer, "Sorry, there is no web page you are looking for.")
}

func apiGetJson(writer http.ResponseWriter, request *http.Request) {
	jsonValue, err := json.Marshal(addNewUser(withMyValues()))
	if err == nil {
		outputJson(writer, jsonValue)
	}
}

func apiSetJson(writer http.ResponseWriter, request *http.Request) {
	CORS(writer, request)
	var user gosession.GSUser
	err := json.NewDecoder(request.Body).Decode(&user)
	if err == nil {
		fmt.Println("Success: ", user)
	}
}

func CORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, UPDATE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	//w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func addNewUser(dependency interface{}) gosession.GSUser {
	u := &gosession.GSUser{
		UserName: "egegulbahar15@gmail.com",
		Role:     gosession.RoleAdmin + gosession.And + gosession.RoleManager,
		Gsu:      dependency,
	}
	return *u
}

func withMyValues() Person {
	p := Person{
		id:       1,
		First:    "Ege",
		Last:     "Gülbahar",
		Age:      21,
		Sex:      "Male",
		Interest: []string{"Golang", "Angular", "MongoDB", "Docker", "Linux"},
		Address:  "İstanbul/Beylikdüzü",
		Phone:    "000-000-00-00",
	}
	return p
}

func outputHtml(writer http.ResponseWriter, str string) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(writer, str)
}

func outputJson(writer http.ResponseWriter, b []byte) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Write(b)
}
```
