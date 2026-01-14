#!/usr/bin/env python3
import re
import sys

def replace_assertions(content):
    # Replace assert.NoError(t, err)
    content = re.sub(
        r'assert\.NoError\(t, (\w+)\)',
        r'if \1 != nil {\n\t\tt.Errorf("unexpected error: %v", \1)\n\t}',
        content
    )

    # Replace assert.Len(t, x, n)
    content = re.sub(
        r'assert\.Len\(t, (\w+), (\d+)\)',
        r'if len(\1) != \2 {\n\t\tt.Errorf("expected length %d, got %d", \2, len(\1))\n\t}',
        content
    )

    # Replace assert.Contains(t, slice, value)
    content = re.sub(
        r'assert\.Contains\(t, (\w+), ([^\)]+)\)',
        r'if !contains(\1, \2) {\n\t\tt.Errorf("expected to contain %v", \2)\n\t}',
        content
    )

    # Replace assert.True(t, condition, "message")
    content = re.sub(
        r'assert\.True\(t, (\w+), "([^"]+)"\)',
        r'if !\1 {\n\t\tt.Error("\2")\n\t}',
        content
    )

    content = re.sub(
        r'assert\.True\(t, (\w+)\)',
        r'if !\1 {\n\t\tt.Error("expected true")\n\t}',
        content
    )

    # Replace assert.False(t, condition, "message")
    content = re.sub(
        r'assert\.False\(t, (\w+), "([^"]+)"\)',
        r'if \1 {\n\t\tt.Error("\2")\n\t}',
        content
    )

    content = re.sub(
        r'assert\.False\(t, (\w+)\)',
        r'if \1 {\n\t\tt.Error("expected false")\n\t}',
        content
    )

    # Replace assert.Equal(t, expected, actual)
    content = re.sub(
        r'assert\.Equal\(t, (\w+), (\w+)\)',
        r'if \2 != \1 {\n\t\tt.Errorf("expected %v, got %v", \1, \2)\n\t}',
        content
    )

    return content

def add_helper_function(content):
    # Add contains helper function if needed
    if 'assert.Contains(t' in content or 'contains(' not in content:
        # Find first function and add helper before it
        helper = '''
func contains[T comparable](slice []T, item T) bool {
    for _, v := range slice {
        if v == item {
            return true
        }
    }
    return false
}
'''
        return helper + content
    return content

if __name__ == '__main__':
    if len(sys.argv) != 2:
        print("Usage: python3 remove_testify.py <file>")
        sys.exit(1)

    with open(sys.argv[1], 'r') as f:
        content = f.read()

    content = add_helper_function(content)
    content = replace_assertions(content)

    with open(sys.argv[1], 'w') as f:
        f.write(content)

    print(f"Processed {sys.argv[1]}")
