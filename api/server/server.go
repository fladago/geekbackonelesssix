//Уровень адаптеров
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/fladago/gbbackone/app/repos/user"
)

//Должен быть сервер, который принимает запросы и вызывает бизнес логику
//Входящий адаптер обращается в бизнес логику. Все строго по стандартам гексогональной архитектуры.
type Server struct {
	srv http.Server
	us  *user.Users
}

//Создаем экземпляр структуры
func NewServer(addr string, h http.Handler) *Server {
	s := &Server{}
	s.srv = http.Server{
		Addr:    addr,
		Handler: h,
		//Важно прописывать таймауты. В дефолтном сервере таймауты не установлены и все соединения
		//будут висеть бесконечно, если бесконечно висячий клиент
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
		//Могут быть заполнены базовые контексты. В них могут находиться логеры и трейсеры.
	}
	return s
}

//Делаем два метода
//Надо стартовать и остановить сервер
//Стартуем с контекстом
func (s *Server) Start(us *user.Users) {
	s.us = us
	go s.srv.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	s.srv.Shutdown(ctx)
	cancel()
}
