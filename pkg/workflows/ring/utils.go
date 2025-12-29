package ring

import "slices"

func uniqueSorted(s []string) []string {
	result := slices.Clone(s)
	slices.Sort(result)
	return slices.Compact(result)
}

