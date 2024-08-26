package crontrigger

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const (
	// Schedules
	everyYear          = "0 0 0 1 1 *"
	everyMonth         = "0 0 0 1 * *"
	everyWeek          = "0 0 0 * * 0"
	everyDay           = "0 0 0 * * *"
	everyDayEasternTZ  = "TZ=America/New_York 0 0 * * * *"
	everyHour          = "0 0 * * * *"
	everyHourFrom9To10 = "0 9-10 * * *"
	everyMinute        = "0 * * * * *"
	everySecond        = "* * * * * *"
	everySecondSecond  = "*/2 * * * * *"

	// Workflow IDs
	workflowID1 = "workflow-id-1"

	// Trigger IDs
	triggerID1 = "test-id-1"
	triggerID2 = "test-id-2"
)

func registerTriggerToCronTriggerService(
	ctx context.Context,
	t *testing.T,
	cts *Service,
	schedule string,
	triggerID string,
) (
	<-chan capabilities.TriggerResponse,
	capabilities.TriggerRegistrationRequest,
	error,
) {
	config, err := values.NewMap(map[string]interface{}{
		"schedule": schedule,
	})
	require.NoError(t, err)

	requestMetadata := capabilities.RequestMetadata{
		WorkflowID: workflowID1,
	}
	request := capabilities.TriggerRegistrationRequest{
		TriggerID: workflowID1 + "|" + triggerID, // TODO: remove wid once added by workflow engine
		Metadata:  requestMetadata,
		Config:    config,
	}
	triggerEventsCh, err := cts.RegisterTrigger(ctx, request)

	return triggerEventsCh, request, err
}

func upwrapCronTriggerEvent(t *testing.T, event capabilities.TriggerEvent) Response {
	response := Response{}
	response.TriggerType = event.TriggerType
	assert.Equal(t, cronTriggerID, response.TriggerType)
	response.ID = event.ID
	err := event.Outputs.UnwrapTo(&response.Payload)
	require.NoError(t, err)
	require.NotNil(t, response.Payload.ScheduledExecutionTime)
	require.NotNil(t, response.Payload.ActualExecutionTime)
	return response
}

func makeTriggerID(number int) string {
	return "test-id-" + strconv.FormatUint(uint64(number), 10)
}

func requireNoChanMsg[T any](t *testing.T, ch <-chan T) {
	timedOut := false
	select {
	case <-ch:
	case <-time.After(100 * time.Millisecond):
		timedOut = true
	}
	require.True(t, timedOut)
}

