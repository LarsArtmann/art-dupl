// Package discordsync contains regression test fixtures extracted from
// real-world DiscordSync feedback. These files are NOT compiled (testdata is
// ignored by go build) but ARE parsed by art-dupl to validate actionability
// classification on real-world Go idioms.
package discordsync

// Finding 1: HTTP error guards — 12 groups in real codebase.
// The bare return is forced by http.HandlerFunc's void signature.
// Expected actionability: error-propagation (suppressed).

func handleGetMessage(w ResponseWriter, r *Request) {
	bundle, err := database.GetMessageBundle(r.Context(), messageID)
	if err != nil {
		writeError(w, r, err, "")

		return
	}

	renderJSON(w, bundle)
}

func handleGetGuilds(w ResponseWriter, r *Request) {
	guilds, err := database.GetGuilds(r.Context())
	if err != nil {
		writeError(w, r, err, "")

		return
	}

	renderJSON(w, guilds)
}

func handleGetChannels(w ResponseWriter, r *Request) {
	channels, err := database.GetChannels(r.Context())
	if err != nil {
		writeError(w, r, err, "")

		return
	}

	renderJSON(w, channels)
}

func handleGetRoles(w ResponseWriter, r *Request) {
	roles, err := database.GetRoles(r.Context())
	if err != nil {
		writeError(w, r, err, "")

		return
	}

	renderJSON(w, roles)
}

func handleGetMembers(w ResponseWriter, r *Request) {
	members, err := database.GetMembers(r.Context())
	if err != nil {
		writeError(w, r, err, "")

		return
	}

	renderJSON(w, members)
}

func handleGetAttachments(w ResponseWriter, r *Request) {
	attachments, err := database.GetAttachments(r.Context())
	if err != nil {
		writeError(w, r, err, "")

		return
	}

	renderJSON(w, attachments)
}
