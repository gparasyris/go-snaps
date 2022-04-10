package snaps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// MatchSnapshot verifies the values match the most recent snap file**
// You can pass multiple parameters
// 	MatchSnapshot(t, 10, "hello world")
// or call MatchSnapshot multiples times
//  MatchSnapshot(t, 10)
//  MatchSnapshot(t, "hello world")
// The difference is the latter will create multiple entries
func MatchSnapshot(t *testing.T, o ...interface{}) {
	t.Helper()

	matchSnapshot(t, o)
}

func matchSnapshot(t testingT, o []interface{}) {
	t.Helper()

	if len(o) == 0 {
		t.Log(yellowText("[warning] MatchSnapshot call without params\n"))
		return
	}

	dir, snapPath := snapDirAndName()
	testID := testsRegistry.getTestID(t.Name(), snapPath)
	snapshot := takeSnapshot(o)
	prevSnapshot, err := getPrevSnapshot(testID, snapPath)

	if errors.Is(err, errSnapNotFound) {
		if isCI {
			t.Error(err)
			return
		}

		err := addNewSnapshot(testID, snapshot, dir, snapPath)
		if err != nil {
			t.Error(err)
		}

		t.Log(greenText(arrowPoint + "New snapshot written.\n"))
		return
	}
	if err != nil {
		t.Error(err)
	}

	diff := prettyDiff(unEscapeEndChars(prevSnapshot), unEscapeEndChars(snapshot))
	if diff != "" {
		if shouldUpdate {
			err := updateSnapshot(testID, snapshot, snapPath)
			if err != nil {
				t.Error(err)
			}

			t.Log(greenText(arrowPoint + "Snapshot updated.\n"))
			return
		}

		t.Error(diff)
	}
}

func getPrevSnapshot(testID, snapPath string) (string, error) {
	f, err := snapshotFileToString(snapPath)
	if err != nil {
		return "", err
	}

	match := dynamicTestIDRegexp(testID).FindStringSubmatch(f)

	if len(match) < 2 {
		return "", errSnapNotFound
	}

	// The second capture group contains the snapshot data
	return match[1], err
}

func snapshotFileToString(name string) (string, error) {
	_, err := os.Stat(name)
	if err != nil {
		return "", errSnapNotFound
	}

	f, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}

	return string(f), err
}

func stringToSnapshotFile(snap, name string) error {
	return os.WriteFile(name, []byte(snap), os.ModePerm)
}

func addNewSnapshot(testID, snapshot, dir, snapPath string) error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(snapPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("\n%s\n%s---\n", testID, snapshot))
	if err != nil {
		return err
	}

	return nil
}

/*
	Returns the dir for snapshots [where the tests run + /snapsDirName]
	and the name [dir + /snapsDirName + /<test-name>.snapsExtName]
*/
func snapDirAndName() (dir, name string) {
	callerPath, _ := baseCaller()
	base := filepath.Base(callerPath)

	dir = filepath.Join(filepath.Dir(callerPath), snapsDir)
	name = filepath.Join(dir, strings.TrimSuffix(base, filepath.Ext(base))+snapsExt)

	return
}

func updateSnapshot(testID, snapshot, snapPath string) error {
	// When t.Parallel a test can override another snapshot as we dump
	// all snapshots
	_m.Lock()
	defer _m.Unlock()
	f, err := snapshotFileToString(snapPath)
	if err != nil {
		return err
	}

	updatedSnapFile := dynamicTestIDRegexp(testID).
		ReplaceAllLiteralString(f, fmt.Sprintf("%s\n%s---", testID, snapshot))

	err = stringToSnapshotFile(updatedSnapFile, snapPath)
	if err != nil {
		return err
	}

	return nil
}
