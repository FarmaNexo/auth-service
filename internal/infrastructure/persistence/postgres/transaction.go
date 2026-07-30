// internal/infrastructure/persistence/postgres/transaction.go
package postgres

import (
	"context"

	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"gorm.io/gorm"
)

// txContextKey es la clave (privada al paquete) bajo la cual se guarda la
// conexión transaccional en el context. El tipo dedicado evita colisiones con
// otras claves de context.
type txContextKey struct{}

// withTx devuelve un context que transporta la conexión transaccional `tx`.
func withTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// txFromContext extrae la conexión transaccional del context, si existe.
func txFromContext(ctx context.Context) *gorm.DB {
	tx, _ := ctx.Value(txContextKey{}).(*gorm.DB)
	return tx
}

// dbConn resuelve la conexión a usar en una operación de repositorio: la
// transacción activa en el context si la hay, o la conexión por defecto. En
// ambos casos se asocia el context para honrar cancelaciones/timeouts.
//
// Todos los repositorios que participan en flujos transaccionales deben acceder
// a la base de datos mediante este helper (en lugar de `r.db.WithContext(ctx)`
// directo) para que sus operaciones se enrolen automáticamente en la
// transacción cuando el TransactionManager esté activo.
func dbConn(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx := txFromContext(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	return fallback.WithContext(ctx)
}

// GormTransactionManager implementa repositories.TransactionManager sobre GORM.
type GormTransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager crea un TransactionManager respaldado por GORM.
func NewTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

// Do ejecuta `fn` dentro de una transacción. Si ya existe una transacción en el
// context (llamada anidada), la reutiliza sin abrir una nueva — la semántica de
// commit/rollback queda gobernada por la transacción más externa.
func (m *GormTransactionManager) Do(ctx context.Context, fn func(txCtx context.Context) error) error {
	if txFromContext(ctx) != nil {
		return fn(ctx)
	}
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(withTx(ctx, tx))
	})
}

var _ repositories.TransactionManager = (*GormTransactionManager)(nil)
