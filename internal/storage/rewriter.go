package storage

import (
	"context"
)

// Rewriter интерфейс для переписывания ссылок в документах
type Rewriter interface {
	Rewrite(ctx context.Context, filePath string) error
}
