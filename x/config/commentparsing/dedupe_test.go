package commentparsing

import (
	"go/parser"
	"go/token"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The types below stand in for a dependency's, with the DocComments methods a generator would have
// written, so a run against any directory outside this module resolves them the compiled way. That
// keeps the arrangements under test to one file, with no package to parse and no file to write.
type dedupeShared struct {
	Value string
}

func (dedupeShared) DocComments() (string, map[string]FieldDoc) {
	return "dedupeShared", map[string]FieldDoc{"Value": {Comment: "Value is documented."}}
}

type dedupeEmbedded struct {
	Flat string
}

func (dedupeEmbedded) DocComments() (string, map[string]FieldDoc) {
	return "dedupeEmbedded", map[string]FieldDoc{"Flat": {Comment: "Flat is documented."}}
}

// dedupeAlpha reaches dedupeShared through every container a config can use, and refers back to
// itself so a walk without memory would not terminate.
type dedupeAlpha struct {
	dedupeEmbedded
	Direct   dedupeShared
	Pointer  *dedupeShared
	Slice    []dedupeShared
	Array    [2]dedupeShared
	Keyed    map[string]dedupeShared
	Pointers []*dedupeShared
	Self     *dedupeAlpha
}

func (dedupeAlpha) DocComments() (string, map[string]FieldDoc) {
	return "dedupeAlpha", map[string]FieldDoc{"Direct": {Comment: "Direct is documented."}}
}

// dedupeBeta reaches the same three types all over again.
type dedupeBeta struct {
	dedupeEmbedded
	Shared dedupeShared
	Alpha  dedupeAlpha
}

func (dedupeBeta) DocComments() (string, map[string]FieldDoc) {
	return "dedupeBeta", map[string]FieldDoc{"Shared": {Comment: "Shared is documented."}}
}

var methodReceiver = regexp.MustCompile(`func \((\w+)\) DocComments\(\)`)

// TestRunProducesEachTypeOnce covers every way a struct can arrive at one Run: named as a root,
// nested in another root's tree, nested in two different structs at once, reached through each
// kind of container, and named as a root while also being nested.
//
// The arrangements have to agree with each other as well as be duplicate-free, since a config
// tree's shape is a property of the types and not of the order a caller happened to list them in.
func TestRunProducesEachTypeOnce(t *testing.T) {
	all := []string{"dedupeAlpha", "dedupeBeta", "dedupeEmbedded", "dedupeShared"}

	for _, tc := range []struct {
		name  string
		roots []any
		want  []string
	}{
		{
			name:  "One root, a type reached through every container and a cycle",
			roots: []any{&dedupeAlpha{}},
			want:  []string{"dedupeAlpha", "dedupeEmbedded", "dedupeShared"},
		},
		{
			name:  "Two roots, each nesting the same two types",
			roots: []any{&dedupeAlpha{}, &dedupeBeta{}},
			want:  all,
		},
		{
			name:  "A root that is also nested in the other root",
			roots: []any{&dedupeBeta{}, &dedupeAlpha{}},
			want:  all,
		},
		{
			name:  "A leaf named as a root as well as nested twice",
			roots: []any{&dedupeShared{}, &dedupeAlpha{}, &dedupeBeta{}},
			want:  all,
		},
		{
			name:  "The same root listed twice",
			roots: []any{&dedupeAlpha{}, &dedupeAlpha{}},
			want:  []string{"dedupeAlpha", "dedupeEmbedded", "dedupeShared"},
		},
		{
			name:  "Every type named as a root, all of them also nested",
			roots: []any{&dedupeShared{}, &dedupeEmbedded{}, &dedupeAlpha{}, &dedupeBeta{}},
			want:  all,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// A directory outside this module makes every type above a dependency's, so nothing
			// is parsed and nothing is written - the packages handed to the generator are the
			// whole observable result.
			dir := writePackage(t, map[string]string{"go.mod": "module example.com/x\n\ngo 1.26\n"})

			var captured []Package
			require.NoError(t, Run(RunArgs{Dir: dir, Roots: tc.roots, Tool: "example.com/x/gen"},
				Generator(func(pkgs []Package) (map[string]string, error) {
					captured = pkgs
					return nil, nil
				})))

			require.Len(t, captured, 1, "every type above is declared in one package")

			counts := map[string]int{}
			for _, typ := range captured[0].Types {
				counts[typ.Name]++
			}
			for name, count := range counts {
				require.Equal(t, 1, count, "%s discovered %d times", name, count)
			}
			require.ElementsMatch(t, tc.want, keysOf(counts))

			// A duplicate would reach the generated source as two methods on one receiver, which
			// does not compile - so the guarantee is checked where it is consumed too. The
			// package clause is supplied here because only a local package carries one, and
			// rendering is what is under test rather than where these types came from.
			local := captured[0]
			local.Name = "example"
			generated := docCommentsFile(local)
			_, err := parser.ParseFile(token.NewFileSet(), GeneratedFileName, generated, parser.SkipObjectResolution)
			require.NoError(t, err)
			for receiver, count := range receiverCounts(generated) {
				require.Equal(t, 1, count, "%s has %d DocComments methods", receiver, count)
			}
			require.ElementsMatch(t, tc.want, keysOf(receiverCounts(generated)))
		})
	}
}

// TestRunOrderDoesNotChangeTheResult pins what a CI diff relies on: the same types, however the
// roots were listed, produce the same output.
func TestRunOrderDoesNotChangeTheResult(t *testing.T) {
	generate := func(roots ...any) string {
		dir := writePackage(t, map[string]string{"go.mod": "module example.com/x\n\ngo 1.26\n"})
		var captured []Package
		require.NoError(t, Run(RunArgs{Dir: dir, Roots: roots, Tool: "example.com/x/gen"},
			Generator(func(pkgs []Package) (map[string]string, error) {
				captured = pkgs
				return nil, nil
			})))
		require.Len(t, captured, 1)
		local := captured[0]
		local.Name = "example"
		return docCommentsFile(local)
	}

	first := generate(&dedupeAlpha{}, &dedupeBeta{})
	require.Equal(t, first, generate(&dedupeBeta{}, &dedupeAlpha{}))
	require.Equal(t, first, generate(&dedupeShared{}, &dedupeBeta{}, &dedupeAlpha{}, &dedupeEmbedded{}))
	require.Equal(t, 1, strings.Count(first, "Flat is documented."))
}

// receiverCounts reports how many DocComments methods a generated file declares per receiver,
// which is what a duplicate would show up as.
func receiverCounts(source string) map[string]int {
	counts := map[string]int{}
	for _, match := range methodReceiver.FindAllStringSubmatch(source, -1) {
		counts[match[1]]++
	}
	return counts
}

func keysOf(counts map[string]int) []string {
	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	return names
}
