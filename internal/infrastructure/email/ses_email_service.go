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
	client      *sesv2.Client
	from        string
	frontendURL string
	logger      *zap.Logger
}

// NewSESEmailService crea el servicio de correo sobre SES.
func NewSESEmailService(ctx context.Context, region, from, frontendURL string, logger *zap.Logger) (*SESEmailService, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS para SES: %w", err)
	}
	return &SESEmailService{
		client:      sesv2.NewFromConfig(cfg),
		from:        from,
		frontendURL: frontendURL,
		logger:      logger,
	}, nil
}

func (s *SESEmailService) SendPasswordReset(ctx context.Context, toEmail string, resetToken string) error {
	subject, htmlBody := passwordResetContent(s.frontendURL, resetToken)

	_, err := s.client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(s.from),
		Destination: &types.Destination{
			ToAddresses: []string{toEmail},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: aws.String(subject)},
				Body: &types.Body{
					Html: &types.Content{Data: aws.String(htmlBody)},
				},
			},
		},
	})
	if err != nil {
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
