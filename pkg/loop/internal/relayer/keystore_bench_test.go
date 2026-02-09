package relayer_test

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func BenchmarkKeystore_Sign(b *testing.B) {
	for _, tt := range []struct {
		name string
		ks   func() core.Keystore
	}{
		{"nop", func() core.Keystore {
			return &benchKeystore{
				sign: func(ctx context.Context, _ string, data []byte) ([]byte, error) {
					return data, nil
				},
			}
		}},
		{"hex", func() core.Keystore {
			return &benchKeystore{
				sign: func(ctx context.Context, _ string, data []byte) ([]byte, error) {
					return []byte(hex.EncodeToString(data)), nil
				},
			}
		}},
		{"ed25519", func() core.Keystore {
			pk := ed25519.NewKeyFromSeed([]byte{31: 42})
			return &benchKeystore{
				sign: func(ctx context.Context, _ string, data []byte) ([]byte, error) {
					return ed25519.Sign(pk, data), nil
				},
			}
		}},
	} {
		b.Run(tt.name, func(b *testing.B) {
			ks := tt.ks()
			b.Run("in-process", func(b *testing.B) {
				ctx := tests.Context(b)
				acct := "0x1234"
				data := []byte("asdf")
				for b.Loop() {
					got, err := ks.Sign(ctx, acct, data)
					require.NoError(b, err)
					require.NotEmpty(b, got)
				}
			})
			b.Run("out-of-process", func(b *testing.B) {
				stopCh := make(chan struct{})
				defer close(stopCh)
				test.PluginTest(b, relayer.PluginKeystoreName, &relayer.GRPCPluginKeystore{
					PluginServer: ks,
					BrokerConfig: net.BrokerConfig{Logger: logger.Nop(), StopCh: stopCh},
				}, func(b *testing.B, ks core.Keystore) {
					b.ResetTimer()
					defer b.StopTimer()

					ctx := tests.Context(b)
					acct := "0x1234"
					data := []byte("asdf")
					for b.Loop() {
						got, err := ks.Sign(ctx, acct, data)
						require.NoError(b, err)
						require.NotEmpty(b, got)
					}
				})
			})
		})
	}
}

type benchKeystore struct {
	core.UnimplementedKeystore
	account func(ctx context.Context) ([]string, error)
	sign    func(ctx context.Context, account string, data []byte) ([]byte, error)
}

func (k benchKeystore) Accounts(ctx context.Context) ([]string, error) {
	return k.account(ctx)
}

func (k benchKeystore) Sign(ctx context.Context, account string, data []byte) ([]byte, error) {
	return k.sign(ctx, account, data)
}

func (k benchKeystore) Decrypt(ctx context.Context, account string, encrypted []byte) ([]byte, error) {
	return nil, nil
}
