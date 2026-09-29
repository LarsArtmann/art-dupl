#!/usr/bin/env bash
# check-annotator.sh — fixture tests for scripts/annotate-status-items.py.
# Exercises: numbered/table/checkbox/any:/@-disambiguation strikes, duplicate-key
# refusal, banned-verdict refusal, atomicity (one miss = no write), --verify
# (prints, writes nothing), and --emit-keys. Exit 0 = all fixtures pass.
set -euo pipefail
cd "$(dirname "$0")/.."

ANN=scripts/annotate-status-items.py
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
pass=0

expect() { # expect <desc> <want> <got>
	if [ "$2" = "$3" ]; then
		pass=$((pass + 1))
	else
		echo "FIXTURE FAIL: $1" >&2
		echo "  want: $2" >&2
		echo "  got:  $3" >&2
		exit 1
	fi
}

# --- fixture file: two numbered lists (restarting 1), a table, checkboxes ---
cat > "$WORK/doc.md" << 'EOF'
## List one

1. First item alpha
2. Second item beta

## List two (restarts numbering)

1. First item gamma
2. Second item delta

| # | Task | Status |
| --- | --- | --- |
| 3 | Table row three | open |
| 4 | Table row four | open |

- [ ] checkbox one
- [ ] checkbox two

Free prose mentioning item beta for any-mode.
EOF
cp "$WORK/doc.md" "$WORK/doc.pristine"

# --- emit-keys: keys generated mechanically for line numbers ---
cat > "$WORK/keys.tsv" << 'EOF'
EOF
python3 "$ANN" --emit-keys "$WORK/doc.md" 3 4 9 10 14 15 16 > "$WORK/keys.out" 2> "$WORK/keys.preview"
expect "emit-keys count" "6" "$(wc -l < "$WORK/keys.out")"
expect "emit-keys disambiguates restart" "1@First item gamma" "$(sed -n 2p "$WORK/keys.out")"
expect "emit-keys preview names line 9" "9: 1. First item gamma" "$(sed -n 2p "$WORK/keys.preview")"

# build the real spec from the emitted keys + hand verdicts
{
	head -2 "$WORK/keys.out" | sed 's/$/\t/'
	sed -n '3p' "$WORK/keys.out" | sed 's/$/\tdone — table verdict/'
	sed -n '4p' "$WORK/keys.out" | sed 's/$/\t/'
	echo -e "any:item beta\t"
	echo -e "checkbox one\t"
} > "$WORK/spec.tsv"

# --- dup-key refusal: BEFORE any file read, nothing written ---
cp "$WORK/doc.pristine" "$WORK/doc.md"
cat "$WORK/spec.tsv" > "$WORK/dup.tsv"
head -1 "$WORK/spec.tsv" >> "$WORK/dup.tsv"
python3 "$ANN" "$WORK/doc.md" "$WORK/dup.tsv" > "$WORK/dup.out" 2>&1 && \
	expect "dup-key exit" "nonzero" "zero" || [ $? -ne 0 ] || true
expect "dup-key refuses" "1" "$?"
cmp -s "$WORK/doc.pristine" "$WORK/doc.md"
expect "dup-key wrote nothing" "same" "same"

# --- banned verdict refusal ---
printf 'checkbox two\topen\n' > "$WORK/banned.tsv"
python3 "$ANN" "$WORK/doc.md" "$WORK/banned.tsv" >/dev/null 2>&1 && \
	expect "banned exit" "nonzero" "zero" || true
expect "banned verdict refuses" "1" "$?"
cmp -s "$WORK/doc.pristine" "$WORK/doc.md"
expect "banned wrote nothing" "same" "same"

# --- atomicity: one unmatched key aborts the whole spec set ---
cp "$WORK/doc.pristine" "$WORK/doc.md"
{ cat "$WORK/spec.tsv"; printf '999\t\n'; } > "$WORK/atomic.tsv"
python3 "$ANN" "$WORK/doc.md" "$WORK/atomic.tsv" >/dev/null 2>&1 && \
	expect "atomic exit" "nonzero" "zero" || true
expect "atomic aborts" "1" "$?"
cmp -s "$WORK/doc.pristine" "$WORK/doc.md"
expect "atomic wrote nothing" "same" "same"

# --- verify mode: resolves and prints, writes nothing ---
python3 "$ANN" --verify "$WORK/doc.md" "$WORK/spec.tsv" > "$WORK/verify.out" 2>&1
expect "verify exit" "0" "$?"
expect "verify resolves 6 keys" "VERIFY OK doc.md: 6 keys resolved (nothing written)" "$(tail -1 "$WORK/verify.out")"
cmp -s "$WORK/doc.pristine" "$WORK/doc.md"
expect "verify wrote nothing" "same" "same"

# --- strike mode: all six keys land, restart list disambiguated ---
python3 "$ANN" "$WORK/doc.md" "$WORK/spec.tsv" > "$WORK/strike.out" 2>&1
expect "strike exit" "0" "$?"
expect "strike reports 6" "OK doc.md: annotated 6 items" "$(cat "$WORK/strike.out")"
expect "list-one 1 struck" "~~1. First item alpha~~" "$(sed -n 3p "$WORK/doc.md" | sed 's/\t.*//')"
expect "list-two restart struck" "~~1. First item gamma~~" "$(sed -n 9p "$WORK/doc.md" | sed 's/\t.*//')"
expect "table row verdict" "~~| 3 | Table row three | open |~~ done — table verdict" "$(sed -n 12p "$WORK/doc.md" | sed 's/ $//')"
expect "any-mode prose struck" "~~Free prose mentioning item beta for any-mode.~~" "$(sed -n 17p "$WORK/doc.md" | sed 's/\t.*//')"
expect "checkbox struck" "~~- [ ] checkbox one~~" "$(sed -n 15p "$WORK/doc.md" | sed 's/\t.*//')"
expect "unspecified checkbox stays bare" "- [ ] checkbox two" "$(sed -n 16p "$WORK/doc.md")"

echo "annotator fixtures OK ($pass assertions)"
