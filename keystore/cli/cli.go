package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"

	ks "github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/pgstore"
)

const (
	KeystoreLoadTimeout = 10 * time.Second
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "./keystore <command>",
		Short:        "CLI for managing keystore keys",
		SilenceUsage: true,
	}
	cmd.PersistentFlags().String("file-path", "", `
Path to keystore file (e.g. path/to/mykeystore.json).
File must already exist, can be empty json file for a new keystore.
Required if --db-url is not set. 
	`)
	cmd.PersistentFlags().String("db-url", "", ` 
Postgres connection URL (e.g. postgres://user:pass@host:5432/db?sslmode=disable).
Required if --file-path is not set.
Requires a database with the encrypted_keystore table initialized as in 
https://github.com/smartcontractkit/chainlink/blob/main/core/store/migrate/migrations/0280_create_keystore_table.sql#L1
	`)
	cmd.PersistentFlags().String("password", "", "keystore password used to encrypt the key material")

	cmd.AddCommand(NewListCmd(), NewCreateCmd(), NewDeleteCmd(), NewExportCmd(), NewImportCmd())
	return cmd
}

type Key struct {
	Name      string    `json:"name"`
	KeyType   string    `json:"key_type"`
	CreatedAt time.Time `json:"created_at"`
	PublicKey []byte    `json:"public_key"`
	Metadata  []byte    `json:"metadata"`
}

func NewListCmd() *cobra.Command {
	return &cobra.Command{
		Use: "list", Short: "List keys",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), KeystoreLoadTimeout)
			defer cancel()
			k, err := loadKeystoreFromFlags(ctx, cmd)
			if err != nil {
				return err
			}
			resp, err := k.GetKeys(ctx, ks.GetKeysRequest{})
			if err != nil {
				return err
			}
			keys := []Key{}
			for _, g := range resp.Keys {
				keys = append(keys, Key{
					Name:      g.KeyInfo.Name,
					KeyType:   string(g.KeyInfo.KeyType),
					CreatedAt: g.KeyInfo.CreatedAt,
					PublicKey: g.KeyInfo.PublicKey,
					Metadata:  g.KeyInfo.Metadata,
				})
			}
			jsonBytes, err := json.Marshal(keys)
			if err != nil {
				return err
			}
			cmd.OutOrStdout().Write(jsonBytes)
			return nil
		},
	}
}

func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "create", Short: "Create a key",
		RunE: func(cmd *cobra.Command, _ []string) error {
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			typ, err := cmd.Flags().GetString("type")
			if err != nil {
				return err
			}
			if name == "" || typ == "" {
				return errors.New("--name and --type are required")
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), KeystoreLoadTimeout)
			defer cancel()
			k, err := loadKeystoreFromFlags(ctx, cmd)
			if err != nil {
				return err
			}
			_, err = k.CreateKeys(ctx, ks.CreateKeysRequest{Keys: []ks.CreateKeyRequest{{KeyName: name, KeyType: ks.KeyType(typ)}}})
			return err
		},
	}
	cmd.Flags().String("name", "", "key name")
	cmd.Flags().String("type", "", "key type (X25519|ecdh-p256|ed25519|ecdsa-secp256k1)")
	return cmd
}

func NewDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "delete", Short: "Delete a key",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			if n == "" {
				return errors.New("--name is required")
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), KeystoreLoadTimeout)
			defer cancel()
			k, err := loadKeystoreFromFlags(ctx, cmd)
			if err != nil {
				return err
			}
			_, err = k.DeleteKeys(ctx, ks.DeleteKeysRequest{KeyNames: []string{n}})
			return err
		},
	}
	cmd.Flags().String("name", "", "key name")
	return cmd
}

func NewExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "export", Short: "Export a key to an encrypted JSON file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			keyNameToExport, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			outputFilePath, err := cmd.Flags().GetString("out")
			if err != nil {
				return err
			}
			exportPassword, err := cmd.Flags().GetString("password")
			if err != nil {
				return err
			}
			if keyNameToExport == "" || outputFilePath == "" || exportPassword == "" {
				return errors.New("--name and --out and --password are required")
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), KeystoreLoadTimeout)
			defer cancel()
			k, err := loadKeystoreFromFlags(ctx, cmd)
			if err != nil {
				return err
			}
			resp, err := k.ExportKeys(ctx, ks.ExportKeysRequest{Keys: []ks.ExportKeyParam{
				{
					KeyName: keyNameToExport,
					Enc: ks.EncryptionParams{
						Password:     exportPassword,
						ScryptParams: ks.DefaultScryptParams,
					},
				},
			}})
			if err != nil {
				return err
			}
			if len(resp.Keys) != 1 {
				return errors.New("unexpected export response")
			}
			return os.WriteFile(outputFilePath, resp.Keys[0].Data, 0o600)
		},
	}
	cmd.Flags().String("name", "", "key name to export")
	cmd.Flags().String("out", "", "output file path for encrypted key JSON")
	cmd.Flags().String("password", "", "export password")
	return cmd
}

func NewImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "import", Short: "Import an encrypted key JSON file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			inputFilePath, err := cmd.Flags().GetString("in")
			if err != nil {
				return err
			}
			if name == "" || inputFilePath == "" {
				return errors.New("--name and --in are required")
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), KeystoreLoadTimeout)
			defer cancel()
			k, err := loadKeystoreFromFlags(ctx, cmd)
			if err != nil {
				return err
			}
			encBytes, err := os.ReadFile(inputFilePath)
			if err != nil {
				return err
			}
			impPass, err := resolvePasswordFromFlags(cmd, "password")
			if err != nil {
				return err
			}
			_, err = k.ImportKeys(ctx, ks.ImportKeysRequest{Keys: []ks.ImportKeyRequest{{KeyName: name, Data: encBytes, Password: impPass}}})
			return err
		},
	}
	cmd.Flags().String("name", "", "key name to import as")
	cmd.Flags().String("in", "", "path to encrypted key JSON to import")
	cmd.Flags().String("password", "", "password for encrypted input")
	return cmd
}

func loadKeystoreFromFlags(ctx context.Context, cmd *cobra.Command) (ks.Keystore, error) {
	root := cmd.Root()
	filePath, err := root.Flags().GetString("file-path")
	if err != nil {
		return nil, err
	}
	dbURL, err := root.Flags().GetString("db-url")
	if err != nil {
		return nil, err
	}
	pass, err := resolvePasswordFromFlags(root, "password")
	if err != nil {
		return nil, err
	}
	if (filePath == "" && dbURL == "") || (filePath != "" && dbURL != "") {
		return nil, errors.New("exactly one of --file-path or --db-url must be set")
	}

	var storage ks.Storage
	if filePath != "" {
		storage = ks.NewFileStorage(filePath)
	} else {
		db, err := sqlx.ConnectContext(ctx, "postgres", dbURL)
		if err != nil {
			return nil, fmt.Errorf("connect db: %w", err)
		}
		storage = pgstore.NewStorage(db, "default")
	}
	// Can revisit whether custom scrypt params are actually needed in a CLI context
	// (I doubt it, so simpler to leave out).
	enc := ks.EncryptionParams{Password: pass, ScryptParams: ks.DefaultScryptParams}
	return ks.LoadKeystore(ctx, storage, enc)
}

func resolvePasswordFromFlags(cmd *cobra.Command, pFlag string) (string, error) {
	p, err := cmd.Flags().GetString(pFlag)
	if err != nil {
		return "", err
	}
	if p != "" {
		return p, nil
	}
	return "", fmt.Errorf("password required: set --%s", pFlag)
}