func TestCronTrigger_SuccessWithStandardCronIntervals(t *testing.T) {
	cases := []struct {
		name     string
		schedule string
		interval [5]time.Duration
	}{
		{
			name:     "success - every second",
			schedule: everySecond,
			interval: [5]time.Duration{
				time.Second,
				time.Second,
				time.Second,
				time.Second,
				time.Second,
			},
		},
		{
			name:     "success - every minute",
			schedule: everyMinute,
			interval: [5]time.Duration{
				time.Minute,
				time.Minute,
				time.Minute,
				time.Minute,
				time.Minute,
			},
		},
		{
			name:     "success - every hour",
			schedule: everyHour,
			interval: [5]time.Duration{
				time.Hour,
				time.Hour,
				time.Hour,
				time.Hour,
				time.Hour,
			},
		},
		{
			name:     "success - every day",
			schedule: everyDay,
			interval: [5]time.Duration{
				24 * time.Hour,
				24 * time.Hour,
				24 * time.Hour,
				24 * time.Hour,
				24 * time.Hour,
			},
		},
		{
			name:     "success - every week on Sunday",
			schedule: everyWeek,
			interval: [5]time.Duration{
				(time.Hour * 24) * 7,
				(time.Hour * 24) * 7,
				(time.Hour * 24) * 7,
				(time.Hour * 24) * 7,
				(time.Hour * 24) * 7,
			},
		},
		{
			name:     "success - every month",
			schedule: everyMonth,
			interval: [5]time.Duration{
				(time.Hour * 24) * 31,
				(time.Hour * 24) * 28,
				(time.Hour * 24) * 31,
				(time.Hour * 24) * 30,
				(time.Hour * 24) * 31,
			},
		},
		{
			name:     "success - every year",
			schedule: everyYear,
			interval: [5]time.Duration{
				(time.Hour * 24) * 365,
				(time.Hour * 24) * 365,
				(time.Hour * 24) * 365,
				(time.Hour * 24) * 365,
				(time.Hour * 24) * 365,
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			fakeClock := clockwork.NewFakeClock()
			// Set time to have 0h0m0s0ms by advancing to rounded times
			fakeClock.Advance(time.Duration(1000000000 - fakeClock.Now().Nanosecond()))
			fakeClock.Advance(60*time.Second - time.Duration(fakeClock.Now().Second())*time.Second)
			fakeClock.Advance(60*time.Minute - time.Duration(fakeClock.Now().Minute())*time.Minute)
			fakeClock.Advance(24*time.Hour - time.Duration(fakeClock.Now().UTC().Hour())*time.Hour)
			// Set time to the first of January
			for fakeClock.Now().UTC().Month() != time.January {
				fakeClock.Advance(24 * time.Hour)
			}
			for fakeClock.Now().UTC().Day() != 1 {
				fakeClock.Advance(24 * time.Hour)
			}
			if tt.schedule == everyWeek {
				// If every week set to Saturday
				for fakeClock.Now().UTC().Weekday() != time.Sunday {
					fakeClock.Advance(24 * time.Hour)
				}
			}

			ts := New(logger.Nop(), fakeClock)
			ctx := tests.Context(t)

			// Start scheduling
			err := ts.Start(ctx)
			require.NoError(t, err)

			// Register trigger
			callback, registerUnregisterRequest, err := registerTriggerToCronTriggerService(
				ctx,
				t,
				ts,
				tt.schedule,
				makeTriggerID(1),
			)
			require.NoError(t, err)
			assert.Equal(t, len(ts.scheduler.Jobs()), 1)

			// Advance to 1ms before scheduled time, there should be no channel message
			fakeClock.Advance(tt.interval[0] - time.Millisecond)
			requireNoChanMsg(t, callback)
			// Pass scheduled time by 1ms
			fakeClock.Advance(2 * time.Millisecond)

			// 1st process
			msg := <-callback
			response := upwrapCronTriggerEvent(t, msg.Event)
			scheduledExecutionTime1, _ := time.Parse(time.RFC3339, response.Payload.ScheduledExecutionTime)

			fakeClock.Advance(tt.interval[1])

			// 2nd process
			msg = <-callback
			response = upwrapCronTriggerEvent(t, msg.Event)
			scheduledExecutionTime2, _ := time.Parse(time.RFC3339, response.Payload.ScheduledExecutionTime)

			fakeClock.Advance(tt.interval[2])

			// 3rd process
			msg = <-callback
			response = upwrapCronTriggerEvent(t, msg.Event)
			scheduledExecutionTime3, _ := time.Parse(time.RFC3339, response.Payload.ScheduledExecutionTime)

			// Unregister the trigger and check that events no longer go on the callback
			require.NoError(t, ts.UnregisterTrigger(ctx, registerUnregisterRequest))
			assert.Equal(t, len(ts.scheduler.Jobs()), 0)
			assert.Equal(t, ts.scheduler.JobsWaitingInQueue(), 0)

			// Skip to when the next execution would be
			fakeClock.Advance(tt.interval[3])

			// One interval after unregistering, should be no new messages
			msg = <-callback
			require.Equal(t, msg, capabilities.TriggerResponse{})

			// Skip to when the next execution would be
			fakeClock.Advance(tt.interval[4])

			// Two intervals after unregistering, should be no new messages
			msg = <-callback
			require.Equal(t, msg, capabilities.TriggerResponse{})

			// Close the service
			require.NoError(t, ts.Close())

			// Check scheduled execution times are every interval
			require.True(t, scheduledExecutionTime3.Equal(scheduledExecutionTime2.Add(tt.interval[2])))
			require.True(t, scheduledExecutionTime3.Equal(scheduledExecutionTime1.Add(tt.interval[1]+tt.interval[2])))
			require.True(t, scheduledExecutionTime2.Equal(scheduledExecutionTime1.Add(tt.interval[1])))
		})
	}
}

