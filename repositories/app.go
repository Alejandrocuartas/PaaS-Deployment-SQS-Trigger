package repositories

import (
	"PaaS-deployment-sqs-trigger/db"
	"PaaS-deployment-sqs-trigger/models"
)

func UpdateApp(app *models.App) error {
	if err := db.PGDB.Save(app).Error; err != nil {
		return err
	}
	return nil
}

func GetAppByUuid(uuid string) (*models.App, error) {
	var app models.App
	if err := db.PGDB.
		Where("uuid = ?", uuid).
		Where("deleted_at IS NULL").
		First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}
