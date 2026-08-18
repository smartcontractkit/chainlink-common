// gen_owners derives workflow owner addresses from org IDs.
// Usage: go run ./pkg/workflows/cmd/gen_owners org_abc org_def org_xyz
package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

func main() {
	orgIDs := os.Args[1:]
	if len(orgIDs) == 0 {
		fmt.Fprintln(os.Stderr, "usage: gen_owners <orgID>...")
		os.Exit(1)
	}

	fmt.Printf("%-40s  %s\n", "orgID", "workflowOwner")
	fmt.Printf("%-40s  %s\n", "-----", "-------------")
	for _, orgID := range orgIDs {
		addr, err := workflows.GenerateWorkflowOwnerAddress("1", orgID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error for %s: %v\n", orgID, err)
			continue
		}
		fmt.Printf("%-40s  0x%s\n", orgID, hex.EncodeToString(addr))
	}
}
