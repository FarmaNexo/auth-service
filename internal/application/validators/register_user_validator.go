// internal/application/validators/register_user_validator.go
package validators

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/pkg/mediator"
)

// RegisterUserValidator valida el comando RegisterUserCommand
// SOLID: Single Responsibility - Solo valida reglas de negocio
type RegisterUserValidator struct {
	emailRegex     *regexp.Regexp
	minPasswordLen int
	maxNameLen     int
}

// NewRegisterUserValidator crea un nuevo validador
func NewRegisterUserValidator() *RegisterUserValidator {
	return &RegisterUserValidator{
		emailRegex:     regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		minPasswordLen: 8,
		maxNameLen:     255,
	}
}

// Validate ejecuta todas las validaciones
// Implementa: mediator.Validator interface
func (v *RegisterUserValidator) Validate(
	ctx context.Context,
	command commands.RegisterUserCommand,
) error {
	var errors []string

	// 1. Validar Email
	if err := v.validateEmail(command.Email); err != nil {
		errors = append(errors, err.Error())
	}

	// 2. Validar Password
	if err := v.validatePassword(command.Password); err != nil {
		errors = append(errors, err.Error())
	}

	// 3. Validar FullName
	if err := v.validateFullName(command.FullName); err != nil {
		errors = append(errors, err.Error())
	}

	// 4. Validar Phone (opcional pero si viene, debe ser válido)
	if command.Phone != "" {
		if err := v.validatePhone(command.Phone); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// Si hay errores, retornarlos
	if len(errors) > 0 {
		return fmt.Errorf("errores de validación: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ========================================
// VALIDATION RULES
// ========================================

// validateEmail valida el formato del email
func (v *RegisterUserValidator) validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email es requerido")
	}

	email = strings.TrimSpace(email)

	if len(email) > 255 {
		return fmt.Errorf("email no puede exceder 255 caracteres")
	}

	if !v.emailRegex.MatchString(email) {
		return fmt.Errorf("formato de email inválido")
	}

	return nil
}

// validatePassword valida la fortaleza del password
func (v *RegisterUserValidator) validatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password es requerido")
	}

	if len(password) < v.minPasswordLen {
		return fmt.Errorf("password debe tener al menos %d caracteres", v.minPasswordLen)
	}

	if len(password) > 100 {
		return fmt.Errorf("password no puede exceder 100 caracteres")
	}

	// Verificar que tenga al menos una letra y un número
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasLetter {
		return fmt.Errorf("password debe contener al menos una letra")
	}

	if !hasNumber {
		return fmt.Errorf("password debe contener al menos un número")
	}

	return nil
}

// validateFullName valida el nombre completo
func (v *RegisterUserValidator) validateFullName(fullName string) error {
	if fullName == "" {
		return fmt.Errorf("nombre completo es requerido")
	}

	fullName = strings.TrimSpace(fullName)

	if len(fullName) < 3 {
		return fmt.Errorf("nombre completo debe tener al menos 3 caracteres")
	}

	if len(fullName) > v.maxNameLen {
		return fmt.Errorf("nombre completo no puede exceder %d caracteres", v.maxNameLen)
	}

	// Verificar que solo contenga letras, espacios y algunos caracteres especiales
	validName := regexp.MustCompile(`^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s'-]+$`).MatchString(fullName)
	if !validName {
		return fmt.Errorf("nombre completo contiene caracteres inválidos")
	}

	return nil
}

// validatePhone valida el formato del teléfono
func (v *RegisterUserValidator) validatePhone(phone string) error {
	phone = strings.TrimSpace(phone)

	if len(phone) < 7 {
		return fmt.Errorf("teléfono debe tener al menos 7 dígitos")
	}

	if len(phone) > 15 {
		return fmt.Errorf("teléfono no puede exceder 15 dígitos")
	}

	// Permitir solo números, espacios, guiones, paréntesis y +
	validPhone := regexp.MustCompile(`^[\d\s\-\(\)\+]+$`).MatchString(phone)
	if !validPhone {
		return fmt.Errorf("formato de teléfono inválido")
	}

	return nil
}

// ========================================
// COMPILE-TIME INTERFACE CHECK
// ========================================

var _ mediator.Validator[commands.RegisterUserCommand, responses.RegisterResponse] = (*RegisterUserValidator)(nil)
