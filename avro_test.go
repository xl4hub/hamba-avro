package avro_test

import "github.com/xl4hub/hamba-avro"

func ConfigTeardown() {
	// Reset the caches
	avro.DefaultConfig = avro.Config{}.Freeze()
}
