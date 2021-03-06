package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Sirupsen/logrus"
)

// Commits returns a set of commits.
// If commitrange is a git still range 12345...54321, then it will be isolated set of commits.
// If commitrange is a single commit, all ancestor commits up through the hash provided.
func Commits(commitrange string) ([]CommitEntry, error) {
	cmdArgs := []string{"git", "--no-pager", "log", `--pretty=format:%H`, commitrange}
	if debug() {
		logrus.Infof("[git] cmd: %q", strings.Join(cmdArgs, " "))
	}
	output, err := exec.Command(cmdArgs[0], cmdArgs[1:]...).Output()
	if err != nil {
		logrus.Errorf("mm[git] cmd: %q", strings.Join(cmdArgs, " "))
		return nil, err
	}
	commitHashes := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]CommitEntry, len(commitHashes))
	for i, commitHash := range commitHashes {
		c, err := LogCommit(commitHash)
		if err != nil {
			return commits, err
		}
		commits[i] = *c
	}
	return commits, nil
}

// FieldNames are for the formating and rendering of the CommitEntry structs.
// Keys here are from git log pretty format "format:..."
var FieldNames = map[string]string{
	"%h":  "abbreviated_commit",
	"%p":  "abbreviated_parent",
	"%t":  "abbreviated_tree",
	"%aD": "author_date",
	"%aE": "author_email",
	"%aN": "author_name",
	"%b":  "body",
	"%H":  "commit",
	"%N":  "commit_notes",
	"%cD": "committer_date",
	"%cE": "committer_email",
	"%cN": "committer_name",
	"%e":  "encoding",
	"%P":  "parent",
	"%D":  "refs",
	"%f":  "sanitized_subject_line",
	"%GS": "signer",
	"%GK": "signer_key",
	"%s":  "subject",
	"%G?": "verification_flag",
}

// Check warns if changes introduce whitespace errors.
// Returns non-zero if any issues are found.
func Check(commit string) ([]byte, error) {
	args := []string{
		"--no-pager", "log", "--check",
		fmt.Sprintf("%s^..%s", commit, commit),
	}
	if exclude := os.Getenv("GIT_CHECK_EXCLUDE"); exclude != "" {
		args = append(args, "--", ".", fmt.Sprintf(":(exclude)%s", exclude))
	}
	cmd := exec.Command("git", args...)
	if debug() {
		logrus.Infof("[git] cmd: %q", strings.Join(cmd.Args, " "))
	}
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

// Show returns the diff of a commit.
//
// NOTE: This could be expensive for very large commits.
func Show(commit string) ([]byte, error) {
	cmd := exec.Command("git", "--no-pager", "show", commit)
	if debug() {
		logrus.Infof("[git] cmd: %q", strings.Join(cmd.Args, " "))
	}
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

// CommitEntry represents a single commit's information from `git`.
// See also FieldNames
type CommitEntry map[string]string

// LogCommit assembles the full information on a commit from its commit hash
func LogCommit(commit string) (*CommitEntry, error) {
	c := CommitEntry{}
	for k, v := range FieldNames {
		cmd := exec.Command("git", "--no-pager", "log", "-1", `--pretty=format:`+k+``, commit)
		if debug() {
			logrus.Infof("[git] cmd: %q", strings.Join(cmd.Args, " "))
		}
		cmd.Stderr = os.Stderr
		out, err := cmd.Output()
		if err != nil {
			logrus.Errorf("[git] cmd: %q", strings.Join(cmd.Args, " "))
			return nil, err
		}
		c[v] = strings.TrimSpace(string(out))
	}

	return &c, nil
}

func debug() bool {
	return len(os.Getenv("DEBUG")) > 0
}

// FetchHeadCommit returns the hash of FETCH_HEAD
func FetchHeadCommit() (string, error) {
	cmdArgs := []string{"git", "--no-pager", "rev-parse", "--verify", "FETCH_HEAD"}
	if debug() {
		logrus.Infof("[git] cmd: %q", strings.Join(cmdArgs, " "))
	}
	output, err := exec.Command(cmdArgs[0], cmdArgs[1:]...).Output()
	if err != nil {
		logrus.Errorf("[git] cmd: %q", strings.Join(cmdArgs, " "))
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// HeadCommit returns the hash of HEAD
func HeadCommit() (string, error) {
	cmdArgs := []string{"git", "--no-pager", "rev-parse", "--verify", "HEAD"}
	if debug() {
		logrus.Infof("[git] cmd: %q", strings.Join(cmdArgs, " "))
	}
	output, err := exec.Command(cmdArgs[0], cmdArgs[1:]...).Output()
	if err != nil {
		logrus.Errorf("[git] cmd: %q", strings.Join(cmdArgs, " "))
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
