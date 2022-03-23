//Бизнес логика не должна знать апи, который относится к внешнему адаптеру, поэтому делаем интерфейс
//Слой выделяется отдельно, так как может быть зашита бизнес логика по аркестрации запросов отдельно
//Вообще это считается аппликейшен уровнем над уровнем бизнес логики.
//Чуть выше уровня бизнес логики
package starter

import (
	"context"
	"sync"

	"github.com/fladago/gbbackone/app/repos/user"
)

type App struct {
	us *user.Users
}

func NewApp(ust user.UserStore) *App {
	a := &App{
		us: user.NewUsers(nil),
	}
	return a
}

type HTTPServer interface {
	Start(us *user.Users)
	Stop(ctx context.Context)
}

//Нужно создать все объекты бизнес логики. Как минимум, все объекты, связанные с бизнес логикой
//Необходимо добавить api, которое будет посылать запросы
func (a *App) Serve(ctx context.Context, wg *sync.WaitGroup, hs HTTPServer) {
	defer wg.Done()
	//Стартуем сервер
	hs.Start(a.us)
	//Ждем, пока контекст заканселится. Грейсфулшатдаун
	//В мейне ничего не делаем. Все делается в бизнес логике
	<-ctx.Done()
	//Стопаем сервер
	hs.Stop(context.Background())

}
