package snaps

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"regexp"
	"strings"
)

var skippedTests = newSyncSlice()

// Wrapper of testing.Skip
//
// Keeps track which snapshots are getting skipped and not marked as obsolete.
func Skip(t testingT, args ...interface{}) {
	t.Helper()

	skippedTests.append(t.Name())
	t.Skip(args...)
}

// Wrapper of testing.Skipf
//
// Keeps track which snapshots are getting skipped and not marked as obsolete.
func Skipf(t testingT, format string, args ...interface{}) {
	t.Helper()

	skippedTests.append(t.Name())
	t.Skipf(format, args...)
}

// Wrapper of testing.SkipNow
//
// Keeps track which snapshots are getting skipped and not marked as obsolete.
func SkipNow(t testingT) {
	t.Helper()

	skippedTests.append(t.Name())
	t.SkipNow()
}

/*
This checks if the parent test is skipped,
or provided a 'runOnly' the testID is part of it

e.g

	func TestParallel (t *testing.T) {
		snaps.Skip(t)
		...
	}

Then every "child" test should be skipped
*/
func testSkipped(testID, runOnly string) bool {
	matched, _ := regexp.Match(runOnly, []byte(testID))

	if runOnly != "" && !matched {
		return true
	}

	// testID form: Test.*/runName - 1
	testName := strings.Split(testID, " - ")[0]

	for _, name := range skippedTests.values {
		if testName == name || strings.HasPrefix(testName, name+"/") {
			return true
		}
	}

	return false
}

func isFileSkipped(dir, filename, runOnly string) bool {
	// When a file is skipped through CLI with -run flag we can track it
	if runOnly == "" {
		return false
	}

	testFilePath := path.Join(dir, "..", strings.TrimSuffix(filename, snapsExt)+".go")
	isSkipped := true

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFilePath, nil, parser.ParseComments)
	if err != nil {
		return false
	}

	for _, decls := range file.Decls {
		funcDecl, ok := decls.(*ast.FuncDecl)

		if !ok {
			continue
		}

		// If the TestFunction is inside the file then it's not skipped
		if funcDecl.Name.String() == runOnly {
			isSkipped = false
		}
	}

	return isSkipped
}
