package sources

import (
	"context"
	"iter"
	"net/http"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// collectResults runs a source iterator and collects results by type.
func collectResults(seq iter.Seq2[Result, error]) (subdomains, urls []Result, errors []error) {
	for r, err := range seq {
		if err != nil {
			errors = append(errors, err)
			continue
		}
		switch r.Type {
		case Subdomain:
			subdomains = append(subdomains, r)
		case URL:
			urls = append(urls, r)
		}
	}
	return
}

// assertResults validates that all results have the expected source, type, and non-empty value.
func assertResults(t *testing.T, results []Result, source string, resultType ResultType) {
	t.Helper()

	for _, r := range results {
		assert.Equal(t, source, r.Source)
		assert.Equal(t, resultType, r.Type)
		assert.NotEmpty(t, r.Value)
	}
}

func TestRegister(t *testing.T) {
	t.Parallel()

	testSource := Source{
		Name:   "test-register-source",
		Yields: Subdomain,
		Run: func(_ context.Context, _ *http.Client, _ string, _ string) iter.Seq2[Result, error] {
			return func(_ func(Result, error) bool) {}
		},
	}

	Register(testSource)

	got := ByName("test-register-source")
	require.NotNil(t, got)
	assert.Equal(t, "test-register-source", got.Name)
	assert.Equal(t, Subdomain, got.Yields)
}

func TestByName(t *testing.T) {
	t.Parallel()

	t.Run("returns_registered_source", func(t *testing.T) {
		testSource := Source{
			Name:   "test-byname-source",
			Yields: Subdomain,
			Run: func(_ context.Context, _ *http.Client, _ string, _ string) iter.Seq2[Result, error] {
				return func(_ func(Result, error) bool) {}
			},
		}
		Register(testSource)

		got := ByName("test-byname-source")

		require.NotNil(t, got)
		assert.Equal(t, "test-byname-source", got.Name)
	})

	t.Run("returns_nil_not_found", func(t *testing.T) {
		got := ByName("nonexistent-source")

		assert.Nil(t, got)
	})
}

func TestAll(t *testing.T) {
	t.Parallel()

	testSource := Source{
		Name:   "test-list-source",
		Yields: URL,
		Run: func(_ context.Context, _ *http.Client, _ string, _ string) iter.Seq2[Result, error] {
			return func(_ func(Result, error) bool) {}
		},
	}
	Register(testSource)

	list := All()

	require.NotEmpty(t, list)

	found := false
	for _, s := range list {
		if s.Name == "test-list-source" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestFilter(t *testing.T) {
	t.Parallel()

	subSource := Source{
		Name:   "test-filter-subdomain",
		Yields: Subdomain,
		Run: func(_ context.Context, _ *http.Client, _ string, _ string) iter.Seq2[Result, error] {
			return func(_ func(Result, error) bool) {}
		},
	}
	urlSource := Source{
		Name:   "test-filter-url",
		Yields: URL,
		Run: func(_ context.Context, _ *http.Client, _ string, _ string) iter.Seq2[Result, error] {
			return func(_ func(Result, error) bool) {}
		},
	}
	bothSource := Source{
		Name:   "test-filter-both",
		Yields: Subdomain | URL,
		Run: func(_ context.Context, _ *http.Client, _ string, _ string) iter.Seq2[Result, error] {
			return func(_ func(Result, error) bool) {}
		},
	}

	Register(subSource)
	Register(urlSource)
	Register(bothSource)

	t.Run("subdomain_filter", func(t *testing.T) {
		subFiltered := ByType(Subdomain)
		subNames := make([]string, len(subFiltered))
		for i, s := range subFiltered {
			subNames[i] = s.Name
		}

		assert.True(t, slices.Contains(subNames, "test-filter-subdomain"))
		assert.True(t, slices.Contains(subNames, "test-filter-both"))
		assert.False(t, slices.Contains(subNames, "test-filter-url"))
	})

	t.Run("url_filter", func(t *testing.T) {
		urlFiltered := ByType(URL)
		urlNames := make([]string, len(urlFiltered))
		for i, s := range urlFiltered {
			urlNames[i] = s.Name
		}

		assert.True(t, slices.Contains(urlNames, "test-filter-url"))
		assert.True(t, slices.Contains(urlNames, "test-filter-both"))
		assert.False(t, slices.Contains(urlNames, "test-filter-subdomain"))
	})
}

func TestNames(t *testing.T) {
	t.Parallel()

	testSource := Source{
		Name:   "test-names-source",
		Yields: Subdomain,
		Run: func(_ context.Context, _ *http.Client, _ string, _ string) iter.Seq2[Result, error] {
			return func(_ func(Result, error) bool) {}
		},
	}
	Register(testSource)

	names := Names()

	require.NotEmpty(t, names)
	assert.True(t, slices.Contains(names, "test-names-source"))
}

func TestResultType(t *testing.T) {
	t.Parallel()

	t.Run("distinct_flags", func(t *testing.T) {
		assert.Zero(t, Subdomain&URL)
	})

	t.Run("combined_includes_both", func(t *testing.T) {
		both := Subdomain | URL

		assert.NotZero(t, both&Subdomain)
		assert.NotZero(t, both&URL)
	})
}
