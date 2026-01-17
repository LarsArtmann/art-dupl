#!/usr/bin/env python3
import re

with open('bdd_test.go', 'r') as f:
    content = f.read()

# Replace build commands
content = re.sub(
    r'exec\.Command\("go", "build", "-o", "\.\./bdd/art-dupl-test", "\.\)"',
    'exec.Command("go", "build", "-o", "bdd/art-dupl-test", "./cmd/art-dupl")',
    content
)

# Remove cmd.Dir lines
content = re.sub(r'cmd\.Dir = "\.\."\n', '', content)

# Replace binary paths in exec.Command
content = re.sub(r'"\.\./bdd/art-dupl-test"', '"./bdd/art-dupl-test"', content)

# Replace os.Remove paths
content = re.sub(r'os\.Remove\("\.\./bdd/art-dupl-test"\)', 'os.Remove("bdd/art-dupl-test")', content)

with open('bdd_test.go', 'w') as f:
    f.write(content)

print("Fixed BDD test paths")
