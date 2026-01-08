package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
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
		Use: "keys",
		Long: ` 
CLI for managing keystore keys. Must specify KEYSTORE_FILE_PATH or KEYSTORE_DB_URL 
and KEYSTORE_PASSWORD in order to load the keystore. 

KEYSTORE_FILE_PATH: is the path to the keystore file, can be empty for a new keystore.
File must already exist. Example to create a new keystore file: touch ./keystore.json

KEYSTORE_DB_URL: is the postgres connection URL. Only use this if your keystore is stored in a pg database.
Requires a pg database with a 'encrypted_keystore' table with the following schema:
CREATE TABLE IF NOT EXISTS encrypted_keystore (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    encrypted_data BYTEA NOT NULL
);

KEYSTORE_PASSWORD is the password used to encrypt the key material before storage.
`,
		Short:        "CLI for managing keystore keys",
		SilenceUsage: true,
	}
	cmd.PersistentFlags().String("file-path", "", "Overrides KEYSTORE_FILE_PATH environment variable")
	cmd.PersistentFlags().String("db-url", "", "Overrides KEYSTORE_DB_URL environment variable")
	cmd.PersistentFlags().String("password", "", "Overrides KEYSTORE_PASSWORD environment variable. Not recommended as will leave shell traces.")

	cmd.AddCommand(NewListCmd(), NewGetCmd(), NewCreateCmd(), NewDeleteCmd(), NewExportCmd(), NewImportCmd(), NewSetMetadataCmd(), NewSignCmd(), NewVerifyCmd(), NewEncryptCmd(), NewDecryptCmd())
	return cmd
}

func NewListCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "list", Short: "List keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), KeystoreLoadTimeout)
			defer cancel()
			k, err := loadKeystore(ctx, cmd)
			if err != nil {
				return err
			}
			resp, err := k.GetKeys(ctx, ks.GetKeysRequest{})
			if err != nil {
				return err
			}
			jsonBytes, err := json.Marshal(resp)
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(jsonBytes)
			return err
		},
	}
	return &cmd
}

func NewGetCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "get", Short: "Get keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKeystoreCommand[ks.GetKeysRequest, ks.GetKeysResponse](cmd, args, func(ctx context.Context, k ks.Keystore, req ks.GetKeysRequest) (ks.GetKeysResponse, error) {
				return k.GetKeys(ctx, req)
			})
		},
	}
	cmd.Flags().StringP("file", "f", "", "input file path (use \"-\" for stdin)")
	cmd.Flags().StringP("data", "d", "", "inline JSON request, e.g. '{\"Keys\": [{\"KeyName\": \"key1\", \"KeyType\": \"X25519\"}]}'")
	return &cmd
}

func NewCreateCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "create", Short: "Create a key",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKeystoreCommand[ks.CreateKeysRequest, ks.CreateKeysResponse](cmd, args, func(ctx context.Context, k ks.Keystore, req ks.CreateKeysRequest) (ks.CreateKeysResponse, error) {
				return k.CreateKeys(ctx, req)
			})
		},
	}
	cmd.Flags().StringP("file", "f", "", "input file path (use \"-\" for stdin)")
	cmd.Flags().StringP("data", "d", "", "inline JSON request e.g. '{\"Keys\": [{\"KeyName\": \"key1\", \"KeyType\": \"X25519\"}]}'")
	return &cmd
}

func NewDeleteCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "delete", Short: "Delete a key",
		RunE: func(cmd *cobra.Command, _ []string) error {
			jsonBytesIn, err := readJSONInput(cmd)
			if err != nil {
				return err
			}
			var req ks.DeleteKeysRequest
			err = json.Unmarshal(jsonBytesIn, &req)
			if err != nil {
				return err
			}
			confirmYes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}
			if !confirmYes {
				// Prompt for confirmation on stdin
				_, err = cmd.OutOrStderr().Write([]byte(fmt.Sprintf("This will permanently delete keys: %v. Type 'yes' to confirm: ", strings.Join(req.KeyNames, ", "))))
				if err != nil {
					return err
				}
				reader := bufio.NewReader(cmd.InOrStdin())
				line, _ := reader.ReadString('\n')
				if strings.TrimSpace(line) != "yes" {
					return errors.New("delete aborted by user")
				}
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), KeystoreLoadTimeout)
			defer cancel()
			k, err := loadKeystore(ctx, cmd)
			if err != nil {
				return err
			}
			// Response is empty, so no need to marshal.
			_, err = k.DeleteKeys(ctx, req)
			return err
		},
	}
	cmd.Flags().StringP("file", "f", "", "input file path (use \"-\" for stdin)")
	cmd.Flags().StringP("data", "d", "", "inline JSON request, e.g. '{\"KeyNames\": [\"key1\", \"key2\"]}'")
	cmd.Flags().Bool("yes", false, "skip confirmation prompt")
	return &cmd
}

func NewExportCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "export", Short: "Export a key to an encrypted JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKeystoreCommand[ks.ExportKeysRequest, ks.ExportKeysResponse](cmd, args, func(ctx context.Context, k ks.Keystore, req ks.ExportKeysRequest) (ks.ExportKeysResponse, error) {
				return k.ExportKeys(ctx, req)
			})
		},
	}
	cmd.Flags().StringP("file", "f", "", "input file path (use \"-\" for stdin)")
	cmd.Flags().StringP("data", "d", "", "inline JSON request, e.g. '{\"Keys\": [{\"KeyName\": \"key1\", \"Enc\": {\"Password\": \"pass\", \"ScryptParams\": {\"N\": 1024, \"P\": 1, \"R\": 8}}}]}'")
	return &cmd
}

func NewImportCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "import", Short: "Import an encrypted key JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKeystoreCommand[ks.ImportKeysRequest, ks.ImportKeysResponse](cmd, args, func(ctx context.Context, k ks.Keystore, req ks.ImportKeysRequest) (ks.ImportKeysResponse, error) {
				return k.ImportKeys(ctx, req)
			})
		},
	}
	cmd.Flags().StringP("file", "f", "", "input file path (use \"-\" for stdin)")
	cmd.Flags().StringP("data", "d", "", "inline JSON request, e.g. '{\"Keys\": [{\"KeyName\": \"key1\", \"Data\": \"encBytes\", \"Password\": \"pass\"}]}'")
	return &cmd
}

func NewSetMetadataCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "set-metadata", Short: "Set metadata for keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKeystoreCommand[ks.SetMetadataRequest, ks.SetMetadataResponse](cmd, args, func(ctx context.Context, k ks.Keystore, req ks.SetMetadataRequest) (ks.SetMetadataResponse, error) {
				return k.SetMetadata(ctx, req)
			})
		},
	}
	cmd.Flags().StringP("file", "f", "", "input file path (use \"-\" for stdin)")
	cmd.Flags().StringP("data", "d", "", "inline JSON request, e.g. '{\"Updates\": [{\"KeyName\": \"key1\", \"Metadata\": \"base64-encoded-metadata\"}]}'")
	return &cmd
}

func runKeystoreCommand[Req any, Resp any](cmd *cobra.Command, args []string, fn func(ctx context.Context, k ks.Keystore,
	req Req) (Resp, error)) error {
	jsonBytes, err := readJSONInput(cmd)
	if err != nil {
		return err
	}
	var req Req
	err = json.Unmarshal(jsonBytes, &req)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), KeystoreLoadTimeout)
	defer cancel()
	k, err := loadKeystore(ctx, cmd)
	if err != nil {
		return err
	}
	resp, err := fn(ctx, k, req)
	if err != nil {
		return err
	}
	jsonBytesOut, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = cmd.OutOrStdout().Write(jsonBytesOut)
	if err != nil {
		return err
	}
	return nil
}

func NewSignCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "sign", Short: "Sign data with a key",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKeystoreCommand[ks.SignRequest, ks.SignResponse](cmd, args, func(ctx context.Context, k ks.Keystore, req ks.SignRequest) (ks.SignResponse, error) {
				return k.Sign(ctx, req)
			})
		},
	}
	cmd.Flags().StringP("file", "f", "", "input file path (use \"-\" for stdin)")
	cmd.Flags().StringP("data", "d", "", `
	Inline JSON request. Data is base64-encoded. 
	Example:
	echo -n 'hello' | base64 
	aGVsbG8=

	./keystore list | jq
	{
	  "Keys": [
	    {
	      "KeyName": "mykey",
	      "KeyType": "Ed25519",
	      "CreatedAt": "2025-01-01T00:00:00Z",
	      "PublicKey": "GJnS+erQbyuEm1byCjXy+6JqyX5hrGLE8oUuHSb9DFc="
	    }
	  ]
	}
	./keystore sign -d '{"KeyName": "mykey", "Data": "aGVsbG8="}' | jq
	{
	  "Signature": "OVPaQIwQAZycQtiGjhwxZ3KmAdXOHczwi3LpwQTCbtMHfy5mmrp0KusICSO0lzCMeQvxJcd5y6f3siQsohQeCg=="
	}
	./keystore verify -d '{"KeyType": "Ed25519", "PublicKey": "GJnS+erQbyuEm1byCjXy+6JqyX5hrGLE8oUuHSb9DFc=", "Data": "aGVsbG8=", "Signature": "OVPaQIwQAZycQtiGjhwxZ3KmAdXOHczwi3LpwQTCbtMHfy5mmrp0KusICSO0lzCMeQvxJcd5y6f3siQsohQeCg=="}' | jq
	{
	  "Valid": true
	}
	`)
	return &cmd
}

func NewVerifyCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "verify", Short: "Verify a signature",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKeystoreCommand[ks.VerifyRequest, ks.VerifyResponse](cmd, args, func(ctx context.Context, k ks.Keystore, req ks.VerifyRequest) (ks.VerifyResponse, error) {
				return k.Verify(ctx, req)
			})
		},
	}
	cmd.Flags().StringP("file", "f", "", "input file path (use \"-\" for stdin)")
	cmd.Flags().StringP("data", "d", "", `inline JSON request. All byte fields are base64-encoded. Example: '{"KeyType": "Ed25519", "PublicKey": "<base64>", "Data": "aGVsbG8=", "Signature": "<base64>"}'`)
	return &cmd
}

func NewEncryptCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "encrypt", Short: "Encrypt data to a remote public key",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKeystoreCommand[ks.EncryptRequest, ks.EncryptResponse](cmd, args, func(ctx context.Context, k ks.Keystore, req ks.EncryptRequest) (ks.EncryptResponse, error) {
				return k.Encrypt(ctx, req)
			})
		},
	}
	cmd.Flags().StringP("file", "f", "", "input file path (use \"-\" for stdin)")
	cmd.Flags().StringP("data", "d", "", ` 
	Inline JSON request. Data/RemotePubKey are base64-encoded.
	Example:
	echo -n 'hello' | base64 
	aGVsbG8=

	./keystore list | jq
	{
	  "Keys": [
	    {
	      "KeyName": "x25519key",
	      "KeyType": "X25519",
	      "CreatedAt": "2025-01-01T00:00:00Z",
	      "PublicKey": "GJnS+erQbyuEm1byCjXy+6JqyX5hrGLE8oUuHSb9DFc="
	    }
	  ]
	}
	./keystore encrypt -d '{"RemoteKeyType": "X25519", "RemotePubKey": "GJnS+erQbyuEm1byCjXy+6JqyX5hrGLE8oUuHSb9DFc=", "Data": "aGVsbG8="}' | jq
	{
	  "EncryptedData": "ZGVjb3JhdGVkRGF0YQ=="
	}
	./keystore decrypt -d '{"KeyName": "x25519key", "EncryptedData": "ZGVjb3JhdGVkRGF0YQ=="}' | jq
	{
	  "Data": "aGVsbG8="
	}
	`)
	return &cmd
}

func NewDecryptCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "decrypt", Short: "Decrypt data with a key",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKeystoreCommand[ks.DecryptRequest, ks.DecryptResponse](cmd, args, func(ctx context.Context, k ks.Keystore, req ks.DecryptRequest) (ks.DecryptResponse, error) {
				return k.Decrypt(ctx, req)
			})
		},
	}
	cmd.Flags().StringP("file", "f", "", "input file path (use \"-\" for stdin)")
	cmd.Flags().StringP("data", "d", "", `inline JSON request. EncryptedData is base64-encoded. Example: '{"KeyName": "mykey", "EncryptedData": "<base64>"}'`)
	return &cmd
}

func loadKeystore(ctx context.Context, cmd *cobra.Command) (ks.Keystore, error) {
	// Use parent command which has the persistent flags.
	// This works whether keystore CLI is standalone or embedded as a subcommand.
	parent := cmd.Parent()
	filePath, err := parent.Flags().GetString("file-path")
	if err != nil {
		return nil, err
	}
	dbURL, err := parent.Flags().GetString("db-url")
	if err != nil {
		return nil, err
	}
	password, err := parent.Flags().GetString("password")
	if err != nil {
		return nil, err
	}
	// Env var fallbacks (flags override)
	if filePath == "" {
		if v := os.Getenv("KEYSTORE_FILE_PATH"); v != "" {
			filePath = v
		}
	}
	if dbURL == "" {
		if v := os.Getenv("KEYSTORE_DB_URL"); v != "" {
			dbURL = v
		}
	}
	if password == "" {
		if v := os.Getenv("KEYSTORE_PASSWORD"); v != "" {
			password = v
		}
	}
	if password == "" {
		return nil, errors.New("keystore password is required")
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
	return ks.LoadKeystore(ctx, storage, password)
}

// readJSONInput reads JSON from either -f/--file (file path, "-" for stdin) or -d/--data (inline JSON).
// Exactly one must be provided.
func readJSONInput(cmd *cobra.Command) ([]byte, error) {
	filePath, err := cmd.Flags().GetString("file")
	if err != nil {
		return nil, err
	}
	data, err := cmd.Flags().GetString("data")
	if err != nil {
		return nil, err
	}

	// Must provide exactly one
	if (filePath == "" && data == "") || (filePath != "" && data != "") {
		return nil, errors.New("exactly one of -f/--file or -d/--data must be provided")
	}

	if data != "" {
		// Inline JSON
		return []byte(data), nil
	}

	// Read from file (or stdin if "-")
	if filePath == "-" {
		return io.ReadAll(cmd.InOrStdin())
	}
	return os.ReadFile(filePath)
}
