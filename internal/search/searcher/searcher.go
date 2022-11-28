package searcher

import (
	"context"

	"github.com/alist-org/alist/v3/internal/model"
)

type Config struct {
	Name       string
	AutoUpdate bool
}

type Searcher interface {
	// Config of the searcher
	Config() Config
	// Search specific keywords in specific path
	Search(ctx context.Context, req model.SearchReq) ([]model.SearchNode, int64, error)
	// Index obj with parent
	Index(ctx context.Context, parent string, obj model.Obj) error
	// Get by parent
	Get(ctx context.Context, parent string) ([]model.SearchNode, error)
	// Del with prefix
	Del(ctx context.Context, prefix string) error
	// Release resource
	Release(ctx context.Context) error
	// Clear all index
	Clear(ctx context.Context) error
}
