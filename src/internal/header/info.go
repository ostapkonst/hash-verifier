package header

import (
	"fmt"
	"time"

	"github.com/ostapkonst/hash-verifier/utils/eof"
)

const (
	Name = "HashVerifier"
	Link = "https://github.com/ostapkonst/hash-verifier"
)

// Version устанавливается при компиляции через -ldflags -X.
var Version = "unknown"

func GetChecksumHeader() string {
	nowUTC := time.Now().UTC()
	rfc3339 := nowUTC.Format(time.RFC3339)

	return fmt.Sprintf(
		"; Generated at %s by %s %s (%s)%s%s",
		rfc3339,
		Name,
		Version,
		Link,
		eof.PlatformEOF,
		eof.PlatformEOF,
	)
}
