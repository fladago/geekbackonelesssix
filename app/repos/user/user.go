package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

//1.Создаем объект
type User struct {
	ID          uuid.UUID
	Name        string
	Data        string
	Permissions int
}

// Создаем интерфейс системы хранения.
// Отделяем слой юзерстора от бизнес логики.
// Бизнес логика ничего не знает как в реальности хранятся данные.
type UserStore interface {
	//Создаем пользователя. На выходе получаем id созданного юзера и ошибку
	Create(ctx context.Context, u User) (*uuid.UUID, error)
	Read(ctx context.Context, uid uuid.UUID) (*User, error)
	Delete(ctx context.Context, uid uuid.UUID) error
	SearchUsers(ctx context.Context, s string) (chan User, error)
}

//2. Создаем коллекцию объектов, чтобы реализовать паттерн репозитория
//Работает с системой хранения
type Users struct {
	ustore UserStore
}

//Система инициализации со стором. Конструктор экземпляра класса
func NewUsers(ustore UserStore) *Users {
	return &Users{
		ustore: ustore,
	}
}

//3. Реализовываем несколько методов.
//В качестве параметра передаем пустой экземпляр пользователя
func (us *Users) Create(ctx context.Context, u User) (*User, error) {
	//Что-то вызываем в системе хранения, чтобы сохранить юзера
	id, err := us.ustore.Create(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("Create user error: %w", err)
	}
	u.ID = *id
	return &u, nil
}
func (us *Users) Read(ctx context.Context, uid uuid.UUID) (*User, error) {
	//
	u, err := us.ustore.Read(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("Read user error: %w", err)
	}

	return u, nil
}

func (us *Users) Delete(ctx context.Context, uid uuid.UUID) (*User, error) {
	//Сначало достаем пользователя через Read
	u, err := us.ustore.Read(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("Search user error: %w", err)
	}

	return u, us.ustore.Delete(ctx, uid)
}

//Стримим юзеров
func (us *Users) SearchUsers(ctx context.Context, s string) (chan User, error) {
	chin, err := us.ustore.SearchUsers(ctx, s)
	if err != nil {
		return nil, err
	}
	chout := make(chan User, 100)
	go func() {
		defer close(chout)
		for {
			//Пишем через селект, чтобы по заканцеленному контексту прервать обработку.
			select {
			//Прерываем контекстом бесконечный цикл for
			case <-ctx.Done():
				return
			//Чтение из закрытого канала вызывает появление пустого юзера
			case u, ok := <-chin:
				if !ok {
					return
				}
				u.Permissions = 0755
				chout <- u
			}

		}
	}()
	return chout, nil
}
