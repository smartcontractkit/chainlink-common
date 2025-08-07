package ocr3

import (
	"google.golang.org/protobuf/proto"

	pbtypes "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
)

type idKey struct {
	WorkflowExecutionId      string
	WorkflowId               string
	WorkflowOwner            string
	WorkflowName             string
	WorkflowDonId            uint32
	WorkflowDonConfigVersion uint32
	ReportId                 string
	KeyId                    string
}

func getIDKey(rq *ReportRequest) idKey {
	return idKey{
		WorkflowExecutionId:      rq.WorkflowExecutionID,
		WorkflowId:               rq.WorkflowID,
		WorkflowOwner:            rq.WorkflowOwner,
		WorkflowName:             rq.WorkflowName,
		WorkflowDonId:            rq.WorkflowDonID,
		WorkflowDonConfigVersion: rq.WorkflowDonConfigVersion,
		ReportId:                 rq.ReportID,
		KeyId:                    rq.KeyID,
	}
}

// varintSize calculates the size of a varint encoding for the given value
func varintSize(x uint64) int {
	if x == 0 {
		return 1
	}
	size := 0
	for x > 0 {
		size++
		x >>= 7
	}
	return size
}

// stringFieldSize calculates the protobuf wire format size for a string field
func stringFieldSize(fieldNumber int, s string) int {
	if len(s) == 0 {
		return 0 // empty strings are omitted in proto3
	}
	tagSize := varintSize(uint64(fieldNumber<<3 | 2)) // wire type 2 for length-delimited
	lengthSize := varintSize(uint64(len(s)))
	return tagSize + lengthSize + len(s)
}

// uint32FieldSize calculates the protobuf wire format size for a uint32 field
func uint32FieldSize(fieldNumber int, value uint32) int {
	if value == 0 {
		return 0 // zero values are omitted in proto3
	}
	tagSize := varintSize(uint64(fieldNumber << 3)) // wire type 0 for varint
	valueSize := varintSize(uint64(value))
	return tagSize + valueSize
}

// calculateIdSize calculates the marshalled size of a single pbtypes.Id
func calculateIdSize(id *pbtypes.Id) int {
	size := 0

	// Field 1: workflowExecutionId (string)
	size += stringFieldSize(1, id.WorkflowExecutionId)

	// Field 2: workflowId (string)
	size += stringFieldSize(2, id.WorkflowId)

	// Field 3: workflowOwner (string)
	size += stringFieldSize(3, id.WorkflowOwner)

	// Field 4: workflowName (string)
	size += stringFieldSize(4, id.WorkflowName)

	// Field 6: reportId (string)
	size += stringFieldSize(6, id.ReportId)

	// Field 7: workflowDonId (uint32)
	size += uint32FieldSize(7, id.WorkflowDonId)

	// Field 8: workflowDonConfigVersion (uint32)
	size += uint32FieldSize(8, id.WorkflowDonConfigVersion)

	// Field 9: keyId (string)
	size += stringFieldSize(9, id.KeyId)

	return size
}

// calculateQuerySize calculates the precise marshalled size of a pbtypes.Query
func calculateQuerySize(ids []*pbtypes.Id) int {
	if len(ids) == 0 {
		return 0
	}

	totalSize := 0

	for _, id := range ids {
		idSize := calculateIdSize(id)
		if idSize > 0 {
			// Each repeated field element includes:
			// - tag for field 1 (ids field in Query message)
			// - length of the Id message
			// - the Id message content
			tagSize := varintSize(uint64(1<<3 | 2)) // field 1, wire type 2
			lengthSize := varintSize(uint64(idSize))
			totalSize += tagSize + lengthSize + idSize
		}
	}

	return totalSize
}

func CheckQuerySizeLimit(cachedSize int, newId *pbtypes.Id, sizeLimit int) (bool, int) {
	// Calculate size if we add one more id
	newIdSize := calculateIdSize(newId)
	var totalSizeWithNewId int
	if newIdSize > 0 {
		// Only add tag and length overhead if the ID has content
		totalSizeWithNewId = cachedSize + varintSize(uint64(1<<3|2)) + varintSize(uint64(newIdSize)) + newIdSize
	} else {
		// Empty IDs don't contribute to the total size
		totalSizeWithNewId = cachedSize
	}

	// Check against limits
	if totalSizeWithNewId > sizeLimit {
		// Stop adding more ids
		return false, cachedSize
	}

	return true, totalSizeWithNewId
}

// messageFieldSize calculates the protobuf wire format size for a message field
func messageFieldSize(fieldNumber int, msg proto.Message) int {
	if msg == nil {
		return 0 // nil messages are omitted in proto3
	}
	msgSize := proto.Size(msg)
	if msgSize == 0 {
		return 0 // empty messages are omitted in proto3
	}
	tagSize := varintSize(uint64(fieldNumber<<3 | 2)) // wire type 2 for length-delimited
	lengthSize := varintSize(uint64(msgSize))
	return tagSize + lengthSize + msgSize
}

// calculateObservationSize calculates the marshalled size of a single pbtypes.Observation
func calculateObservationSize(obs *pbtypes.Observation) int {
	size := 0

	// Field 1: id (Id message)
	size += messageFieldSize(1, obs.Id)

	// Field 4: observations (values.v1.List message)
	size += messageFieldSize(4, obs.Observations)

	// Field 5: overriddenEncoderName (string)
	size += stringFieldSize(5, obs.OverriddenEncoderName)

	// Field 6: overriddenEncoderConfig (values.v1.Map message)
	size += messageFieldSize(6, obs.OverriddenEncoderConfig)

	return size
}

