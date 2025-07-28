package v2

import (
	"log/slog"
	"slices"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func appendAttrsToGroup(groups []string, actualAttrs []slog.Attr, newAttrs ...slog.Attr) []slog.Attr {
	if len(groups) == 0 {
		actualAttrsCopy := make([]slog.Attr, 0, len(actualAttrs)+len(newAttrs))
		actualAttrsCopy = append(actualAttrsCopy, actualAttrs...)
		actualAttrsCopy = append(actualAttrsCopy, newAttrs...)

		return uniqAttrs(actualAttrsCopy)
	}

	actualAttrs = slices.Clone(actualAttrs)

	for i := range actualAttrs {
		attr := actualAttrs[i]
		if attr.Key == groups[0] && attr.Value.Kind() == slog.KindGroup {
			actualAttrs[i] = slog.Group(groups[0], toAnySlice(appendAttrsToGroup(groups[1:], attr.Value.Group(), newAttrs...))...)

			return actualAttrs
		}
	}

	return uniqAttrs(
		append(
			actualAttrs,
			slog.Group(
				groups[0],
				toAnySlice(appendAttrsToGroup(groups[1:], []slog.Attr{}, newAttrs...))...,
			),
		),
	)
}

func toAnySlice(values []slog.Attr) []any {
	newSlice := make([]any, 0, len(values))
	for _, v := range values {
		newSlice = append(newSlice, v)
	}

	return newSlice
}

// @TODO: should be recursive
func uniqAttrs(attrs []slog.Attr) []slog.Attr {
	return uniqByLast(attrs, func(item slog.Attr) string {
		return item.Key
	})
}

func uniqByLast[T any, U comparable](collection []T, iteratee func(item T) U) []T {
	result := make([]T, 0, len(collection))
	seen := make(map[U]int, len(collection))
	seenIndex := 0

	for _, item := range collection {
		key := iteratee(item)

		if index, ok := seen[key]; ok {
			result[index] = item
			continue
		}

		seen[key] = seenIndex
		seenIndex++
		result = append(result, item)
	}

	return result
}

func appendRecordAttrsToAttrs(attrs []slog.Attr, groups []string, record *slog.Record) []slog.Attr {
	output := make([]slog.Attr, 0, len(attrs)+record.NumAttrs())
	output = append(output, attrs...)

	record.Attrs(func(attr slog.Attr) bool {
		for i := len(groups) - 1; i >= 0; i-- {
			attr = slog.Group(groups[i], attr)
		}

		output = append(output, attr)

		return true
	})

	return output
}

func attrsToMap(attrs ...slog.Attr) map[string]any {
	output := map[string]any{}

	attrsByKey := groupValuesByKey(attrs)
	for k, values := range attrsByKey {
		v := mergeAttrValues(values...)
		if v.Kind() == slog.KindGroup {
			output[k] = attrsToMap(v.Group()...)
		} else {
			output[k] = v.Any()
		}
	}

	return output
}

func groupValuesByKey(attrs []slog.Attr) map[string][]slog.Value {
	result := map[string][]slog.Value{}

	for _, item := range attrs {
		key := item.Key
		result[key] = append(result[key], item.Value)
	}

	return result
}

func mergeAttrValues(values ...slog.Value) slog.Value {
	v := values[0]

	for i := 1; i < len(values); i++ {
		if v.Kind() != slog.KindGroup || values[i].Kind() != slog.KindGroup {
			v = values[i]

			continue
		}

		v = slog.GroupValue(append(v.Group(), values[i].Group()...)...)
	}

	return v
}

func convert(loggerAttr []slog.Attr, groups []string, record *slog.Record) []zap.Field {
	// aggregate all attributes
	attrs := appendRecordAttrsToAttrs(loggerAttr, groups, record)

	// handler formatter
	fields := attrsToMap(attrs...)

	output := []zapcore.Field{}
	for k, v := range fields {
		output = append(output, zap.Any(k, v))
	}

	return output
}
