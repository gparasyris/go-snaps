package snaps

import (
	"bytes"

	"github.com/sergi/go-diff/diffmatchpatch"
)

var dmp = diffmatchpatch.New()

func prettyDiff(expected, received string) string {
	diffs := dmp.DiffMain(expected, received, false)
	buff := bytes.Buffer{}

	if len(diffs) == 1 && diffs[0].Type == 0 {
		return ""
	}

	buff.WriteString("\n")
	buff.WriteString(redBG(" Snapshot "))
	buff.WriteString("\n")
	buff.WriteString(greenBG(" Received "))
	buff.WriteString("\n\n")

	for _, diff := range diffs {
		switch diff.Type {
		case 0:
			buff.WriteString(dimText(diff.Text))
		case -1:
			buff.WriteString(redText(diff.Text))
		case 1:
			buff.WriteString(greenText(diff.Text))
		}
	}

	buff.WriteString("\n")
	return buff.String()
}
