package messaging

import "strings"

// ParseBootstrapServers splits comma-separated seed brokers (e.g. KAFKA_BOOTSTRAP_SERVERS).
func ParseBootstrapServers(s string) []string {
	var out []string
	for p := range strings.SplitSeq(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
