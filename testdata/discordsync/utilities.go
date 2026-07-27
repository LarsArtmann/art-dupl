package discordsync

// Finding 7: cross-package nil-guard utility — borderline true positive.
// 3-line nil-guard utilities in different packages.
// Expected: reported (not suppressed) — these ARE semantic clones but small.

func webhookIDStr(id *Snowflake) string {
	if id == nil {
		return ""
	}

	return id.String()
}

func optSnowflakeToString(id *Snowflake) string {
	if id == nil {
		return ""
	}

	return id.String()
}
