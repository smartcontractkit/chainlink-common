package ocr3test

type ConsensusInput[T any] struct {
	Observations []T
}

type singleConsensusInput[T any] struct {
	Observations T
}