func TestCronTrigger_Load(t *testing.T) {
	const numTriggers = 1_000
	const numExecutions = 3
	require.Greater(t, numTriggers, 0)
	require.Greater(t, numExecutions, 0)

	fakeClock := clockwork.NewRealClock()

	ts := New(logger.Nop(), fakeClock)
	ctx := tests.Context(t)

	var callbacks [numTriggers]<-chan capabilities.TriggerResponse
	var unregisterRequests [numTriggers]capabilities.TriggerRegistrationRequest

	// Register triggers
	for triggerIdx := 0; triggerIdx < numTriggers; triggerIdx++ {
		callback, unregisterRequest, err := registerTriggerToCronTriggerService(
			ctx,
			t,
			ts,
			everySecond,
			makeTriggerID(triggerIdx+1),
		)
		require.NoError(t, err)
		callbacks[triggerIdx] = callback
		unregisterRequests[triggerIdx] = unregisterRequest
	}
	assert.Equal(t, len(ts.scheduler.Jobs()), numTriggers)

	// Start scheduling
	err := ts.Start(ctx)
	require.NoError(t, err)

	// Process "numExecutions" times
	var timestamps [numTriggers][numExecutions]time.Time
	var scheduledExecTimes [numTriggers][numExecutions]time.Time

	for execIdx := 0; execIdx < numExecutions; execIdx++ {
		for triggerIdx := 0; triggerIdx < numTriggers; triggerIdx++ {
			msg := <-callbacks[triggerIdx]
			response := upwrapCronTriggerEvent(t, msg.Event)
			scheduledExecutionTime, _ := time.Parse(time.RFC3339Nano, response.Payload.ScheduledExecutionTime)
			scheduledExecTimes[triggerIdx][execIdx] = scheduledExecutionTime
			actualExecutionTime, _ := time.Parse(time.RFC3339Nano, response.Payload.ActualExecutionTime)
			timestamps[triggerIdx][execIdx] = actualExecutionTime
		}
	}

	// Unregister the trigger and check that events no longer go on the callback
	for i := 0; i < numTriggers; i++ {
		require.NoError(t, ts.UnregisterTrigger(ctx, unregisterRequests[i]))
	}

	assert.Equal(t, len(ts.scheduler.Jobs()), 0)
	assert.Equal(t, ts.scheduler.JobsWaitingInQueue(), 0)

	// Wait a second to ensure no more events
	time.Sleep(time.Second * 1)
	for i := 0; i < numTriggers; i++ {
		msg := <-callbacks[i]
		require.Equal(t, capabilities.TriggerResponse{}, msg)
	}

	// Close the service
	require.NoError(t, ts.Close())

	var scheduledActualDelta [numTriggers * numExecutions]int64

	for execIdx := 0; execIdx < numExecutions; execIdx++ {
		for triggerIdx := 0; triggerIdx < numTriggers; triggerIdx++ {
			// Check all scheduled execution times at each process are the same across all triggers
			if triggerIdx > 0 {
				require.True(t, scheduledExecTimes[0][execIdx].Equal(scheduledExecTimes[triggerIdx][execIdx]))
			}
			// Check that executions happened every second
			if execIdx > 0 {
				require.True(t, scheduledExecTimes[triggerIdx][execIdx].Equal(scheduledExecTimes[triggerIdx][execIdx-1].Add(time.Second)))
			}
			// Check that actual execution time is after scheduled time
			require.True(t, timestamps[triggerIdx][execIdx].After(scheduledExecTimes[triggerIdx][execIdx]))
			// Check that scheduled time and actual time did not differ more than 1 second
			require.False(t, timestamps[triggerIdx][execIdx].After(scheduledExecTimes[triggerIdx][execIdx].Add(time.Second)))
			// Store time difference between scheduled and actual
			scheduledActualDelta[triggerIdx*numExecutions+execIdx] = timestamps[triggerIdx][execIdx].Sub(scheduledExecTimes[triggerIdx][execIdx]).Milliseconds()
		}
	}

	var averageDelta int64
	for _, num := range scheduledActualDelta {
		averageDelta += num
	}
	averageDelta = averageDelta / int64(len(scheduledActualDelta))
	fmt.Println("Average Delta: ", averageDelta, "ms")
}