// calculateObservationsSize calculates the precise marshalled size of a pbtypes.Observations
func calculateObservationsSize(observations []*pbtypes.Observation) int {
	if len(observations) == 0 {
		return 0
	}

	totalSize := 0

	for _, obs := range observations {
		obsSize := calculateObservationSize(obs)
		if obsSize > 0 {
			// Each repeated field element includes:
			// - tag for field 1 (observations field in Observations message)
			// - length of the Observation message
			// - the Observation message content
			tagSize := varintSize(uint64(1<<3 | 2)) // field 1, wire type 2
			lengthSize := varintSize(uint64(obsSize))
			totalSize += tagSize + lengthSize + obsSize
		}
	}

	return totalSize
}

// CheckObservationSizeLimit checks if adding a new observation would exceed the size limit
func CheckObservationSizeLimit(cachedSize int, newObs *pbtypes.Observation, sizeLimit int) (bool, int) {
	// Calculate size if we add one more observation
	newObsSize := calculateObservationSize(newObs)
	var totalSizeWithNewObs int
	if newObsSize > 0 {
		// Only add tag and length overhead if the observation has content
		totalSizeWithNewObs = cachedSize + varintSize(uint64(1<<3|2)) + varintSize(uint64(newObsSize)) + newObsSize
	} else {
		// Empty observations don't contribute to the total size
		totalSizeWithNewObs = cachedSize
	}

	// Check against limits
	if totalSizeWithNewObs > sizeLimit {
		// Stop adding more observations
		return false, cachedSize
	}

	return true, totalSizeWithNewObs
}

// repeatedStringFieldSize calculates the protobuf wire format size for repeated string fields
func repeatedStringFieldSize(fieldNumber int, strings []string) int {
	totalSize := 0
	for _, s := range strings {
		if len(s) > 0 {
			// Each string in repeated field has its own tag and length
			tagSize := varintSize(uint64(fieldNumber<<3 | 2)) // wire type 2 for length-delimited
			lengthSize := varintSize(uint64(len(s)))
			totalSize += tagSize + lengthSize + len(s)
		}
	}
	return totalSize
}

// calculateObservationsMessageSize calculates the marshalled size of a pbtypes.Observations message
func calculateObservationsMessageSize(observations *pbtypes.Observations) int {
	if observations == nil {
		return 0
	}

	size := 0

	// Field 1: observations (repeated Observation)
	for _, obs := range observations.Observations {
		obsSize := calculateObservationSize(obs)
		if obsSize > 0 {
			tagSize := varintSize(uint64(1<<3 | 2)) // field 1, wire type 2
			lengthSize := varintSize(uint64(obsSize))
			size += tagSize + lengthSize + obsSize
		}
	}

	// Field 2: registeredWorkflowIds (repeated string)
	size += repeatedStringFieldSize(2, observations.RegisteredWorkflowIds)

	// Field 3: timestamp (google.protobuf.Timestamp message)
	size += messageFieldSize(3, observations.Timestamp)

	return size
}

// CheckObservationsSizeLimit checks if adding a new observation to a pbtypes.Observations would exceed the size limit
func CheckObservationsSizeLimit(cachedSize int, newObs *pbtypes.Observation, sizeLimit int) (bool, int) {
	// Calculate size if we add one more observation to the observations field
	newObsSize := calculateObservationSize(newObs)
	var totalSizeWithNewObs int
	if newObsSize > 0 {
		// Only add tag and length overhead if the observation has content
		totalSizeWithNewObs = cachedSize + varintSize(uint64(1<<3|2)) + varintSize(uint64(newObsSize)) + newObsSize
	} else {
		// Empty observations don't contribute to the total size
		totalSizeWithNewObs = cachedSize
	}

	// Check against limits
	if totalSizeWithNewObs > sizeLimit {
		// Stop adding more observations
		return false, cachedSize
	}

	return true, totalSizeWithNewObs
}

// calculateReportSize calculates the marshalled size of a single pbtypes.Report
func calculateReportSize(report *pbtypes.Report) int {
	if report == nil {
		return 0
	}

	size := 0

	// Field 1: id (Id message)
	size += messageFieldSize(1, report.Id)

	// Field 2: outcome (AggregationOutcome message)
	size += messageFieldSize(2, report.Outcome)

	return size
}

// calculateReportsSize calculates the precise marshalled size of current_reports from pbtypes.Outcome
func calculateReportsSize(reports []*pbtypes.Report) int {
	if len(reports) == 0 {
		return 0
	}

	totalSize := 0

	for _, report := range reports {
		reportSize := calculateReportSize(report)
		if reportSize > 0 {
			// Each repeated field element includes:
			// - tag for field 2 (current_reports field in Outcome message)
			// - length of the Report message
			// - the Report message content
			tagSize := varintSize(uint64(2<<3 | 2)) // field 2, wire type 2
			lengthSize := varintSize(uint64(reportSize))
			totalSize += tagSize + lengthSize + reportSize
		}
	}

	return totalSize
}

// CheckReportSizeLimit checks if adding a new report to the outcome would exceed size limits
func CheckReportSizeLimit(cachedSize int, newReport *pbtypes.Report, sizeLimit int) (bool, int) {
	if newReport == nil {
		return true, cachedSize
	}

	// Calculate size if we add one more report
	newReportSize := calculateReportSize(newReport)
	var totalSizeWithNewReport int
	if newReportSize > 0 {
		// Only add tag and length overhead if the report has content
		totalSizeWithNewReport = cachedSize + varintSize(uint64(2<<3|2)) + varintSize(uint64(newReportSize)) + newReportSize
	} else {
		// Empty reports don't contribute to the total size
		totalSizeWithNewReport = cachedSize
	}

	// Check against limits
	if totalSizeWithNewReport > sizeLimit {
		// Stop adding more reports
		return false, cachedSize
	}

	return true, totalSizeWithNewReport
}
