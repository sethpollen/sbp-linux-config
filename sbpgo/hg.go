// Utilities for dealing with Mercurial repositories. Mercurial is written in
// Python, and as such it is pretty slow. It takes about 50 ms to run
// "hg status" on a repo. This is too long to run inline in my command prompt.
// So this file contains some hacks to query Mercurial repo information without
// invoking any part of the Mercurial codebase.
package sbpgo

import (
	"errors"
	"os"
	"path"
)

// Encapsulates information about an Hg repo.
type HgInfo struct {
	// Name of this Hg repo.
	RepoName string
	// Full path to the root of this repo.
	RepoPath string
	// Pwd, relative to the root repo path.
	RelativePwd string
	// True if there are uncommitted local changes.
	Dirty bool
}

func GetHgInfo(pwd string) (*HgInfo, error) {
	repoPath, err := getHgRepoRoot(pwd)
	if err != nil {
		return nil, err
	}

	// Now that we know we are in an Hg repo, it's worth paying the cost to run
	// hg status.
	status, err := EvalCommandSync(pwd, "hg", "status")
	if err != nil {
		return nil, err
	}

	var info = new(HgInfo)
	info.RepoName = path.Base(repoPath)
	info.RepoPath = repoPath
	info.RelativePwd = RelativePath(pwd, repoPath)
	info.Dirty = (status != "")
	return info, nil
}

func getHgRepoRoot(pwd string) (string, error) {
	repoPath, err := SearchParents(pwd, isHgRepoRoot)
	if err != nil {
		return "", errors.New("Not in an Hg repo")
	}
	return repoPath, nil
}

func isHgRepoRoot(pwd string) bool {
	var metaDir = path.Join(pwd, ".hg")
	fileInfo, err := os.Stat(metaDir)
	return err == nil && fileInfo.IsDir()
}

// A Module that matches any directory inside an Hg repo.
type hgModule struct {
	result chan *HgInfo
	err    chan error
}

func (self *hgModule) Prepare(env *PromptEnv) {
	go func() {
		result, err := GetHgInfo(env.Pwd)
		if err != nil {
			self.err <- err
		} else {
			self.result <- result
		}
	}()
}

func (self *hgModule) Match(env *PromptEnv, updateCache bool) bool {
	select {
	case <-self.err:
		return false
	case hgInfo := <-self.result:
		env.Info = hgInfo.RepoName
		if hgInfo.Dirty {
			env.Info += " *"
		}
		env.Flag = append(env.Flag, Stylize("hg", Magenta, nil)...)
		env.Pwd = hgInfo.RelativePwd
		return true
	}
}

func (self *hgModule) Description() string {
	return "hg"
}

func HgModule() *hgModule {
	return &hgModule{make(chan *HgInfo), make(chan error)}
}
