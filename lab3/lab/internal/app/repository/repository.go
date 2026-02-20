package repository

import (
	"github.com/minio/minio-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db                *gorm.DB
	Minio             *minio.Client // С большой буквы!
	Minio_bucket_name string
}

type RepositorySettings struct {
	PostgresDSN     string
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucketName string
}

func New(settings *RepositorySettings) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(settings.PostgresDSN), &gorm.Config{}) // подключаемся к БД
	if err != nil {
		return nil, err
	}

	// Используем простой New без Options
	minioClient, err := minio.New(
		settings.MinioEndpoint,
		settings.MinioAccessKey,
		settings.MinioSecretKey,
		false, // SSL off
	)
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:                db,
		Minio:             minioClient, // Убедись, что M большая
		Minio_bucket_name: settings.MinioBucketName,
	}, nil
}
