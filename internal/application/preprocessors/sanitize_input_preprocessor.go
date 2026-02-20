// internal/application/preprocessors/sanitize_input_preprocessor.go
package preprocessors

import (
	"context"
	"strings"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"go.uber.org/zap"
)

// SanitizeInputPreProcessor limpia y normaliza los inputs
// CQRS: Pre-Processor que se ejecuta ANTES del Handler
type SanitizeInputPreProcessor struct {
	logger *zap.Logger
}

// NewSanitizeInputPreProcessor crea un nuevo preprocessor
func NewSanitizeInputPreProcessor(logger *zap.Logger) *SanitizeInputPreProcessor {
	return &SanitizeInputPreProcessor{
		logger: logger,
	}
}

// Process ejecuta la sanitización
func (p *SanitizeInputPreProcessor) Process(
	ctx context.Context,
	request interface{},
) error {
	// Type assertion para cada tipo de command
	switch cmd := request.(type) {
	case *commands.RegisterUserCommand:
		p.sanitizeRegisterCommand(cmd)
	case commands.RegisterUserCommand:
		// Si viene por valor, no podemos modificarlo
		// Loggear advertencia
		p.logger.Warn("RegisterUserCommand recibido por valor, no se puede sanitizar")
	default:
		// Otros commands no necesitan sanitización (por ahora)
	}

	return nil
}

// sanitizeRegisterCommand limpia el command de registro
func (p *SanitizeInputPreProcessor) sanitizeRegisterCommand(cmd *commands.RegisterUserCommand) {
	// 1. Email: trim, lowercase
	cmd.Email = strings.TrimSpace(cmd.Email)
	cmd.Email = strings.ToLower(cmd.Email)

	// 2. FullName: trim, title case
	cmd.FullName = strings.TrimSpace(cmd.FullName)
	cmd.FullName = p.titleCase(cmd.FullName)

	// 3. Phone: trim, remove common separators
	if cmd.Phone != "" {
		cmd.Phone = strings.TrimSpace(cmd.Phone)
		// Opcional: normalizar formato de teléfono
	}

	// 4. Password: trim (pero mantener case-sensitive)
	cmd.Password = strings.TrimSpace(cmd.Password)

	p.logger.Debug("Input sanitizado",
		zap.String("command", "RegisterUserCommand"),
		zap.String("email", cmd.Email),
	)
}

// ========================================
// HELPER METHODS
// ========================================

// titleCase convierte a Title Case
func (p *SanitizeInputPreProcessor) titleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}
