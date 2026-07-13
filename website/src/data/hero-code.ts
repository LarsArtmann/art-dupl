export const heroCode = `$ art-dupl

Found 23 clone groups (148 duplicate statements)

  Group 1  [Type 2 — renamed]  42 tokens  3 occurrences
  ├─ internal/handler/user.go:45-67
  ├─ internal/handler/order.go:38-60
  └─ internal/handler/product.go:51-73

  Group 2  [Type 1 — exact]  28 tokens  2 occurrences
  ├─ pkg/cache/lru.go:112-125
  └─ pkg/cache/ttl.go:98-111

  ... 21 more groups

Health: B  |  12 files affected  |  2,840 tokens duplicated

$ art-dupl --html > report.html    # Interactive HTML report
$ art-dupl stats                    # Project health statistics`;
