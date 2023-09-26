package loop_test

import (
	"sync"
	"testing"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop"
	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/test"
)

func TestScale(t *testing.T) {
	start := make(chan struct{})
	var wg, ready sync.WaitGroup
	stopCh := newStopCh(t)

	const rCount = 50
	const mCount = 100
	relayers := make([]loop.PluginRelayer, rCount)
	for i := 0; i < rCount; i++ {
		ready.Add(1)
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			relayers[i] = newPluginRelayerExec(t, stopCh)
			ready.Done()
			<-start
			test.TestPluginRelayer(t, relayers[i])
		}(i)
	}
	ready.Wait()
	for i := 0; i < mCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			pr := relayers[i%rCount]
			pm := newPluginMedianExec(t, stopCh)
			p := newMedianProvider(t, pr)
			<-start
			test.PluginMedianTest{MedianProvider: p}.TestPluginMedian(t, pm)
		}(i)
	}

	close(start)
	wg.Wait()
}
