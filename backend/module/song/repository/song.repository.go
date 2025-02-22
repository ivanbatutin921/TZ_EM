package song_repository

import (
	"context"
	"errors"
	"fmt"
	"root/module/song/dto"

	"gorm.io/gorm"
)

func (r *SongRepository) CheckTable(ctx context.Context, groupName string) (int, error) {
	// Проверяем, существует ли запись с таким именем группы
	existingGroup := new(dto.Group)
	result := r.db.WithContext(ctx).Where("name = ?", groupName).First(existingGroup)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Если запись не найдена, создаем новую
			newGroup := dto.Group{
				Name: groupName, // Используем переданное имя группы
			}
			if err := r.db.WithContext(ctx).Create(&newGroup).Error; err != nil {
				return 0, fmt.Errorf("не удалось создать запись: %w", err)
			}
			// Возвращаем ID созданной записи
			return int(newGroup.ID), nil
		}
		// Если произошла другая ошибка, возвращаем её
		return 0, fmt.Errorf("ошибка при поиске записи: %w", result.Error)
	}

	// Если запись найдена, возвращаем её ID
	return int(existingGroup.ID), nil
}