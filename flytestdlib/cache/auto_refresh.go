package cache

import (
	"context"

	"github.com/flyteorg/flyte/v2/flytestdlib/errors"
)

type ItemID = string
type Batch = []ItemWrapper

const (
	ErrNotFound errors.ErrorCode = "NOT_FOUND"
)


// AutoRefresh with regular GetOrCreate and Delete along with background asynchronous refresh. Caller provides
// callbacks for create, refresh and delete item.
// The cache doesn't provide apis to update items.
type AutoRefresh interface {
	// Starts background refresh of items. To shutdown the cache, cancel the context.
	Start(ctx context.Context) error

	// Get item by id.
	Get(id ItemID) (Item, error)

	// Get object if exists else create it.
	GetOrCreate(id ItemID, item Item) (Item, error)

	// DeleteDelayed queues an item for deletion. It Will get deleted as part of the next Sync cycle. Until the next sync
	// cycle runs, Get and GetOrCreate will continue to return the Item in its previous state.
	DeleteDelayed(id ItemID) error
}

type Item interface {
	IsTerminal() bool
}

// Items are wrapped inside an ItemWrapper to be stored in the cache.
type ItemWrapper interface {
	GetID() ItemID
	GetItem() Item
}

// Represents the response for the sync func
type ItemSyncResponse struct {
	ID     ItemID
	Item   Item
	Action SyncAction
}

// Possible actions for the cache to take as a result of running the sync function on any given cache item
type SyncAction int

const (
	Unchanged SyncAction = iota

	// The item returned has been updated and should be updated in the cache
	Update
)

// SyncFunc func type. Your implementation of this function for your cache instance is responsible for returning
// The new Item and what action should be taken.  The sync function has no insight into your object, and needs to be
// told explicitly if the new item is different from the old one.
type SyncFunc func(ctx context.Context, batch Batch) (
	updatedBatch []ItemSyncResponse, err error)

// CreateBatchesFunc is a func type. Your implementation of this function for your cache instance is responsible for
// subdividing the list of cache items into batches.
type CreateBatchesFunc func(ctx context.Context, snapshot []ItemWrapper) (batches []Batch, err error)
