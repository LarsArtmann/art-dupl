package bdd

const (
	goldenFile1      = "file1.go"
	goldenFile2      = "file2.go"
	goldenTestFile1  = "test1.go"
	goldenSmallFile1 = "small1.go"
	goldenLargeFile1 = "large1.go"
	largeFile2       = "large2.go"
	regularFile1     = "regular1.go"
	regularFile2     = "regular2.go"
	smallFile2       = "small2.go"
	testFile2        = "test2.go"
	testThreshold50  = "15"
	testThreshold15  = "2"
	testThreshold10  = "1"
	flagKeyThreshold = "threshold"
	serviceFile1     = "service1.go"
	serviceFile2     = "service2.go"
)

// dupFuncSource returns a multi-statement function source that the actionability
// filter does NOT suppress. Trivial single-statement bodies such as
// `func f() { println(1) }` are classified as non-actionable boilerplate and
// dropped from semantic output, which silently breaks BDD fixtures that only
// need a reliably-detected clone. Keep the body at >=3 distinct statements.
func dupFuncSource(name string) string {
	return "func " + name + "() {\n" +
		"\tx := 1\n" +
		"\ty := x + 2\n" +
		"\tprintln(y)\n" +
		"}"
}
