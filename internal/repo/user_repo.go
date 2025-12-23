package repo

import (
	"context"
	"errors"

	"example.com/pz9-auth/internal/core"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailTaken = errors.New("email already in use")

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) AutoMigrate() error {
	// Сначала проверим существование таблицы
	if !r.db.Migrator().HasTable(&core.User{}) {
		// Таблицы нет - создаем с AutoMigrate
		return r.db.AutoMigrate(&core.User{})
	}

	// Таблица есть - проверим структуру
	if !r.db.Migrator().HasColumn(&core.User{}, "password_hash") {
		// Столбца нет - добавляем без NOT NULL сначала
		if err := r.db.Migrator().AddColumn(&core.User{}, "password_hash"); err != nil {
			return err
		}
		// Обновляем существующие записи
		r.db.Exec("UPDATE users SET password_hash = '' WHERE password_hash IS NULL")
		// Теперь добавляем NOT NULL
		return r.db.Migrator().AlterColumn(&core.User{}, "password_hash")
	}

	return nil
}

func (r *UserRepo) Create(ctx context.Context, u *core.User) error {
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrEmailTaken
		}
		return err
	}
	return nil
}

func (r *UserRepo) ByEmail(ctx context.Context, email string) (core.User, error) {
	var u core.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.User{}, ErrUserNotFound
	}
	return u, err
}
