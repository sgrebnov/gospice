package gospice

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/apache/arrow/go/v14/arrow/array"
)

// Execute a basic query and check for columns and rows
func TestFireQuery(t *testing.T) {
	spice := NewSpiceClient()
	defer spice.Close()

	var ApiKey string
	if v, exists := os.LookupEnv("SPICE_API_KEY"); exists {
		ApiKey = v
	} else {
		ApiKey = TEST_API_KEY
	}

	if err := spice.Init(ApiKey); err != nil {
		panic(fmt.Errorf("error initializing SpiceClient: %w", err))
	}

	t.Run("Recent Ethereum Blocks", func(t *testing.T) {
		reader, err := spice.FireQuery(context.Background(), "SELECT number, \"timestamp\", hash FROM eth.recent_blocks ORDER BY number LIMIT 10")
		if err != nil {
			panic(fmt.Errorf("error querying: %w", err))
		}
		defer reader.Release()

		schema := reader.Schema()
		if !schema.HasField("number") {
			t.Fatalf("Schema does not have field 'number'")
		}
		if !schema.HasField("timestamp") {
			t.Fatalf("Schema does not have field 'timestamp'")
		}
		if !schema.HasField("hash") {
			t.Fatalf("Schema does not have field 'hash'")
		}

		for reader.Next() {
			record := reader.Record()
			defer record.Release()

			if record.NumRows() != 10 {
				t.Fatalf("Expected 10 rows, got %d", record.NumRows())
			}

			col0 := record.Column(0)
			defer col0.Release()

			blockNumber := col0.(*array.Int64).Value(0)
			if blockNumber <= 16410468 {
				t.Fatalf("Expected block number > 16410468, got %d", blockNumber)
			}

			col1 := record.Column(1)
			defer col1.Release()

			timestamp := col1.(*array.Int64).Value(0)
			fiveMinutesAgo := time.Now().Add(-time.Minute * 5).Unix()
			if timestamp > fiveMinutesAgo {
				t.Fatalf("Expected timestamp > %d, got %d", fiveMinutesAgo, timestamp)
			}

			col2 := record.Column(2)
			defer col2.Release()

			hash := col2.(*array.String).Value(0)
			if len(hash) != 66 {
				t.Fatalf("Expected hash length 66, got %d", len(hash))
			}
		}
	})
}
