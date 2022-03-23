// В хендлерах работаем только с бизнес логикой. Ничего не знаем про базу данных

// Никаких обращений к пакету usermemstore не будет

//Только непосредственное обращение в бизнес логику

//Роутер это внешний адаптер, который принимает входящие запросы, обрабатывает

//преобразует в нужный вид для того, чтобы передать управление бизнес логике

package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fladago/geekbackonelesssix/app/repos/user"
	"github.com/fladago/geekbackonelesssix/db/mem/usermemstore"
)

func TestRouter_CreateUser(t *testing.T) {
	ust := usermemstore.NewUsers()
	us := user.NewUsers(ust)
	rt := NewRouter(us)
	h := rt.AuthMiddleware(http.HandlerFunc(rt.CreateUser)).ServeHTTP

	w := &httptest.ResponseRecorder{}
	r := httptest.NewRequest("POST", "/create", strings.NewReader(`{"name":"user1"}`))
	r.SetBasicAuth("admin", "admin")
	h(w, r)
	if w.Code != http.StatusCreated {
		t.Error("status wrong")
	}

	//Настройка клиента
	//(&http.Client{}).Get(httptest.NewServer(nil).URL)
	//Сам клиент можно применять в любом пакете уровня внешнего адаптера
}
