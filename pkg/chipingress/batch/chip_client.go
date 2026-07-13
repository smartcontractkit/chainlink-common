package batch

// Well-known chip_client metric label values for batch client metrics.
// Pass to WithChipClient when wiring batch.NewBatchClient.
const (
	ChipClientBeholder       = "beholder"
	ChipClientDurableEmitter = "durable_emitter"
)
