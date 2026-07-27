package discordsync

// True positives: genuinely harmful duplication that should be reported.
// These are multi-statement patterns with real logic that could be extracted.

// invokeService pattern — duplicated across multiple service handlers.
// Each handler fetches a service, checks availability, executes, wraps errors.
func invokeServiceA(ctx Context, req RequestA) (ResponseA, error) {
	svc := getServiceA(ctx)
	if svc == nil {
		return ResponseA{}, errServiceUnavailable
	}

	resp, err := svc.Execute(ctx, req)
	if err != nil {
		return ResponseA{}, fmtErrorf("service A failed: %w", err)
	}

	return resp, nil
}

func invokeServiceB(ctx Context, req RequestB) (ResponseB, error) {
	svc := getServiceB(ctx)
	if svc == nil {
		return ResponseB{}, errServiceUnavailable
	}

	resp, err := svc.Execute(ctx, req)
	if err != nil {
		return ResponseB{}, fmtErrorf("service B failed: %w", err)
	}

	return resp, nil
}

func invokeServiceC(ctx Context, req RequestC) (ResponseC, error) {
	svc := getServiceC(ctx)
	if svc == nil {
		return ResponseC{}, errServiceUnavailable
	}

	resp, err := svc.Execute(ctx, req)
	if err != nil {
		return ResponseC{}, fmtErrorf("service C failed: %w", err)
	}

	return resp, nil
}

// addIfPositive64 pattern — duplicated atomic counter helper.
// Both sites add a value to an atomic counter only if it's positive.
func addIfPositive64A(counter *int64, val int64) {
	if val > 0 {
		*counter += val
	}
}

func addIfPositive64B(counter *int64, val int64) {
	if val > 0 {
		*counter += val
	}
}
