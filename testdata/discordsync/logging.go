package discordsync

// Finding 9: single-site error logging — unique messages per site.
// Each site has a unique message, unique context keys.
// Expected actionability: error-propagation or single-call-expression (suppressed).

func logAttachmentError(err error) {
	if err != nil {
		logError("failed to mark attachment as failed", err)
	}
}

func logQueryError(err error) {
	if err != nil {
		logError("failed to query attachments", err)
	}
}

func logSyncError(err error) {
	if err != nil {
		logError("failed to sync messages", err)
	}
}