func TestCronTrigger_RegisterTriggerBeforeStart(t *testing.T) {
	ts := New(logger.Nop(), nil)
	ctx := tests.Context(t)

	// Register trigger
	callback, registerUnregisterRequest, err := registerTriggerToCronTriggerService(
		ctx,
		t,
		ts,
		everySecond,
		makeTriggerID(1),
	)
	require.NoError(t, err)
	assert.Equal(t, len(ts.scheduler.Jobs()), 1)

	// Start scheduling
	err = ts.Start(ctx)
	require.NoError(t, err)

	// 1st process
	msg := <-callback
	response := upwrapCronTriggerEvent(t, msg.Event)
	scheduledExecutionTime1, _ := time.Parse(time.RFC3339, response.Payload.ScheduledExecutionTime)
	actualExecutionTime1, _ := time.Parse(time.RFC3339, response.Payload.ActualExecutionTime)

	// 2nd process
	msg = <-callback
	response = upwrapCronTriggerEvent(t, msg.Event)
	scheduledExecutionTime2, _ := time.Parse(time.RFC3339, response.Payload.ScheduledExecutionTime)
	actualExecutionTime2, _ := time.Parse(time.RFC3339, response.Payload.ActualExecutionTime)

	// Unregister the trigger and check that events no longer go on the callback
	require.NoError(t, ts.UnregisterTrigger(ctx, registerUnregisterRequest))
	assert.Equal(t, len(ts.scheduler.Jobs()), 0)
	assert.Equal(t, ts.scheduler.JobsWaitingInQueue(), 0)

	// Close the service
	require.NoError(t, ts.Close())

	// Check that executions happened every second
	require.True(t, scheduledExecutionTime2.Equal(scheduledExecutionTime1.Add(time.Second)))
	// Check that actual execution time is after scheduled time
	require.True(t, actualExecutionTime1.After(scheduledExecutionTime1))
	require.True(t, actualExecutionTime2.After(scheduledExecutionTime2))
	// Check that scheduled time and actual time did not differ more than 1 second
	require.False(t, actualExecutionTime1.After(scheduledExecutionTime1.Add(time.Second)))
	require.False(t, actualExecutionTime2.After(scheduledExecutionTime2.Add(time.Second)))
}

func absDiffInt(x, y int32) int32 {
	if x < y {
		return y - x
	}
	return x - y
}

