package solana

import "time"

const EventSignatureLength = 8

type EventSignature [EventSignatureLength]byte

type EventIdl EventIDLTypes

type EventIDLTypes struct {
	Event IdlEvent
	Types IdlTypeDefSlice
}

type IdlEvent struct {
	Name   string          `json:"name"`
	Fields []IdlEventField `json:"fields"`
}

type IdlEventField struct {
	Name  string  `json:"name"`
	Type  IdlType `json:"type"`
	Index bool    `json:"index"`
}

type IdlTypeDefSlice []IdlTypeDef

type IdlTypeDef struct {
	Name string       `json:"name"`
	Type IdlTypeDefTy `json:"type"`
}

type IdlTypeDefTy struct {
	Kind IdlTypeDefTyKind `json:"kind"`

	Fields   *IdlTypeDefStruct   `json:"fields,omitempty"`
	Variants IdlEnumVariantSlice `json:"variants,omitempty"`
	Codec    string              `json:"codec,omitempty"`
}

type IdlField struct {
	Name string   `json:"name"`
	Docs []string `json:"docs"` // @custom
	Type IdlType  `json:"type"`
}

type IdlEnumVariantSlice []IdlEnumVariant

type IdlEnumVariant struct {
	Name   string         `json:"name"`
	Docs   []string       `json:"docs"` // @custom
	Fields *IdlEnumFields `json:"fields,omitempty"`
}

type IdlEnumFields struct {
	IdlEnumFieldsNamed *IdlEnumFieldsNamed
	IdlEnumFieldsTuple *IdlEnumFieldsTuple
}

type IdlEnumFieldsNamed []IdlField

type IdlEnumFieldsTuple []IdlType

type IdlTypeDefStruct = []IdlField

// Wrapper type:
type IdlType struct {
	AsString         IdlTypeAsString
	AsIdlTypeVec     *IdlTypeVec
	asIdlTypeOption  *IdlTypeOption
	AsIdlTypeDefined *IdlTypeDefined
	AsIdlTypeArray   *IdlTypeArray
}
type IdlTypeAsString string

const (
	IdlTypeBool      IdlTypeAsString = "bool"
	IdlTypeU8        IdlTypeAsString = "u8"
	IdlTypeI8        IdlTypeAsString = "i8"
	IdlTypeU16       IdlTypeAsString = "u16"
	IdlTypeI16       IdlTypeAsString = "i16"
	IdlTypeU32       IdlTypeAsString = "u32"
	IdlTypeI32       IdlTypeAsString = "i32"
	IdlTypeU64       IdlTypeAsString = "u64"
	IdlTypeI64       IdlTypeAsString = "i64"
	IdlTypeU128      IdlTypeAsString = "u128"
	IdlTypeI128      IdlTypeAsString = "i128"
	IdlTypeBytes     IdlTypeAsString = "bytes"
	IdlTypeString    IdlTypeAsString = "string"
	IdlTypePublicKey IdlTypeAsString = "publicKey"

	// Custom additions:
	IdlTypeUnixTimestamp IdlTypeAsString = "unixTimestamp"
	IdlTypeHash          IdlTypeAsString = "hash"
	IdlTypeDuration      IdlTypeAsString = "duration"
)

type IdlTypeVec struct {
	Vec IdlType `json:"vec"`
}

type IdlTypeOption struct {
	Option IdlType `json:"option"`
}

// User defined type.
type IdlTypeDefined struct {
	Defined string `json:"defined"`
}

// Wrapper type:
type IdlTypeArray struct {
	Thing IdlType
	Num   int
}

type IdlTypeDefTyKind string

const (
	IdlTypeDefTyKindStruct IdlTypeDefTyKind = "struct"
	IdlTypeDefTyKindEnum   IdlTypeDefTyKind = "enum"
	IdlTypeDefTyKindCustom IdlTypeDefTyKind = "custom"
)

type SubKeyPaths [][]string

// matches cache-filter
// this filter defines what logs should be cached
// cached logs can be retrieved with [types.SolanaService.QueryLogsFromCache]
type LPFilterQuery struct {
	Name            string
	Address         PublicKey
	EventName       string
	EventSig        EventSignature
	StartingBlock   int64
	EventIdl        EventIdl
	SubkeyPaths     SubKeyPaths
	Retention       time.Duration
	MaxLogsKept     int64
	IncludeReverted bool
}

// matches lp-parsed solana logs
type Log struct {
	ChainID        string
	LogIndex       int64
	BlockHash      Hash
	BlockNumber    int64
	BlockTimestamp time.Time
	Address        PublicKey
	EventSig       EventSignature
	TxHash         Signature
	Data           []byte
	SequenceNum    int64
	Error          *string
}
