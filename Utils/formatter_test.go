package utils

import (
	"fmt"
	"testing"
	"time"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}
	message := fmt.Sprintf("%v != %v", a, b)
	t.Fatal(message)
}

func TestByteCountIEC(t *testing.T) {

	assertEqual(t, ByteCountIEC(1 << 10), "1.0 KiB")
	assertEqual(t, ByteCountIEC(0), "0 B")
	assertEqual(t, ByteCountIEC(1023), "1023 B")
	assertEqual(t, ByteCountIEC(1 << 20), "1.0 MiB")
	assertEqual(t, ByteCountIEC(1 << 30), "1.0 GiB")
	assertEqual(t, ByteCountIEC(1 << 40), "1.0 TiB")
}

func TestDurationToString(t *testing.T) {
	assertEqual(t, DurationToString(time.Microsecond*10), "10µs")
	assertEqual(t, DurationToString(time.Millisecond*10 + time.Second*2), "2.01s")
	assertEqual(t, DurationToString(time.Millisecond*100 + time.Second*2), "2.1s")
	// Should round up
	assertEqual(t, DurationToString(time.Millisecond*19 + time.Second*2), "2.02s")
	// Should round down
	assertEqual(t, DurationToString(time.Millisecond*14 + time.Second*2), "2.01s")
}
