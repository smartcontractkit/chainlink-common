package contexts

import "context"

const transmissionScheduleCtxKey key = "transmissionScheduleCtx"

func WithTransmissionSchedule(ctx context.Context, schedule string) context.Context {
	return context.WithValue(ctx, transmissionScheduleCtxKey, schedule)
}

func TransmissionScheduleValue(ctx context.Context) string {
	return Value[string](ctx, transmissionScheduleCtxKey)
}
