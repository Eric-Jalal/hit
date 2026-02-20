package ci

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/elisa-content-delivery/hit/internal/github"
)

var (
	ghAnnotationRe = regexp.MustCompile(`::error\s+file=([^,]+),line=(\d+)(?:,col=(\d+))?::(.+)`)
	goErrorRe      = regexp.MustCompile(`^([^\s]+\.go):(\d+):(\d+):\s+(.+)`)
	goTestFailRe   = regexp.MustCompile(`^--- FAIL:\s+(\S+)`)
)

func ParseAnnotations(log string) []github.ErrorAnnotation {
	var annotations []github.ErrorAnnotation

	for _, line := range strings.Split(log, "\n") {
		line = strings.TrimSpace(line)

		if matches := ghAnnotationRe.FindStringSubmatch(line); matches != nil {
			lineNum, _ := strconv.Atoi(matches[2])
			col := 0
			if matches[3] != "" {
				col, _ = strconv.Atoi(matches[3])
			}
			annotations = append(annotations, github.ErrorAnnotation{
				File:    matches[1],
				Line:    lineNum,
				Column:  col,
				Message: matches[4],
			})
			continue
		}

		if matches := goErrorRe.FindStringSubmatch(line); matches != nil {
			lineNum, _ := strconv.Atoi(matches[2])
			col, _ := strconv.Atoi(matches[3])
			annotations = append(annotations, github.ErrorAnnotation{
				File:    matches[1],
				Line:    lineNum,
				Column:  col,
				Message: matches[4],
			})
			continue
		}

		if matches := goTestFailRe.FindStringSubmatch(line); matches != nil {
			annotations = append(annotations, github.ErrorAnnotation{
				Message: "FAIL: " + matches[1],
			})
		}
	}

	return annotations
}

func IsErrorLine(line string) bool {
	line = strings.TrimSpace(line)
	if ghAnnotationRe.MatchString(line) {
		return true
	}
	if goErrorRe.MatchString(line) {
		return true
	}
	if goTestFailRe.MatchString(line) {
		return true
	}
	if strings.Contains(line, "FAIL") || strings.Contains(line, "ERROR") {
		return true
	}
	return false
}
