//адаптер базы данных. Адаптер исходящего запроса.
package usermemstore

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/fladago/gbbackone/app/repos/user"
	"github.com/google/uuid"
)

//Фишка-проверка. Убеждаемся, что наш тип Users соответствует интерфейсу бизнес логики
//Переменная пустышка, которая не компилируется ни во что.
//Но с точки зрения синтаксического анализа, она позволяет компилятору проверить синтаксис,
//что конструкция справа с набором методов ниже в этом пакете соответствует интерфейсу из бизнес логики
var _ user.UserStore = &Users{}

type Users struct {
	sync.Mutex
	m map[uuid.UUID]user.User
}

func NewUsers() *Users {
	return &Users{
		m: make(map[uuid.UUID]user.User),
	}
}

//В сторе, если заканцелен контекст, мы должны возвращать ошибки
//В случае со сторами, принято использовать локальный контекст для каждого метода.

func (us *Users) Create(ctx context.Context, u user.User) (*uuid.UUID, error) {
	//По хорошему, надо использовать воркеры, и не использовать каналы. надо переделать с
	//мьютекса на каналы. Чем плох мьютекс? Все операции будут в один поток. Они будут разделяться
	//мьютексом. Это табличная блокировка в терминах баз данных.
	//Делать на воркерах и не хранить в мапе. Вместо мапы использовать сигментированные мапы( мапы мап)
	//Сегментировать по ключу
	us.Lock()
	defer us.Unlock()
	select {
	//Если контекст прервался, то вернем нил и ошибку
	case <-ctx.Done():
		return nil, ctx.Err()
	//Здесь ничего не делаем
	default:
	}
	uid := uuid.New()
	u.ID = uid
	us.m[u.ID] = u
	return &uid, nil
}

func (us *Users) Read(ctx context.Context, uid uuid.UUID) (*user.User, error) {
	us.Lock()
	defer us.Unlock()
	select {
	//Если контекст прервался, то вернем нил и ошибку
	case <-ctx.Done():
		return nil, ctx.Err()
	//Здесь ничего не делаем
	default:
	}
	u, ok := us.m[uid]
	if ok {
		return &u, nil
	}
	return nil, sql.ErrNoRows
}

//Не возвращаем ошибку, если не нашли
func (us *Users) Delete(ctx context.Context, uid uuid.UUID) error {
	us.Lock()
	defer us.Unlock()
	select {
	//Если контекст прервался, то вернем нил и ошибку
	case <-ctx.Done():
		return ctx.Err()
	//Здесь ничего не делаем
	default:
	}
	//Если пользователя не было, то нам совершенно не нужно удалять ошибку

	delete(us.m, uid)
	return nil

}

//Стримим юзеров из базы данных
func (us *Users) SearchUsers(ctx context.Context, s string) (chan user.User, error) {
	//FIXME: здесь нужно использовать паттерн Unit of Work
	//Должна быть бизнес транзакция
	us.Lock()
	defer us.Unlock()
	select {
	//Если контекст прервался, то вернем нил и ошибку
	case <-ctx.Done():
		return nil, ctx.Err()
	//Здесь ничего не делаем
	default:
	}
	//FIXME: переделать на дерево остатков
	chout := make(chan user.User, 100)
	go func() {
		defer close(chout)
		//лочимся
		us.Lock()
		defer us.Unlock()
		//перебираем мапу
		for _, u := range us.m {

			if strings.Contains(u.Name, s) {
				//Может залочиться на канале, если заполнили весь буфер, пока кто-нибудь не вычитает
				//Можно пробросить канал done
				//Но здесь сделаем таймаут
				select {
				//либо контекст заканселился
				case <-ctx.Done():
					return
					//если за 2 секунды не смогли вычесть и отправить, то мы выйдем
					//либо таймаут сработал
				case <-time.After(2 * time.Second):
					return
					//либо успешно отправили в канал
				case chout <- u:
				}

			}
		}
	}()
	return chout, nil
}
