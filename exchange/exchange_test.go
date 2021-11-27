package exchange

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestGetData(t *testing.T) {
	r := NewRestClient("Test", "https://api.coinmarketcap.com/v1/ticker/?convert=INR", "15s")
	Result, err := r.GetData()
	if err != nil {
		t.Errorf("Error occurred : %v", err)
		return
	}

	for k, v := range Result {
		fmt.Printf("Key: %d , Value, %v\n", k, v)
	}
}

func BytesToString(data []byte) string {
	return string(data[:])
}
