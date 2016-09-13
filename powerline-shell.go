// Copyright 2014 Matt Martz <matt@sivel.net>
// All Rights Reserved.
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/otann/powerline-go/powerline"
	"strconv"
)

// Colors
const fg = "12"
const bg = "0"
const separatorColor = "8"

const homeFg = bg
const homeBg = "10"

const cleanFg = bg
const cleanBg = "14"
const dirtyFg = bg
const dirtyBg = "2"

func getCurrentWorkingDir() (string, []string) {
	dir, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}
	userDir := strings.Replace(dir, os.Getenv("HOME"), "~", 1)
	userDir = strings.TrimSuffix(userDir, "/")
	parts := strings.Split(userDir, "/")
	return dir, parts
}

func getVirtualEnv() (string, []string, string) {
	var parts []string
	virtualEnv := os.Getenv("VIRTUAL_ENV")
	if virtualEnv == "" {
		return "", parts, ""
	}

	parts = strings.Split(virtualEnv, "/")

	virtualEnvName := path.Base(virtualEnv)
	return virtualEnv, parts, virtualEnvName
}

func isWritableDir(dir string) bool {
	tmpPath := path.Join(dir, ".powerline-write-test")
	_, err := os.Create(tmpPath)
	if err != nil {
		return false
	}
	os.Remove(tmpPath)
	return true
}

func getGitInformation() (string, bool) {
	var status string
	var staged bool
	stdout, _ := exec.Command("git", "status", "--ignore-submodules").Output()
	reBranch := regexp.MustCompile(`^(HEAD detached at|HEAD detached from|On branch) (\S+)`)
	matchBranch := reBranch.FindStringSubmatch(string(stdout))
	if len(matchBranch) > 0 {
		if matchBranch[2] == "detached" {
			status = matchBranch[2]
		} else {
			status = matchBranch[2]
		}
	}

	reStatus := regexp.MustCompile(`Your branch is (ahead|behind).*?([0-9]+) comm`)
	matchStatus := reStatus.FindStringSubmatch(string(stdout))
	if len(matchStatus) > 0 {
		status = fmt.Sprintf("%s %s", status, matchStatus[2])
		if matchStatus[1] == "behind" {
			status = fmt.Sprintf("%s\u21E3", status)
		} else if matchStatus[1] == "ahead" {
			status = fmt.Sprintf("%s\u21E1", status)
		}
	}

	staged = !strings.Contains(string(stdout), "nothing to commit")
	if strings.Contains(string(stdout), "Untracked files") {
		status = fmt.Sprintf("%s +", status)
	}

	return status, staged
}

func addCwd(cwdParts []string, ellipsis string, separator string) [][]string {
	segments := [][]string{}

	// if it is root, return immediately
	if len(cwdParts) == 1 && cwdParts[0] != "~" {
		return [][]string{[]string{fg, bg, "/"}}
	}

	// deal with first segment
	if cwdParts[0] == "~" {
		// if it was home, then it's a real segment, so we remove it
		segments = append(segments, []string{homeFg, homeBg, "~"})
	} else {
		// if it is not home, then print root
		segments = append(segments, []string{fg, bg, "/", separator, separatorColor})
	}
	cwdParts = cwdParts[1:]

	// add second segment, if it's not last one
	length := len(cwdParts)
	if length >= 4 {
		segments = append(segments, []string{fg, bg, cwdParts[0], separator, separatorColor})
	}

	// ellipsis
	if length == 3 {
		segments = append(segments, []string{fg, bg, cwdParts[0], separator, separatorColor})
	} else if length > 3 {
		segments = append(segments, []string{fg, bg, ellipsis, separator, separatorColor})
	}

	// add but-last segment
	if length > 1 {
		segments = append(segments, []string{fg, bg, cwdParts[length-2], separator, separatorColor})
	}

	// add last segment
	if length > 0 {
		segments = append(segments, []string{fg, bg, cwdParts[length-1]})
	}

	return segments
}

func addVirtulEnvName() []string {
	_, _, virtualEnvName := getVirtualEnv()
	if virtualEnvName != "" {
		return []string{"000", "035", virtualEnvName}
	}

	return nil
}

func addLock(cwd string, lock string) []string {
	if !isWritableDir(cwd) {
		return []string{"254", "124", lock}
	}

	return nil
}

func addGitInfo() []string {
	gitStatus, gitStaged := getGitInformation()
	if gitStatus != "" {
		if gitStaged {
			return []string{dirtyFg, dirtyBg, gitStatus}
		} else {
			return []string{cleanFg, cleanBg, gitStatus}
		}
	} else {
		return nil
	}
}

func addExitCode(code string) []string {
	i, err := strconv.Atoi(code)
	if err != nil || i == 0 {
		return nil
	} else {
		return []string{"15", "1", code}
	}
}

func main() {
	shell := "bash"

	if len(os.Args) > 1 {
		shell = os.Args[1]
	}

	exitCode := "0"
	if len(os.Args) > 2 {
		exitCode = os.Args[2]
	}

	p := powerline.NewPowerline(shell)
	cwd, cwdParts := getCurrentWorkingDir()

	p.AppendSegment(addVirtulEnvName())
	p.AppendSegments(addCwd(cwdParts, p.Ellipsis, p.SeparatorThin))
	p.AppendSegment(addLock(cwd, p.Lock))
	p.AppendSegment(addGitInfo())
	p.AppendSegment(addExitCode(exitCode))

	fmt.Print(p.PrintSegments())
}
