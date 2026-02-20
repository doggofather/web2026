package repository

import (
	"fmt"

	"lab1/internal/app/ds"
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

/*
func (r *Repository) AddSpeaker(id int) () {
	var speakers []ds.Speaker
	err := r.db.Where("name ILIKE ?", "%"+title+"%").Find(&speakers).Error
	if err != nil {
		return nil, err
	}
	return speakers, nil
}
*/
