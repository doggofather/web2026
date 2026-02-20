package main

import (
	"context"
	"fmt"
	"os"

	"lab1/internal/app/config"
	"lab1/internal/app/dsn"
	"lab1/internal/app/handler"
	"lab1/internal/app/redis"
	"lab1/internal/app/repository"
	"lab1/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := repository.New(&repository.RepositorySettings{
		PostgresDSN:     postgresString,
		MinioEndpoint:   os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey:  os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey:  os.Getenv("MINIO_SECRET_KEY"),
		MinioBucketName: os.Getenv("MINIO_BUCKET_NAME"),
	})
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	// 1. Инициализируем Redis
	rdb, err := redis.New(context.Background(), conf.Redis) // используем conf из твоего кода
	if err != nil {
		logrus.Fatalf("error initializing redis: %v", err)
	}

	// 2. Передаем и репозиторий, и редис в хендлер
	// ВАЖНО: проверь, что NewHandler принимает два аргумента!
	hand := handler.NewHandler(rep, rdb)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
