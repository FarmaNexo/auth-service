// internal/domain/repositories/transaction_manager.go
package repositories

import "context"

// TransactionManager coordina la ejecución de múltiples operaciones de
// repositorio dentro de una única transacción de base de datos (Unit of Work).
//
// La transacción se propaga a los repositorios a través del context: el callback
// recibe un context "tx-aware" que cada repositorio detecta para reutilizar la
// misma conexión transaccional. Si el callback retorna error se hace rollback de
// todas las operaciones; si retorna nil se hace commit.
//
// Las llamadas externas (publicación de eventos SQS, HTTP a otros servicios)
// NO deben ejecutarse dentro del callback: una transacción nunca debe permanecer
// abierta a lo largo de una operación de red.
type TransactionManager interface {
	Do(ctx context.Context, fn func(txCtx context.Context) error) error
}
