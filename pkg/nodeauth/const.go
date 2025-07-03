package nodeauth

import "time"
	
const (
	workflowJWTExpiration                            = 5 * time.Minute
	workflowDONType                                  = "workflowDON"
	EnvironmentNameProductionMainnet EnvironmentName = "production_mainnet"
	EnvironmentNameProductionTestnet EnvironmentName = "production_testnet"
)
