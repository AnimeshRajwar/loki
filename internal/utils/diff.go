package utils

import (
	"fmt"
	"strings"
)

type DiffOp int

const (
	DiffKeep DiffOp = iota
	DiffDelete
	DiffInsert
)

type DiffLine struct {
	Op   DiffOp
	Text string
}

// MyersDiff implements Myers' $O((N+M)D)$ diff algorithm.
func MyersDiff(a, b []string) []DiffLine {
	n := len(a)
	m := len(b)
	if n == 0 && m == 0 {
		return nil
	}

	if n == 0 {
		res := make([]DiffLine, m)
		for i, line := range b {
			res[i] = DiffLine{Op: DiffInsert, Text: line}
		}
		return res
	}
	if m == 0 {
		res := make([]DiffLine, n)
		for i, line := range a {
			res[i] = DiffLine{Op: DiffDelete, Text: line}
		}
		return res
	}

	maxD := n + m
	v := make([]int, 2*maxD+1)
	history := make([][]int, 0, maxD)

	v[maxD+1] = 0

	found := false
	var d int
	for d = 0; d <= maxD; d++ {
		vCopy := make([]int, len(v))
		copy(vCopy, v)
		history = append(history, vCopy)

		for k := -d; k <= d; k += 2 {
			idx := k + maxD
			var x int
			if k == -d || (k != d && v[idx-1] < v[idx+1]) {
				x = v[idx+1] // move down (insertion)
			} else {
				x = v[idx-1] + 1 // move right (deletion)
			}
			y := x - k

			for x < n && y < m && a[x] == b[y] {
				x++
				y++
			}
			v[idx] = x

			if x >= n && y >= m {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	var path []DiffLine
	x := n
	y := m

	for d > 0 {
		k := x - y
		idx := k + maxD
		vPrev := history[d]

		var prevK int
		if k == -d || (k != d && vPrev[idx-1] < vPrev[idx+1]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}

		prevX := vPrev[prevK+maxD]
		prevY := prevX - prevK

		for x > prevX && y > prevY {
			path = append(path, DiffLine{Op: DiffKeep, Text: a[x-1]})
			x--
			y--
		}

		if x > prevX {
			path = append(path, DiffLine{Op: DiffDelete, Text: a[x-1]})
			x--
		} else if y > prevY {
			path = append(path, DiffLine{Op: DiffInsert, Text: b[y-1]})
			y--
		}
		d--
	}

	for x > 0 && y > 0 {
		path = append(path, DiffLine{Op: DiffKeep, Text: a[x-1]})
		x--
		y--
	}
	for x > 0 {
		path = append(path, DiffLine{Op: DiffDelete, Text: a[x-1]})
		x--
	}
	for y > 0 {
		path = append(path, DiffLine{Op: DiffInsert, Text: b[y-1]})
		y--
	}

	// Reverse the path to get chronological order
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}

type Hunk struct {
	startA, lenA int
	startB, lenB int
	lines        []DiffLine
}

// FormatDiff groups the DiffLines into hunks and formats a git-like colored diff.
func FormatDiff(a, b []string, filepath string) string {
	diffLines := MyersDiff(a, b)
	if len(diffLines) == 0 {
		return ""
	}

	hasChanges := false
	for _, dl := range diffLines {
		if dl.Op != DiffKeep {
			hasChanges = true
			break
		}
	}
	if !hasChanges {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("diff --loki %s\n", filepath))
	sb.WriteString(fmt.Sprintf("--- a/%s\n", filepath))
	sb.WriteString(fmt.Sprintf("+++ b/%s\n", filepath))

	const contextSize = 3
	var hunks []Hunk

	n := len(diffLines)
	i := 0

	type lineInfo struct {
		dl    DiffLine
		lineA int
		lineB int
	}

	infos := make([]lineInfo, n)
	curA, curB := 1, 1
	for idx, dl := range diffLines {
		info := lineInfo{dl: dl}
		if dl.Op == DiffKeep {
			info.lineA = curA
			info.lineB = curB
			curA++
			curB++
		} else if dl.Op == DiffDelete {
			info.lineA = curA
			curA++
		} else if dl.Op == DiffInsert {
			info.lineB = curB
			curB++
		}
		infos[idx] = info
	}

	for i < n {
		if infos[i].dl.Op == DiffKeep {
			i++
			continue
		}

		startIdx := i - contextSize
		if startIdx < 0 {
			startIdx = 0
		}

		lastChangeIdx := i
		for j := i + 1; j < n; j++ {
			if infos[j].dl.Op != DiffKeep {
				lastChangeIdx = j
			} else if j-lastChangeIdx > 2*contextSize {
				break
			}
		}

		hunkEndIdx := lastChangeIdx + contextSize
		if hunkEndIdx >= n {
			hunkEndIdx = n - 1
		}

		hunkLines := infos[startIdx : hunkEndIdx+1]

		hunkStartA := 0
		hunkLenA := 0
		hunkStartB := 0
		hunkLenB := 0

		for _, info := range hunkLines {
			if info.lineA > 0 && hunkStartA == 0 {
				hunkStartA = info.lineA
			}
			if info.lineB > 0 && hunkStartB == 0 {
				hunkStartB = info.lineB
			}
			if info.dl.Op == DiffKeep {
				hunkLenA++
				hunkLenB++
			} else if info.dl.Op == DiffDelete {
				hunkLenA++
			} else if info.dl.Op == DiffInsert {
				hunkLenB++
			}
		}

		if hunkStartA == 0 {
			hunkStartA = curA
		}
		if hunkStartB == 0 {
			hunkStartB = curB
		}

		var hLines []DiffLine
		for _, info := range hunkLines {
			hLines = append(hLines, info.dl)
		}

		hunks = append(hunks, Hunk{
			startA: hunkStartA,
			lenA:   hunkLenA,
			startB: hunkStartB,
			lenB:   hunkLenB,
			lines:  hLines,
		})

		i = hunkEndIdx + 1
	}

	for _, h := range hunks {
		sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", h.startA, h.lenA, h.startB, h.lenB))
		for _, dl := range h.lines {
			switch dl.Op {
			case DiffKeep:
				sb.WriteString(" " + dl.Text + "\n")
			case DiffDelete:
				sb.WriteString(ColorText("-"+dl.Text, "error") + "\n")
			case DiffInsert:
				sb.WriteString(ColorText("+"+dl.Text, "success") + "\n")
			}
		}
	}

	return sb.String()
}
