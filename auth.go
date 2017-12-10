package main

import (
	"net/http"
	"strings"
	"log"
	"fmt"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"
	"crypto/md5"
	"io"
)
import gomniauthcommon "github.com/stretchr/gomniauth/common"

type ChatUser interface {
	UniqueID() string
	AvatarURL() string
}
type chatUser struct {
	gomniauthcommon.User
	uniqueId string
}
func (u chatUser) UniqueID() string {
	return u.uniqueId
}

type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)  {
	if cookie, err := r.Cookie("auth"); err == http.ErrNoCookie ||
		cookie.Value == "" {
		// No Authorization
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		// Occurred some other error
		panic(err.Error())
	} else {
		// Successed. Call Handler that is wrapped.
		h.next.ServeHTTP(w, r)
	}
}

func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

// loginHandlerはサードパーティへのログインへの処理を受け持ちます
// パスの形式： /auth/(action)/(provider)
func loginHandler(w http.ResponseWriter, r *http.Request)  {
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]
	switch action {
	case "login":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln("Failed to get auth provider:", provider, "-", err)
		}
		loginUrl, err := provider.GetBeginAuthURL(nil, nil)
		if err != nil {
			log.Fatalln("Error occurred while calling GetBeginAuthURL:", provider, "-", err)
		}
		w.Header().Set("Location", loginUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
	case "callback":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln("Failed to get auth provider:", provider, "-", err)
		}
		creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		if err != nil {
			log.Fatalln("Couldn't complete authentication", provider, "-", err)
		}
		user, err := provider.GetUser(creds)
		if err != nil {
			log.Fatalln("Failed to GetUser", provider, "-", err)
		}

		chatUser := &chatUser{User: user}
		m := md5.New()
		io.WriteString(m, strings.ToLower(user.Email()))
		chatUser.uniqueId = fmt.Sprintf("%x", m.Sum(nil))
		avatarURL, err := avatars.AvatarURL(chatUser)
		if err != nil {
			log.Fatalln("failed to get AvatarURL", "-", err)
		}

		authCookieValue := objx.New(map[string]interface{}{
			"userid": chatUser.uniqueId,
			"name": user.Name(),
			"avatar_url": avatarURL,
		}).MustBase64()
		http.SetCookie(w, &http.Cookie{
			Name:  "auth",
			Value: authCookieValue,
			Path:  "/"})
		w.Header()["Location"] = []string{"/chat"}
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "No Action %s", action)
	}

}

func logoutHandler(w http.ResponseWriter, r *http.Request)  {
	http.SetCookie(w, &http.Cookie{
		Name: "auth",
		Value: "",
		Path: "/",
		MaxAge: -1,
	})
	w.Header()["Location"] = []string{"/chat"}
	w.WriteHeader(http.StatusTemporaryRedirect)
}