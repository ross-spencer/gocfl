package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"net/url"
)

const (
	ContentDirectory1_1 = "content"
)

type InventoryV1_1 struct {
	*InventoryBase
}

func NewInventoryV1_1(ctx context.Context, object Object, id string, digest checksum.DigestAlgorithm, logger *logging.Logger) (*InventoryV1_1, error) {
	ivUrl, _ := url.Parse(string(InventorySpec1_1))
	ib, err := NewInventoryBase(ctx, object, id, ivUrl, digest, ContentDirectory1_1, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create InventoryBase")
	}

	i := &InventoryV1_1{InventoryBase: ib}
	return i, nil
}
