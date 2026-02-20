// internal/infrastructure/persistence/postgres/user_repository_impl.go
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/farmanexo/auth-service/internal/domain/entities"
	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserRepositoryImpl implementa UserRepository usando PostgreSQL
// SOLID: Dependency Inversion - Implementa la interfaz del dominio
type UserRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserRepository crea una nueva instancia del repositorio
func NewUserRepository(db *gorm.DB, logger *zap.Logger) repositories.UserRepository {
	return &UserRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// ========================================
// REPOSITORY METHODS
// ========================================

// Create crea un nuevo usuario
func (r *UserRepositoryImpl) Create(ctx context.Context, user *entities.User) error {
	r.logger.Debug("Creando usuario en BD",
		zap.String("email", user.Email),
	)

	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		r.logger.Error("Error creando usuario",
			zap.Error(result.Error),
			zap.String("email", user.Email),
		)
		return fmt.Errorf("error creating user: %w", result.Error)
	}

	r.logger.Info("Usuario creado exitosamente",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
	)

	return nil
}

// FindByEmail busca un usuario por email
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	r.logger.Debug("Buscando usuario por email",
		zap.String("email", email),
	)

	var user entities.User
	result := r.db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.Debug("Usuario no encontrado",
				zap.String("email", email),
			)
			return nil, ErrUserNotFound
		}
		r.logger.Error("Error buscando usuario por email",
			zap.Error(result.Error),
			zap.String("email", email),
		)
		return nil, fmt.Errorf("error finding user by email: %w", result.Error)
	}

	return &user, nil
}

// FindByID busca un usuario por ID
func (r *UserRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	r.logger.Debug("Buscando usuario por ID",
		zap.String("user_id", id.String()),
	)

	var user entities.User
	result := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		r.logger.Error("Error buscando usuario por ID",
			zap.Error(result.Error),
			zap.String("user_id", id.String()),
		)
		return nil, fmt.Errorf("error finding user by id: %w", result.Error)
	}

	return &user, nil
}

// Update actualiza un usuario
func (r *UserRepositoryImpl) Update(ctx context.Context, user *entities.User) error {
	r.logger.Debug("Actualizando usuario",
		zap.String("user_id", user.ID.String()),
	)

	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		r.logger.Error("Error actualizando usuario",
			zap.Error(result.Error),
			zap.String("user_id", user.ID.String()),
		)
		return fmt.Errorf("error updating user: %w", result.Error)
	}

	r.logger.Info("Usuario actualizado exitosamente",
		zap.String("user_id", user.ID.String()),
	)

	return nil
}

// UpdateLoginInfo actualiza información de login
func (r *UserRepositoryImpl) UpdateLoginInfo(ctx context.Context, userID uuid.UUID) error {
	r.logger.Debug("Actualizando info de login",
		zap.String("user_id", userID.String()),
	)

	result := r.db.WithContext(ctx).
		Model(&entities.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at": gorm.Expr("NOW()"),
			"login_count":   gorm.Expr("login_count + 1"),
			"updated_at":    gorm.Expr("NOW()"),
		})

	if result.Error != nil {
		r.logger.Error("Error actualizando login info",
			zap.Error(result.Error),
			zap.String("user_id", userID.String()),
		)
		return fmt.Errorf("error updating login info: %w", result.Error)
	}

	return nil
}

// ExistsByEmail verifica si existe un usuario con el email
func (r *UserRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	r.logger.Debug("Verificando existencia de email",
		zap.String("email", email),
	)

	var count int64
	result := r.db.WithContext(ctx).
		Model(&entities.User{}).
		Where("email = ? AND deleted_at IS NULL", email).
		Count(&count)

	if result.Error != nil {
		r.logger.Error("Error verificando existencia de email",
			zap.Error(result.Error),
			zap.String("email", email),
		)
		return false, fmt.Errorf("error checking email existence: %w", result.Error)
	}

	return count > 0, nil
}

// Delete realiza soft delete del usuario
func (r *UserRepositoryImpl) Delete(ctx context.Context, userID uuid.UUID) error {
	r.logger.Debug("Eliminando usuario (soft delete)",
		zap.String("user_id", userID.String()),
	)

	result := r.db.WithContext(ctx).
		Delete(&entities.User{}, "id = ?", userID)

	if result.Error != nil {
		r.logger.Error("Error eliminando usuario",
			zap.Error(result.Error),
			zap.String("user_id", userID.String()),
		)
		return fmt.Errorf("error deleting user: %w", result.Error)
	}

	r.logger.Info("Usuario eliminado exitosamente",
		zap.String("user_id", userID.String()),
		zap.Int64("rows_affected", result.RowsAffected),
	)

	return nil
}

// ========================================
// DOMAIN ERRORS
// ========================================

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

// ========================================
// COMPILE-TIME INTERFACE CHECK
// ========================================

var _ repositories.UserRepository = (*UserRepositoryImpl)(nil)
