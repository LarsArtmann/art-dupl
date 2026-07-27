package discordsync

// Finding 2: queryError wrapping — 27 groups in real codebase.
// Each site calls queryError with a unique operation string.
// The wrapper IS the already-extracted helper; unique strings are parameters.
// Expected actionability: error-wrapping (suppressed via structural matching).
//
// Fixture design: each function has unique trailing code so the clone detector
// isolates just the if-statement, not the trailing return.

func queryStaleProgress(ctx Context) (Progress, error) {
	progress, err := db.QueryStale(ctx)
	if err != nil {
		return nil, queryError(err, "failed to query stale backfill progress")
	}

	validateStale(progress)

	return progress, nil
}

func queryBackfillProgress(ctx Context) (Progress, error) {
	progress, err := db.QueryProgress(ctx)
	if err != nil {
		return nil, queryError(err, "failed to query backfill progress")
	}

	logBackfill(progress)

	return progress, nil
}

func queryGuildMember(ctx Context) (Member, error) {
	member, err := db.QueryMember(ctx)
	if err != nil {
		return nil, queryError(err, "failed to get guild member")
	}

	enrichMember(member)

	return member, nil
}

func queryChannelMessages(ctx Context) ([]Message, error) {
	messages, err := db.QueryMessages(ctx)
	if err != nil {
		return nil, queryError(err, "failed to query channel messages")
	}

	sortMessages(messages)

	return messages, nil
}

func queryRolePermissions(ctx Context) (Permissions, error) {
	perms, err := db.QueryPermissions(ctx)
	if err != nil {
		return nil, queryError(err, "failed to query role permissions")
	}

	mergeDefaults(perms)

	return perms, nil
}

func queryAttachmentMeta(ctx Context) (Meta, error) {
	meta, err := db.QueryMeta(ctx)
	if err != nil {
		return nil, queryError(err, "failed to query attachment metadata")
	}

	validateMeta(meta)

	return meta, nil
}