func TestCronTrigger_TimeWindows(t *testing.T) {
	fakeClock := clockwork.NewFakeClock()
	// Set time to have 0ms by advancing to next truncated second
	fakeClock.Advance(fakeClock.Now().Truncate(time.Second).Add(time.Second).Sub(fakeClock.Now()))
	// Set time to 8:50am UTC
	hour, min, sec := fakeClock.Now().UTC().Clock()
	fakeClock.Advance(time.Duration(absDiffInt(int32(sec), 0)) * time.Second)
	fakeClock.Advance(time.Duration(absDiffInt(int32(min), 50)) * time.Minute)
	fakeClock.Advance(time.Duration(absDiffInt(int32(hour), 8)) * time.Hour)

	ts := New(logger.Nop(), fakeClock)
	ctx := tests.Context(t)

	// Register trigger
	callback, registerUnregisterRequest, err := registerTriggerToCronTriggerService(
		ctx,
		t,
		ts,
		everyHourFrom9To10,
		makeTriggerID(1),
	)
	require.NoError(t, err)
	assert.Equal(t, len(ts.scheduler.Jobs()), 1)

	// Start scheduling
	err = ts.Start(ctx)
	require.NoError(t, err)

	// Advance to 9am
	fakeClock.Advance(10 * time.Minute)

	// 1st process @ 9am UTC
	msg := <-callback
	response := upwrapCronTriggerEvent(t, msg.Event)
	scheduledExecutionTime1, _ := time.Parse(time.RFC3339, response.Payload.ScheduledExecutionTime)

	// Advance to 10am
	fakeClock.Advance(time.Hour)

	// 2nd process @ 10am UTC
	msg = <-callback
	response = upwrapCronTriggerEvent(t, msg.Event)
	scheduledExecutionTime2, _ := time.Parse(time.RFC3339, response.Payload.ScheduledExecutionTime)

	// Advance to 9am UTC next day
	fakeClock.Advance(time.Hour * 23)

	// should not process again until next day
	msg = <-callback
	response = upwrapCronTriggerEvent(t, msg.Event)
	scheduledExecutionTime3, _ := time.Parse(time.RFC3339, response.Payload.ScheduledExecutionTime)

	// Unregister the trigger and check that events no longer go on the callback
	require.NoError(t, ts.UnregisterTrigger(ctx, registerUnregisterRequest))
	assert.Equal(t, len(ts.scheduler.Jobs()), 0)
	assert.Equal(t, ts.scheduler.JobsWaitingInQueue(), 0)

	// Close the service
	require.NoError(t, ts.Close())

	// Check scheduled execution times are every second
	require.True(t, scheduledExecutionTime3.Equal(scheduledExecutionTime2.Add(23*time.Hour)))
	require.True(t, scheduledExecutionTime3.Equal(scheduledExecutionTime1.Add(24*time.Hour)))
	require.True(t, scheduledExecutionTime2.Equal(scheduledExecutionTime1.Add(time.Hour)))
}

func TestCronTrigger_MultipleRealClock(t *testing.T) {
	realClock := clockwork.NewRealClock()
	ts := New(logger.Nop(), realClock)
	ctx := tests.Context(t)

	callback1, registerUnregisterRequest1, err := registerTriggerToCronTriggerService(
		ctx,
		t,
		ts,
		everySecond,
		triggerID1,
	)
	require.NoError(t, err)

	callback2, registerUnregisterRequest2, err := registerTriggerToCronTriggerService(
		ctx,
		t,
		ts,
		everySecondSecond,
		triggerID2,
	)
	require.NoError(t, err)

	assert.Equal(t, len(ts.scheduler.Jobs()), 2)

	// Start scheduling
	err = ts.Start(ctx)
	require.NoError(t, err)

	// 1st second
	msg1 := <-callback1
	response1 := upwrapCronTriggerEvent(t, msg1.Event)
	scheduledExecutionTime1_1, _ := time.Parse(time.RFC3339, response1.Payload.ScheduledExecutionTime)

	// 2nd second
	msg1 = <-callback1
	response1 = upwrapCronTriggerEvent(t, msg1.Event)
	scheduledExecutionTime1_2, _ := time.Parse(time.RFC3339, response1.Payload.ScheduledExecutionTime)
	eventID1Run2 := response1.ID

	msg2 := <-callback2
	response2 := upwrapCronTriggerEvent(t, msg2.Event)
	scheduledExecutionTime2_1, _ := time.Parse(time.RFC3339, response2.Payload.ScheduledExecutionTime)
	eventID2Run2 := response2.ID

	// 3rd second
	msg1 = <-callback1
	response1 = upwrapCronTriggerEvent(t, msg1.Event)
	scheduledExecutionTime1_3, _ := time.Parse(time.RFC3339, response1.Payload.ScheduledExecutionTime)

	// 4th second
	msg1 = <-callback1
	response1 = upwrapCronTriggerEvent(t, msg1.Event)
	scheduledExecutionTime1_4, _ := time.Parse(time.RFC3339, response1.Payload.ScheduledExecutionTime)
	eventID1Run4 := response1.ID

	msg2 = <-callback2
	response2 = upwrapCronTriggerEvent(t, msg2.Event)
	scheduledExecutionTime2_2, _ := time.Parse(time.RFC3339, response2.Payload.ScheduledExecutionTime)
	eventID2Run4 := response2.ID

	// Unregister the trigger and check that events no longer go on the callback
	require.NoError(t, ts.UnregisterTrigger(ctx, registerUnregisterRequest1))
	require.NoError(t, ts.UnregisterTrigger(ctx, registerUnregisterRequest2))

	msg1 = <-callback1
	require.Equal(t, msg1, capabilities.TriggerResponse{})
	msg2 = <-callback2
	require.Equal(t, msg2, capabilities.TriggerResponse{})
	time.Sleep(time.Second)
	msg1 = <-callback1
	require.Equal(t, msg1, capabilities.TriggerResponse{})
	msg2 = <-callback2
	require.Equal(t, msg2, capabilities.TriggerResponse{})

	// Close the service
	require.NoError(t, ts.Close())

	// Check scheduled execution times
	// Trigger 1 happened every second
	require.True(t, scheduledExecutionTime1_4.Equal(scheduledExecutionTime1_3.Add(time.Second)))
	require.True(t, scheduledExecutionTime1_4.Equal(scheduledExecutionTime1_2.Add(time.Second*2)))
	require.True(t, scheduledExecutionTime1_4.Equal(scheduledExecutionTime1_1.Add(time.Second*3)))
	require.True(t, scheduledExecutionTime1_3.Equal(scheduledExecutionTime1_2.Add(time.Second)))
	require.True(t, scheduledExecutionTime1_3.Equal(scheduledExecutionTime1_1.Add(time.Second*2)))
	require.True(t, scheduledExecutionTime1_2.Equal(scheduledExecutionTime1_1.Add(time.Second)))
	// Trigger 2 happened every second second
	require.True(t, scheduledExecutionTime2_2.Equal(scheduledExecutionTime2_1.Add(time.Second*2)))
	// The 2nd and 4th second have the same event ID
	require.Equal(t, eventID1Run2, eventID2Run2)
	require.Equal(t, eventID1Run4, eventID2Run4)
}

