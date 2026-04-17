package headers

import "strings"

type Decision struct {
	Allow             bool
	RequiresChallenge bool
}

type Analyzer struct {
	enabled         bool
	blockedKeywords []string
	seoAllowlist    []string
}

func NewAnalyzer(enabled bool, blockedKeywords []string, seoAllowlist []string) *Analyzer {
	normalize := func(values []string) []string {
		out := make([]string, 0, len(values))
		for _, v := range values {
			v = strings.ToLower(strings.TrimSpace(v))
			if v != "" {
				out = append(out, v)
			}
		}
		return out
	}
	return &Analyzer{
		enabled:         enabled,
		blockedKeywords: normalize(blockedKeywords),
		seoAllowlist:    normalize(seoAllowlist),
	}
}

func (a *Analyzer) Analyze(userAgent string) Decision {
	if !a.enabled {
		return Decision{Allow: true}
	}
	ua := strings.ToLower(strings.TrimSpace(userAgent))
	if ua == "" {
		return Decision{Allow: false, RequiresChallenge: true}
	}
	for _, bot := range a.seoAllowlist {
		if strings.Contains(ua, bot) {
			return Decision{Allow: true}
		}
	}
	for _, keyword := range a.blockedKeywords {
		if strings.Contains(ua, keyword) {
			return Decision{Allow: false, RequiresChallenge: true}
		}
	}
	return Decision{Allow: true}
}
