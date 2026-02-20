package repository

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"

	"lab1/internal/app/ds"

	"github.com/minio/minio-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (r *Repository) GetSpeakers() ([]ds.Speaker, error) {
	var speakers []ds.Speaker
	err := r.db.Find(&speakers).Error
	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
	if err != nil {
		return nil, err
	}
	if len(speakers) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return speakers, nil
}

func (r *Repository) GetSpeaker(id int) (ds.Speaker, error) {
	speaker := ds.Speaker{}
	err := r.db.Where("id = ?", id).First(&speaker).Error
	if err != nil {
		return ds.Speaker{}, err
	}
	return speaker, nil
}

func (r *Repository) GetSpeakersByTitle(title string) ([]ds.Speaker, error) {
	var speakers []ds.Speaker
	err := r.db.Where("name ILIKE ?", "%"+title+"%").Find(&speakers).Error
	if err != nil {
		return nil, err
	}
	return speakers, nil
}

func (r *Repository) DeleteSpeakerHard(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Сначала удаляем связи в промежуточной таблице
		if err := tx.Where("speaker_id = ?", id).Delete(&ds.SpeakerMeetup{}).Error; err != nil {
			return err
		}

		// 2. Теперь удаляем самого спикера
		if err := tx.Delete(&ds.Speaker{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}
func (r *Repository) AddSpeakerToDraft(userID, speakerID uint) error {
	// Найти или создать черновик для пользователя
	var meetup ds.Meetup
	if err := r.db.Where("creator_id = ? AND status = ?", userID, "draft").First(&meetup).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			meetup = ds.Meetup{
				Title:     "Draft Meetup",
				CreatorID: userID,
			}
			r.db.Create(&meetup)
		} else {
			return err
		}
	}
	// Добавить спикера в м-м таблицу
	sm := ds.SpeakerMeetup{
		SpeakerID: speakerID,
		MeetupID:  meetup.ID,
	}
	return r.db.Create(&sm).Error
}

// Новые методы для Users
func (r *Repository) RegisterUser(user *ds.User) error {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)
	return r.db.Create(user).Error
}

func (r *Repository) AuthenticateUser(login, password string) (*ds.User, error) {
	var user ds.User
	// Query only by login to find the user
	if err := r.db.Where("login = ?", login).First(&user).Error; err != nil {
		// If user not found or any DB error, return a generic error
		return nil, fmt.Errorf("invalid credentials")
	}

	// Now compare the provided plain-text password against the stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// Password doesn't match
		return nil, fmt.Errorf("invalid credentials")
	}

	return &user, nil
}

func (r *Repository) GetUserProfile(id uint) (*ds.User, error) {
	var user ds.User
	return &user, r.db.First(&user, id).Error
}

func (r *Repository) UpdateUserProfile(user *ds.User) error {
	// В кавычках пишем имя колонки из Postgres (id),
	// а вторым аргументом передаем поле из структуры (user.UUID)
	return r.db.Model(&ds.User{}).Where("id = ?", user.UUID).Updates(user).Error
}

func (r *Repository) Register(user *ds.User) error {
	// При uint ID == 0 считается пустым/незаданным.
	// Если в БД стоит auto_increment, можно просто передавать объект,
	// и база сама назначит следующий ID.
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByLogin(login string) (*ds.User, error) {
	var user ds.User

	err := r.db.Where("login = ?", login).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}
func (r *Repository) AddOrReplaceStockImage(ID uint, header *multipart.FileHeader) error {
	// Название будущего файла - <номер акции>.png

	filename := strconv.FormatUint(uint64(ID), 10)

	// Расширение .png будет даже у картинок, исходно его не имевших.
	// Это не совсем хорошо, но отображаться всё будет.

	// Открываем файл
	file, err := header.Open()
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}
	// Как бы мы ни вышли из этой функции, файл обязательно будет закрыт
	defer file.Close()

	// header уже содержит заголовок с типом файла, но для пущей уверенности
	// мы получим Content-Type на базе реального содержимого файла

	// Тип полученного файла хранится в его первых 512 байтах
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)

	if err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	// Определяем тип файла по его содержимому
	contentType := http.DetectContentType(buffer)

	// Позиция чтения файла возвращается к исходной
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("ошибка перемещения по файловому потоку: %w", err)
	}

	_, err = r.Minio.PutObject(
		r.Minio_bucket_name,
		filename,
		file,
		header.Size,
		minio.PutObjectOptions{
			ContentType: contentType,
		})

	if err != nil {
		return fmt.Errorf("не удалось добавить объект в хранилище minio: %w", err)
	}

	// Эндпоинт здесь должен совпадать с эндпоинтом MinIO в конфигурации
	err = r.db.Model(&ds.Speaker{}).Where("id = ?", ID).UpdateColumn("pic", "http://127.0.0.1:9000/"+r.Minio_bucket_name+"/"+filename).Error
	if err != nil {
		// Если не удалось сохранить в БД, удаляем из MinIO
		r.Minio.RemoveObject(r.Minio_bucket_name, filename)
		return fmt.Errorf("ошибка сохранения пути к изображению: %w", err)
	}

	return nil
}

func (r *Repository) CreateSpeaker(speaker *ds.Speaker) error {
	return r.db.Create(speaker).Error
}
func (r *Repository) CreateSpeakerMeetup(binding *ds.SpeakerMeetup) error {
	// 1. Создаем запись
	if err := r.db.Create(binding).Error; err != nil {
		return err
	}

	// 2. Подгружаем связанные данные перед возвратом (Preload)
	return r.db.Preload("Speaker").Preload("Meetup").First(binding, binding.ID).Error
}
func (r *Repository) UpdateSpeakerImage(id uint, url string) error {
	// ЗАМЕНИ "pic" НА "img" (или на то имя, которое у тебя в БД)
	return r.db.Model(&ds.Speaker{}).Where("id = ?", id).Update("img", url).Error
}
