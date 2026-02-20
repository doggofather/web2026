package pkg

import (
	"context"
	"fmt"

	"lab1/internal/app/config"
	"lab1/internal/app/handler"

	"lab1/internal/app/redis"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
	redis   *redis.Client
}

func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler, ctx context.Context) (*Application, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, err
	}
	redisClient, err := redis.New(ctx, cfg.Redis)
	if err != nil {
		return nil, err
	}
	return &Application{
		Config:  c,
		Router:  r,
		Handler: h,
		redis:   redisClient,
	}, nil
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")

	a.Handler.RegisterHandler(a.Router)
	a.Handler.RegisterStatic(a.Router)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}