func TestCronTrigger_TimeZone(t *testing.T) {
	fakeClock := clockwork.NewFakeClock()
	location, _ := time.LoadLocation("America/New_York")
	// Set time to have 0ms by advancing to next truncated second
	fakeClock.Advance(time.Duration(1000000000 - fakeClock.Now().Nanosecond()))
	// Set time to 23:50pm Eastern
	now := fakeClock.Now().In(location)
	hour, min, sec := now.Clock()
	fakeClock.Advance(time.Duration(absDiffInt(int32(sec), 60)) * time.Second)
	fakeClock.Advance(time.Duration(absDiffInt(int32(min), 49)) * time.Minute)
	fakeClock.Advance(time.Duration(absDiffInt(int32(hour), 23)) * time.Hour)

	ts := New(logger.Nop(), fakeClock)
	ctx := tests.Context(t)

	// Register trigger
	callback, registerUnregisterRequest, err := registerTriggerToCronTriggerService(
		ctx,
		t,
		ts,
		everyDayEasternTZ,
		makeTriggerID(1),
	)
	require.NoError(t, err)
	assert.Equal(t, len(ts.scheduler.Jobs()), 1)

	// Start scheduling
	err = ts.Start(ctx)
	require.NoError(t, err)

	// Advance to 1ms before trigger
	fakeClock.Advance(9*time.Minute + 59*time.Second + 999*time.Millisecond)

	// There should be no channel message
	requireNoChanMsg(t, callback)

	// Advance to next 12am Eastern
	fakeClock.Advance(time.Millisecond)

	// 1st process @ 12am Eastern
	msg := <-callback
	response := upwrapCronTriggerEvent(t, msg.Event)
	scheduledExecutionTime, _ := time.Parse(time.RFC3339, response.Payload.ScheduledExecutionTime)

	// Unregister the trigger and check that events no longer go on the callback
	require.NoError(t, ts.UnregisterTrigger(ctx, registerUnregisterRequest))
	assert.Equal(t, len(ts.scheduler.Jobs()), 0)
	assert.Equal(t, ts.scheduler.JobsWaitingInQueue(), 0)

	// Close the service
	require.NoError(t, ts.Close())

	// Check scheduled execution is at 12am Eastern
	timezone, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	expectedEasternExecution := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, timezone)
	require.True(t, scheduledExecutionTime.Equal(expectedEasternExecution))
}

