package streaming

import (
	"io/ioutil"
	"math/big"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/luno/luno-go"
	"github.com/luno/luno-go/decimal"
)

type orderbookStatistics struct {
	Sequence  int64
	AskCount  int
	BidCount  int
	AskVolume decimal.Decimal
	BidVolume decimal.Decimal
}

func TestHandleMessageWithOrderbook(t *testing.T) {
	expectedCallbackCount := 0
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage([]byte("\"\""))
	mp.HandleMessage(loadFromFile(t, "fixture_orderbook.json"))
	mp.HandleMessage([]byte("\"\""))
	mp.HandleMessage([]byte("\"\""))

	expected := orderbookStatistics{
		Sequence:  40413238,
		AskCount:  9214,
		BidCount:  3248,
		AskVolume: decimal.New(big.NewInt(784815424), 6),
		BidVolume: decimal.New(big.NewInt(2695234253), 6),
	}

	actual := calculateOrderbookStatistics(mp.orderbook.GetSnapshot())

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestHandleMessageWithInvalidOrderbook(t *testing.T) {
	expectedCallbackCount := 0
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	err := mp.HandleMessage([]byte(`{"sequence": "40413238","asks": {"id": "BXEMZSYBRFYHSCF","price": "92655.00","volume": "0.495769"}, "bids": [{"id": "BXBAYA687URRT28","price": "92654.00","volume": "1.834379"}]}`))

	if err == nil {
		t.Errorf("Expected error to be returned")
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestHandleMessageWithDelete(t *testing.T) {
	expectedCallbackCount := 1
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage(loadFromFile(t, "fixture_orderbook.json"))
	mp.HandleMessage([]byte(`{"sequence":"40413239","trade_updates":null,"create_update":null,"delete_update":{"order_id":"BXNC7TGBBJJ885S"},"timestamp":1530887350936}`))

	expected := orderbookStatistics{
		Sequence:  40413239,
		AskCount:  9214,
		BidCount:  3247,
		AskVolume: decimal.New(big.NewInt(784815424), 6),
		BidVolume: decimal.New(big.NewInt(2692184753), 6),
	}

	actual := calculateOrderbookStatistics(mp.orderbook.GetSnapshot())

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestHandleMessageWithInvalidDelete(t *testing.T) {
	expectedCallbackCount := 0
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage(loadFromFile(t, "fixture_orderbook.json"))
	err := mp.HandleMessage([]byte(`{"sequence":"40413239","trade_updates":null,"create_update":null,"delete_update":{"order_id":{"order_id":"BXNC7TGBBJJ885S"}},"timestamp":1530887350936}`))

	if err == nil {
		t.Errorf("Expected error to be returned")
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestHandleMessageWithBuyTrade(t *testing.T) {
	expectedCallbackCount := 1
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage(loadFromFile(t, "fixture_orderbook.json"))
	mp.HandleMessage([]byte(`{"sequence":"40413239","trade_updates":[{"base":"0.094976","counter":"8800.00128","maker_order_id":"BXEMZSYBRFYHSCF","taker_order_id":"BXGGSPFECZKFQ34","order_id":"BXEMZSYBRFYHSCF"}],"create_update":null,"delete_update":null,"timestamp":1530887351827}`))

	expected := orderbookStatistics{
		Sequence:  40413239,
		AskCount:  9214,
		BidCount:  3248,
		AskVolume: decimal.New(big.NewInt(784720448), 6),
		BidVolume: decimal.New(big.NewInt(2695234253), 6),
	}

	actual := calculateOrderbookStatistics(mp.orderbook.GetSnapshot())

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestHandleMessageWithNonpositiveBuyTrade(t *testing.T) {
	expectedCallbackCount := 0
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage(loadFromFile(t, "fixture_orderbook.json"))
	err := mp.HandleMessage([]byte(`{"sequence":"40413239","trade_updates":[{"base":"-0.094976","counter":"8800.00128","maker_order_id":"BXEMZSYBRFYHSCF","taker_order_id":"BXGGSPFECZKFQ34","order_id":"BXEMZSYBRFYHSCF"}],"create_update":null,"delete_update":null,"timestamp":1530887351827}`))

	if err == nil {
		t.Errorf("Expected error to be returned")
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestHandleMessageWithOversizedBuyTrade(t *testing.T) {
	expectedCallbackCount := 0
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage(loadFromFile(t, "fixture_orderbook.json"))
	err := mp.HandleMessage([]byte(`{"sequence":"40413239","trade_updates":[{"base":"1.094976","counter":"8800.00128","maker_order_id":"BXEMZSYBRFYHSCF","taker_order_id":"BXGGSPFECZKFQ34","order_id":"BXEMZSYBRFYHSCF"}],"create_update":null,"delete_update":null,"timestamp":1530887351827}`))

	if err == nil {
		t.Errorf("Expected error to be returned")
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestHandleMessageWithSellTrade(t *testing.T) {
	expectedCallbackCount := 1
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage(loadFromFile(t, "fixture_orderbook.json"))
	mp.HandleMessage([]byte(`{"sequence":"40413239","trade_updates":[{"base":"1.834379","counter":"169962.55187","maker_order_id":"BXBAYA687URRT28","taker_order_id":"BXGGSPFECZKFQ34","order_id":"BXBAYA687URRT28"}],"create_update":null,"delete_update":null,"timestamp":1530887351827}`))

	expected := orderbookStatistics{
		Sequence:  40413239,
		AskCount:  9214,
		BidCount:  3247,
		AskVolume: decimal.New(big.NewInt(784815424), 6),
		BidVolume: decimal.New(big.NewInt(2693399874), 6),
	}

	actual := calculateOrderbookStatistics(mp.orderbook.GetSnapshot())

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestHandleMessageWithOversizedSellTrade(t *testing.T) {
	expectedCallbackCount := 0
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage(loadFromFile(t, "fixture_orderbook.json"))
	err := mp.HandleMessage([]byte(`{"sequence":"40413239","trade_updates":[{"base":"1.83438","counter":"169962.55187","maker_order_id":"BXBAYA687URRT28","taker_order_id":"BXGGSPFECZKFQ34","order_id":"BXBAYA687URRT28"}],"create_update":null,"delete_update":null,"timestamp":1530887351827}`))

	if err == nil {
		t.Errorf("Expected error to be returned")
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestHandleMessageWithTradeMatchingNoOrder(t *testing.T) {
	expectedCallbackCount := 0
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage(loadFromFile(t, "fixture_orderbook.json"))
	err := mp.HandleMessage([]byte(`{"sequence":"40413239","trade_updates":[{"base":"1.83438","counter":"169962.55187","maker_order_id":"BX_INVALID","taker_order_id":"BXGGSPFECZKFQ34","order_id":"BX_INVALID"}],"create_update":null,"delete_update":null,"timestamp":1530887351827}`))

	if err == nil {
		t.Errorf("Expected error to be returned")
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestHandleMessageWithCreateBid(t *testing.T) {
	expectedCallbackCount := 1
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage(loadFromFile(t, "fixture_orderbook.json"))
	mp.HandleMessage([]byte(`{"sequence":"40413239","trade_updates":null,"create_update":{"order_id":"BXKQ7P9GK27486F","type":"BID","price":"88501.00","volume":"3.0485"},"delete_update":null,"timestamp":1530887351155}`))

	expected := orderbookStatistics{
		Sequence:  40413239,
		AskCount:  9214,
		BidCount:  3249,
		AskVolume: decimal.New(big.NewInt(784815424), 6),
		BidVolume: decimal.New(big.NewInt(2698282753), 6),
	}

	actual := calculateOrderbookStatistics(mp.orderbook.GetSnapshot())

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestHandleMessageWithCreateAsk(t *testing.T) {
	expectedCallbackCount := 1
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage(loadFromFile(t, "fixture_orderbook.json"))
	mp.HandleMessage([]byte(`{"sequence":"40413239","trade_updates":null,"create_update":{"order_id":"BXKQ7P9GK27486F","type":"ASK","price":"92655.00","volume":"3.0485"},"delete_update":null,"timestamp":1530887351155}`))

	expected := orderbookStatistics{
		Sequence:  40413239,
		AskCount:  9215,
		BidCount:  3248,
		AskVolume: decimal.New(big.NewInt(787863924), 6),
		BidVolume: decimal.New(big.NewInt(2695234253), 6),
	}

	actual := calculateOrderbookStatistics(mp.orderbook.GetSnapshot())

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestHandleMessageWithCreateInvalidOrderType(t *testing.T) {
	expectedCallbackCount := 0
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage(loadFromFile(t, "fixture_orderbook.json"))
	err := mp.HandleMessage([]byte(`{"sequence":"40413239","trade_updates":null,"create_update":{"order_id":"BXKQ7P9GK27486F","type":"INVALID","price":"92655.00","volume":"3.0485"},"delete_update":null,"timestamp":1530887351155}`))

	if err == nil {
		t.Errorf("Expected error to be returned")
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestHandleMessageWithUpdateBeforeOrderbook(t *testing.T) {
	expectedCallbackCount := 0
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage([]byte(`{"sequence":"40413239","trade_updates":null,"create_update":{"order_id":"BXKQ7P9GK27486F","type":"BID","price":"88501.00","volume":"3.0485"},"delete_update":null,"timestamp":1530887351155}`))

	actualSeq, actualBids, actualAsks := mp.orderbook.GetSnapshot()

	if 0 != actualSeq {
		t.Errorf("Expected sequence to be 0, got %v", actualSeq)
	}
	if nil != actualBids {
		t.Errorf("Expected bids to be nil, got %v", actualBids)
	}
	if nil != actualAsks {
		t.Errorf("Expected asks to be nil, got %v", actualAsks)
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestHandleMessageWithPreviousUpdate(t *testing.T) {
	expectedCallbackCount := 0
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage(loadFromFile(t, "fixture_orderbook.json"))
	mp.HandleMessage([]byte(`{"sequence":"40413237","trade_updates":null,"create_update":{"order_id":"BXKQ7P9GK27486F","type":"BID","price":"88501.00","volume":"3.0485"},"delete_update":null,"timestamp":1530887351155}`))
	mp.HandleMessage([]byte(`{"sequence":"40413238","trade_updates":null,"create_update":{"order_id":"BXKQ7P9GK27486F","type":"BID","price":"88501.00","volume":"3.0485"},"delete_update":null,"timestamp":1530887351155}`))

	expected := orderbookStatistics{
		Sequence:  40413238,
		AskCount:  9214,
		BidCount:  3248,
		AskVolume: decimal.New(big.NewInt(784815424), 6),
		BidVolume: decimal.New(big.NewInt(2695234253), 6),
	}

	actual := calculateOrderbookStatistics(mp.orderbook.GetSnapshot())

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestHandleMessageWithOutOfSequenceUpdate(t *testing.T) {
	expectedCallbackCount := 0
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage(loadFromFile(t, "fixture_orderbook.json"))
	err := mp.HandleMessage([]byte(`{"sequence":"40413240","trade_updates":null,"create_update":{"order_id":"BXKQ7P9GK27486F","type":"BID","price":"88501.00","volume":"3.0485"},"delete_update":null,"timestamp":1530887351155}`))

	if err == nil {
		t.Errorf("Expected error to be returned")
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func TestReset(t *testing.T) {
	expectedCallbackCount := 0
	var callbackCount int

	mp := &messageProcessor{
		updateCallback: func(update UpdateMessage) {
			callbackCount++
		},
	}

	mp.HandleMessage(loadFromFile(t, "fixture_orderbook.json"))
	mp.Reset()

	actualSeq, actualBids, actualAsks := mp.orderbook.GetSnapshot()

	if 0 != actualSeq {
		t.Errorf("Expected sequence to be 0, got %v", actualSeq)
	}
	if nil != actualBids {
		t.Errorf("Expected bids to be nil, got %v", actualBids)
	}
	if nil != actualAsks {
		t.Errorf("Expected asks to be nil, got %v", actualAsks)
	}

	if expectedCallbackCount != callbackCount {
		t.Errorf("Expected callback to be called %d times, instead of %d times", expectedCallbackCount, callbackCount)
	}
}

func calculateOrderbookStatistics(sequence int64, bids []luno.OrderBookEntry, asks []luno.OrderBookEntry) orderbookStatistics {
	var stats = orderbookStatistics{
		Sequence:  sequence,
		AskCount:  len(asks),
		BidCount:  len(bids),
		AskVolume: decimal.New(new(big.Int), 6),
		BidVolume: decimal.New(new(big.Int), 6),
	}

	for _, ask := range asks {
		stats.AskVolume = stats.AskVolume.Add(ask.Volume)
	}

	for _, bid := range bids {
		stats.BidVolume = stats.BidVolume.Add(bid.Volume)
	}

	return stats
}

func loadFromFile(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}
