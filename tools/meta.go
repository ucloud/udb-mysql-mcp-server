package tools

// Mode controls which tool risk levels are registered.
type Mode string

const (
	ModeReadonly  Mode = "readonly"
	ModeReadWrite Mode = "readwrite"
	ModeAdmin     Mode = "admin"
)

// RiskLevel classifies the operational impact of a tool.
type RiskLevel string

const (
	RiskRead      RiskLevel = "read"
	RiskWrite     RiskLevel = "write"
	RiskWriteLow  RiskLevel = "write-low"
	RiskWriteMid  RiskLevel = "write-mid"
	RiskWriteHigh RiskLevel = "write-high"
	RiskCritical  RiskLevel = "critical"
)

// Tool is local catalog metadata used to filter tools by mode.
type Tool struct {
	Name        string
	Description string
	Risk        RiskLevel
}

// ModeAllowsRisk reports whether a tool risk is permitted under the server mode.
func ModeAllowsRisk(mode Mode, risk RiskLevel) bool {
	switch mode {
	case ModeReadonly:
		return risk == RiskRead
	case ModeReadWrite:
		switch risk {
		case RiskRead, RiskWrite, RiskWriteLow, RiskWriteMid:
			return true
		default:
			return false
		}
	case ModeAdmin:
		switch risk {
		case RiskRead, RiskWrite, RiskWriteLow, RiskWriteMid, RiskWriteHigh, RiskCritical:
			return true
		default:
			return false
		}
	default:
		return false
	}
}
