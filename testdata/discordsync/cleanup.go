package discordsync

// Finding 3: defer cleanup — 12 groups in real codebase.
// Three variants: bare cancel(), FuncLit wrapping rows.Close(), FuncLit tx.Rollback().
// Expected actionability: raii-defer (suppressed).

func handleEventA(ctx Context) {
	ctx, cancel := newHandlerContext()
	defer cancel()
	processEvent(ctx)
}

func handleEventB(ctx Context) {
	ctx, cancel := newHandlerContext()
	defer cancel()
	processEvent(ctx)
}

func handleEventC(ctx Context) {
	ctx, cancel := newHandlerContext()
	defer cancel()
	processEvent(ctx)
}

func processAttachmentRows(rows Rows) {
	defer func() { _ = rows.Close() }()
	scanRows(rows)
}

func processMessageRows(rows Rows) {
	defer func() { _ = rows.Close() }()
	scanRows(rows)
}

func processMemberRows(rows Rows) {
	defer func() { _ = rows.Close() }()
	scanRows(rows)
}

func transactionalInsertA(tx Tx) {
	defer func() { _ = tx.Rollback() }()
	insertRecord(tx)
}

func transactionalInsertB(tx Tx) {
	defer func() { _ = tx.Rollback() }()
	insertRecord(tx)
}

func transactionalInsertC(tx Tx) {
	defer func() { _ = tx.Rollback() }()
	insertRecord(tx)
}
