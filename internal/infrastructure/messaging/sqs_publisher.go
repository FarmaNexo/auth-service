// internal/infrastructure/messaging/sqs_publisher.go
package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/farmanexo/auth-service/internal/domain/events"
	"github.com/farmanexo/auth-service/internal/domain/services"
	"github.com/farmanexo/auth-service/pkg/config"
	"go.uber.org/zap"
)

// SQSEventPublisher implementa EventPublisher usando AWS SQS
type SQSEventPublisher struct {
	sqsClient *sqs.Client
	queueURL  string
	logger    *zap.Logger
}

// NewSQSEventPublisher crea una nueva instancia de SQSEventPublisher
func NewSQSEventPublisher(
	awsCfg config.AWSConfig,
	sqsCfg config.SQSConfig,
	logger *zap.Logger,
) (*SQSEventPublisher, error) {
	// Opciones para cargar la configuración AWS
	optFns := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(awsCfg.Region),
	}

	// Si hay endpoint configurado (LocalStack), usar credenciales estáticas fake
	if awsCfg.Endpoint != "" {
		optFns = append(optFns, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		))
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), optFns...)
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	// Crear cliente SQS con endpoint personalizado si corresponde
	sqsOptFns := []func(*sqs.Options){}
	if awsCfg.Endpoint != "" {
		sqsOptFns = append(sqsOptFns, func(o *sqs.Options) {
			o.BaseEndpoint = &awsCfg.Endpoint
		})
	}

	sqsClient := sqs.NewFromConfig(cfg, sqsOptFns...)

	logger.Info("SQS EventPublisher inicializado",
		zap.String("region", awsCfg.Region),
		zap.String("queue_url", sqsCfg.AuthEventsQueueURL),
		zap.String("endpoint", awsCfg.Endpoint),
	)

	return &SQSEventPublisher{
		sqsClient: sqsClient,
		queueURL:  sqsCfg.AuthEventsQueueURL,
		logger:    logger,
	}, nil
}

// Publish publica un evento de autenticación en SQS
func (p *SQSEventPublisher) Publish(ctx context.Context, event events.AuthEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    &p.queueURL,
		MessageBody: stringPtr(string(body)),
	}

	_, err = p.sqsClient.SendMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("error enviando mensaje a SQS: %w", err)
	}

	p.logger.Debug("Evento publicado en SQS",
		zap.String("event_type", event.EventType),
		zap.String("user_id", event.UserID),
		zap.String("queue_url", p.queueURL),
	)

	return nil
}

func stringPtr(s string) *string {
	return &s
}

// Verificación en tiempo de compilación
var _ services.EventPublisher = (*SQSEventPublisher)(nil)
