package store

type reportTargetSpec struct {
	table         string
	contentColumn string
}

func reportTargetSpecFor(targetType string) (reportTargetSpec, bool) {
	switch targetType {
	case reportTargetDebate:
		return reportTargetSpec{table: "debates", contentColumn: "topic"}, true
	case reportTargetTurn:
		return reportTargetSpec{table: "turns", contentColumn: "content"}, true
	case reportTargetComment:
		return reportTargetSpec{table: "comments", contentColumn: "content"}, true
	default:
		return reportTargetSpec{}, false
	}
}
