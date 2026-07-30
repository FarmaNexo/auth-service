package email

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/farmanexo/auth-service/internal/domain/services"
	"go.uber.org/zap"
)

// SESEmailService envía correos vía Amazon SES usando el SDK de AWS. Las credenciales
// se toman de la cadena por defecto (el rol IAM del task ECS), sin secretos de larga
// vida. Se usa en Dev/Prod; en local se usa SMTP contra Mailpit.
type SESEmailService struct {
	client           *sesv2.Client
	from             string
	frontendURL      string
	configurationSet string
	logger           *zap.Logger
}

// NewSESEmailService crea el servicio de correo sobre SES. configurationSet es opcional:
// si se indica, cada envío se asocia a esa configuration set (tracking de bounces/quejas
// independiente de la config por defecto de la identidad).
func NewSESEmailService(ctx context.Context, region, from, frontendURL, configurationSet string, logger *zap.Logger) (*SESEmailService, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS para SES: %w", err)
	}
	return &SESEmailService{
		client:           sesv2.NewFromConfig(cfg),
		from:             from,
		frontendURL:      frontendURL,
		configurationSet: configurationSet,
		logger:           logger,
	}, nil
}

func (s *SESEmailService) SendPasswordReset(ctx context.Context, toEmail string, resetToken string) error {
	subject, htmlBody, textBody := passwordResetContent(s.frontendURL, resetToken)

	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(s.from),
		Destination: &types.Destination{
			ToAddresses: []string{toEmail},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: aws.String(subject)},
				Body: &types.Body{
					Html: &types.Content{Data: aws.String(htmlBody)},
					Text: &types.Content{Data: aws.String(textBody)},
				},
			},
		},
	}
	if s.configurationSet != "" {
		input.ConfigurationSetName = aws.String(s.configurationSet)
	}

	if _, err := s.client.SendEmail(ctx, input); err != nil {
		s.logger.Error("Error enviando correo de restablecimiento vía SES",
			zap.String("to", toEmail),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Correo de restablecimiento enviado (SES)",
		zap.String("to", toEmail),
	)
	return nil
}

var _ services.EmailService = (*SESEmailService)(nil)
