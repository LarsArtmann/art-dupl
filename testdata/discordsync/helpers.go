package discordsync

// Finding 6: helper invocation + bool guard — 5 groups in real codebase.
// X, ok := sharedHelper(...); if !ok { return }
// The helper IS the extraction; calling it N times is the goal.
// Expected actionability: bool-guard (suppressed).

func handleGuildEndpoint1(w ResponseWriter, r *Request) {
	guildID, ok := requireQueryParam(w, r, "guild_id")
	if !ok {
		return
	}

	renderGuild(w, guildID)
}

func handleGuildEndpoint2(w ResponseWriter, r *Request) {
	guildID, ok := requireQueryParam(w, r, "guild_id")
	if !ok {
		return
	}

	renderGuildList(w, guildID)
}

func handleGuildEndpoint3(w ResponseWriter, r *Request) {
	guildID, ok := requireQueryParam(w, r, "guild_id")
	if !ok {
		return
	}

	renderGuildSettings(w, guildID)
}
