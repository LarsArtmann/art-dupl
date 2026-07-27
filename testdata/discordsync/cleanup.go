package discordsync

// Finding 3: defer cleanup — 12 groups in real codebase.
// Three variants: bare cancel(), FuncLit wrapping rows.Close(), FuncLit tx.Rollback().
// Expected actionability: raii-defer (suppressed).
//
// Fixture design: each function has unique trailing code so the clone detector
// isolates just the defer/cancel pattern.

func handleEventA(ctx Context) {
	ctx, cancel := newHandlerContext()
	defer cancel()
	dispatchEventA(ctx)
}

func handleEventB(ctx Context) {
	ctx, cancel := newHandlerContext()
	defer cancel()
	dispatchEventB(ctx)
}

func handleEventC(ctx Context) {
	ctx, cancel := newHandlerContext()
	defer cancel()
	dispatchEventC(ctx)
}

func processAttachmentRows(rows Rows) {
	defer func() { _ = rows.Close() }()
	scanAttachments(rows)
}

func processMessageRows(rows Rows) {
	defer func() { _ = rows.Close() }()
	scanMessages(rows)
}

func processMemberRows(rows Rows) {
	defer func() { _ = rows.Close() }()
	scanMembers(rows)
}

func transactionalInsertA(tx Tx) {
	defer func() { _ = tx.Rollback() }()
	insertRecordA(tx)
}

func transactionalInsertB(tx Tx) {
	defer func() { _ = tx.Rollback() }()
	insertRecordB(tx)
}

func transactionalInsertC(tx Tx) {
	defer func() { _ = tx.Rollback() }()
	insertRecordC(tx)
}
