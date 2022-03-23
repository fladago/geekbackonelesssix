package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/fladago/geekbackonelesssix/api/handler"
	"github.com/fladago/geekbackonelesssix/api/server"
	"github.com/fladago/geekbackonelesssix/app/repos/user"
	"github.com/fladago/geekbackonelesssix/app/starter"
	"github.com/fladago/geekbackonelesssix/db/mem/usermemstore"
)

func main() {
	//Реализуем грейсфул шатдаун
	//Создадим глобальный стартовый контекст относительно бекграунд контекста
	//Он будет прерываем по ctrl+c
	//Поскольку все контексты будут наследоваться от этого контекста, то ctrl+c прервет программу
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	//Заканцелем контекст при выходе из майна, хотя никогда не будет вызван
	defer cancel()

	ust := usermemstore.NewUsers()
	//Передаем стор в app
	a := starter.NewApp(ust)
	us := user.NewUsers(ust)
	h := handler.NewRouter(us)
	srv := server.NewServer(":8000", h)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	//Пробрасываем контекст, чтобы отловить сигналы из операционной системы
	//Также передаем сервер
	go a.Serve(ctx, wg, srv)
	<-ctx.Done()
	//канцелим контекст, потом дожидаемся всех горутин
	cancel()
	wg.Wait()
}
