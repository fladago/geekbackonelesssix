// В хендлерах работаем только с бизнес логикой. Ничего не знаем про базу данных
// Никаких обращений к пакету usermemstore не будет
//Только непосредственное обращение в бизнес логику
//Роутер это внешний адаптер, который принимает входящие запросы, обрабатывает
//преобразует в нужный вид для того, чтобы передать управление бизнес логике

package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/fladago/gbbackone/app/repos/user"
	"github.com/google/uuid"
)

type Router struct {
	*http.ServeMux
	us *user.Users
}

func NewRouter(us *user.Users) *Router {
	r := &Router{
		ServeMux: http.NewServeMux(),
		us:       us,
	}
	//Любую функцию, которая имеет сигнатуру CreateUser(w http.ResponseWriter, r *http.Request)
	//Превратить в HandlerFunc, и подсовывать везде как хендлер
	//В го у самих функций тоже могут быть методы. Это приведение типов, т.к. сигнатура функции совпадает

	r.HandleFunc("/create", r.AuthMiddleware(http.HandlerFunc(r.CreateUser)).ServeHTTP)
	r.HandleFunc("/read", r.AuthMiddleware(http.HandlerFunc(r.ReadUser)).ServeHTTP)
	r.HandleFunc("/delete", r.AuthMiddleware(http.HandlerFunc(r.DeleteUser)).ServeHTTP)
	r.HandleFunc("/search", r.AuthMiddleware(http.HandlerFunc(r.SearchUser)).ServeHTTP)

	return r
}

//презентабельность для внешних клиентов определяется адаптером
//можно делать в адаптерах  структуры адаптеров
//из нескольких объектов бизнес логики можно собрать в адаптере какое-то представление
//адаптер может вызывать несколько методов бизнес логики
type User struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Data        string    `json:"data"`
	Permissions int       `json:"permissions"`
}

//мидлваре функция проверки на ошибку. Выносим отдельно
//такого вида миделвары создают сторонние роутеры gin, chi
func (rt *Router) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//проверка на авторизацию
		if u, p, ok := r.BasicAuth(); !ok || !(u == "admin" && p == "admin") {
			http.Error(w, "unautofized", http.StatusUnauthorized)
			return
		}

		//Можно создать свой контекст, через который получать id
		// ctx = context.WithValue(r.Context(), CtxIDKey{}, uid)
		// r = r.WithContext(ctx)

		//если все ок, то отправляет все дальше. Так можно выстроить целую цепочку миддлвар,
		//которые что-то делают до того, как в метод получим переменные w,r
		next.ServeHTTP(w, r)
	})
}

//Должны пробросить хэндлеры в бизнес логику в аппликейшен в starter.go
//В реквесте приходит json
//Вначале необходимо проверить авторизацию
func (rt *Router) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	u := User{}
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		//badrequest потому что тело не соответствует формату json
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	//создаем объект бизнес логики
	bu := user.User{
		Name: u.Name,
		Data: u.Data,
	}
	//в request есть свой контекст
	nbu, err := rt.us.Create(r.Context(), bu)
	if err != nil {
		http.Error(w, "error when creating", http.StatusInternalServerError)
		return
	}
	//возвращаем код ответа 200 ок
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(
		User{
			ID:          nbu.ID,
			Name:        nbu.Name,
			Data:        nbu.Data,
			Permissions: nbu.Permissions,
		},
	)

}

//в read получим id
//read?uid=...
func (rt *Router) ReadUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}
	//парсим url
	suid := r.URL.Query().Get("uid")
	if suid == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(suid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//если uid пустой, тоже проверяем на ошибку
	if (uid == uuid.UUID{}) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nbu, err := rt.us.Read(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "error when reading", http.StatusInternalServerError)
		}

		return
	}

	_ = json.NewEncoder(w).Encode(
		User{
			ID:          nbu.ID,
			Name:        nbu.Name,
			Data:        nbu.Data,
			Permissions: nbu.Permissions,
		},
	)
}
func (rt *Router) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	suid := r.URL.Query().Get("uid")
	if suid == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(suid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//если uid пустой, тоже проверяем на ошибку
	if (uid == uuid.UUID{}) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nbu, err := rt.us.Delete(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "error when reading", http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(
		User{
			ID:          nbu.ID,
			Name:        nbu.Name,
			Data:        nbu.Data,
			Permissions: nbu.Permissions,
		},
	)
}

//search?q=...
//При каждом запросе, в этом хэндлере запустится отдельная горутина
func (rt *Router) SearchUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	//Одна проверка. Только если строка пустая, то возвращаем одну ошибку
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	ch, err := rt.us.SearchUsers(r.Context(), q)
	if err != nil {
		http.Error(w, "error when reading", http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	//Под каждый запрос клиента создается отдельная горутина
	//И эта горутина приходит в этот метод через роутер
	//Стримим json поэтому нужно в скобках и через запятую между объектами давать
	first := true
	fmt.Fprintf(w, "[")
	defer fmt.Fprintf(w, "]")
	for {
		select {
		case <-r.Context().Done():
			return
		case u, ok := <-ch:
			//Если закрылся канал, а канал закрылся, если его закрыла бизнес логика
			//а бизнес логика закрылась, если его закрыла база данных
			if !ok {
				return
			}
			if first {
				first = false
			} else {
				fmt.Fprintf(w, ",")
			}
			_ = enc.Encode(
				User{
					ID:          u.ID,
					Name:        u.Name,
					Data:        u.Data,
					Permissions: u.Permissions,
				})
			//Стримим флушером, чанками
			w.(http.Flusher).Flush()
		}
	}
}
