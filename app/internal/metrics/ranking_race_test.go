package metrics

import (
	"fmt"
	"sync"
	"testing"
)

// Regression: Ranking used to iterate usageCounts without usageMu while
// RecordUsage inserted new tools — a fatal "concurrent map iteration and
// map write" under load. Run with -race to catch a reintroduction.
func TestRankingConcurrentWithRecordUsage(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Go(func() { RecordUsage(fmt.Sprintf("race-tool-%d", i)) })
		wg.Go(func() { _ = Ranking() })
	}
	wg.Wait()

	ranking := Ranking()
	for i := 1; i < len(ranking); i++ {
		if ranking[i-1].Count < ranking[i].Count {
			t.Fatalf("ranking not sorted by count desc: %+v", ranking)
		}
	}
}
