package encryption

import (
	"context"
	"errors"

	"github.com/mycontroller-org/server/v2/pkg/types"
	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	"go.uber.org/zap"
)

const (
	contextKey types.ContextKey = "encryption_api"
)

type Encryption struct {
	logger           *zap.Logger
	secret           string
	specialKeys      []string
	encryptionPrefix string
}

func New(logger *zap.Logger, secret string, specialKeys []string, encryptionPrefix string) *Encryption {
	if len(specialKeys) == 0 {
		specialKeys = cloneUtil.DefaultSpecialKeys
	}
	return &Encryption{
		logger:           logger,
		secret:           secret,
		specialKeys:      specialKeys,
		encryptionPrefix: encryptionPrefix,
	}
}

func FromContext(ctx context.Context) (*Encryption, error) {
	enc, ok := ctx.Value(contextKey).(*Encryption)
	if !ok {
		return nil, errors.New("invalid encryption instance received in context")
	}
	if enc == nil {
		return nil, errors.New("encryption instance not provided in context")
	}
	return enc, nil
}

func WithContext(ctx context.Context, enc *Encryption) context.Context {
	return context.WithValue(ctx, contextKey, enc)
}

func (e *Encryption) EncryptSecrets(source interface{}) error {
	return cloneUtil.UpdateSecrets(source, e.secret, e.encryptionPrefix, true, e.specialKeys)
}

func (e *Encryption) DecryptSecrets(source interface{}) error {
	return cloneUtil.UpdateSecrets(source, e.secret, e.encryptionPrefix, false, e.specialKeys)
}
