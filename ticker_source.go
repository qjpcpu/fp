package fp

import (
	"reflect"
	"time"
)

func NewTickerSource(interval time.Duration) *TickerSource {
	ticker := time.NewTicker(interval)
	return &TickerSource{ticker: ticker}
}

type TickerSource struct {
	ticker *time.Ticker
}

func (cs *TickerSource) Stop()                  { cs.ticker.Stop() }
func (cs *TickerSource) ElemType() reflect.Type { return reflect.TypeOf(time.Time{}) }
func (cs *TickerSource) Next() (reflect.Value, bool) {
	if tm, ok := <-cs.ticker.C; ok {
		return reflect.ValueOf(tm), true
	}
	return reflect.Value{}, false
}

func NewDelaySource(interval time.Duration) Source {
	return &delaySource{interval: interval}
}

type delaySource struct {
	interval time.Duration
}

func (cs *delaySource) ElemType() reflect.Type { return reflect.TypeOf(time.Time{}) }
func (cs *delaySource) Next() (reflect.Value, bool) {
	time.Sleep(cs.interval)
	return reflect.ValueOf(time.Now()), true
}