func TestCronTrigger_RegisterTrigger(t *testing.T) {
	cases := []struct {
		name              string
		schedule          string
		shouldErr         bool
		expectedErrString string
	}{
		// No Error
		{
			name:              "valid cron schedule - 6 entries",
			schedule:          everySecond,
			shouldErr:         false,
			expectedErrString: "",
		},
		{
			name:              "valid cron schedule - 5 entries",
			schedule:          "* * * * *",
			shouldErr:         false,
			expectedErrString: "",
		},

		// Error
		{
			name:              "invalid cron schedule - empty",
			schedule:          "",
			shouldErr:         true,
			expectedErrString: "gocron: CronJob: crontab parse failure\nexpected 5 to 6 fields, found 0: []",
		},
		{
			name:              "invalid cron schedule - not a cron schedule",
			schedule:          "d d d d d",
			shouldErr:         true,
			expectedErrString: "gocron: CronJob: crontab parse failure\nfailed to parse int from d: strconv.Atoi: parsing \"d\": invalid syntax",
		},
		{
			name:              "invalid cron schedule - invalid timezone",
			schedule:          "TZ=moon * * * * *",
			shouldErr:         true,
			expectedErrString: "gocron: CronJob: crontab parse failure\nprovided bad location moon: unknown time zone moon",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ts := New(logger.Nop(), nil)
			ctx := tests.Context(t)

			_, _, err := registerTriggerToCronTriggerService(
				ctx,
				t,
				ts,
				tt.schedule,
				triggerID1,
			)

			if tt.shouldErr {
				require.Error(t, err)
				if tt.expectedErrString != "" {
					require.Equal(t, tt.expectedErrString, err.Error())
				}
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, ts.Close())
		})
	}
}

func TestCronTrigger_RegisterTriggerDuplicateError(t *testing.T) {
	ts := New(logger.Nop(), nil)
	ctx := tests.Context(t)

	config, err := values.NewMap(map[string]interface{}{
		"schedule": everySecond,
	})
	require.NoError(t, err)

	requestMetadata := capabilities.RequestMetadata{
		WorkflowID: workflowID1,
	}
	request := capabilities.TriggerRegistrationRequest{
		TriggerID: triggerID1,
		Metadata:  requestMetadata,
		Config:    config,
	}

	_, err = ts.RegisterTrigger(ctx, request)
	require.NoError(t, err)
	_, err = ts.RegisterTrigger(ctx, request)
	require.Error(t, err)
	require.Equal(t, "triggerId test-id-1 already registered", err.Error())
}

func TestCronTrigger_UnregisterTriggerError(t *testing.T) {
	ts := New(logger.Nop(), nil)
	ctx := tests.Context(t)

	config, err := values.NewMap(map[string]interface{}{
		"schedule": everySecond,
	})
	require.NoError(t, err)

	requestMetadata := capabilities.RequestMetadata{
		WorkflowID: workflowID1,
	}
	request := capabilities.TriggerRegistrationRequest{
		TriggerID: "invalid",
		Metadata:  requestMetadata,
		Config:    config,
	}

	err = ts.UnregisterTrigger(ctx, request)
	require.Error(t, err)
	require.Equal(t, "triggerId invalid not found", err.Error())
}

func TestCronTrigger_CloseStartErrors(t *testing.T) {
	ts := New(logger.Nop(), nil)
	ctx := tests.Context(t)

	err := ts.Start(ctx)
	require.NoError(t, err)
	err = ts.Close()
	require.NoError(t, err)
	err = ts.Start(ctx)
	require.Error(t, err)
}

func TestCronTrigger_GenerateSchema(t *testing.T) {
	ts := New(logger.Nop(), nil)
	schema, err := ts.Schema()
	require.NoError(t, err)
	var shouldUpdate = false
	if shouldUpdate {
		err = os.WriteFile("../testdata/fixtures/cron/schema.json", []byte(schema), 0600)
		require.NoError(t, err)
	}

	fixture, err := os.ReadFile("../testdata/fixtures/cron/schema.json")
	require.NoError(t, err)

	utils.AssertJSONEqual(t, fixture, []byte(schema))
}
