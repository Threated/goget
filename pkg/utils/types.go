package utils

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type GitUrlType string

const (
	Blob GitUrlType = "blob"
	Tree            = "tree"
)

type GitResType string

const (
	File GitResType = "file"
	Dir             = "dir"
)

type RepoInfo struct {
	User    string
	Repo    string
	UrlType GitUrlType
	Branch  string
	Path    []string
}

type Result struct {
	Context *FileInfo
	Err     error
}


func NewRepoInfoFromUrl(urlString string) (*RepoInfo, error) {
	urlInfo, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	path := strings.Split(urlInfo.Path[1:], "/")
	if len(path) <= 4 {
		return nil, errors.New("Url must contain user, reponame, branch and a subfile or subfolder name.")
	}
	return &RepoInfo{
		User:    path[0],
		Repo:    path[1],
		UrlType: GitUrlType(path[2]),
		Branch:  path[3],
		Path:    path[4:],
	}, nil
}

func (repo *RepoInfo) String() string {
	return fmt.Sprintf("RepoInfo { User: %s, Repo: %s, UrlType: %s, Branch: %s, Path: %s}",
		repo.User,
		repo.Repo,
		repo.UrlType,
		repo.Branch,
		repo.Path,
	)
}

func (repo *RepoInfo) Url() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", repo.User, repo.Repo, strings.Join(repo.Path, "/"))
}

type FileInfo struct {
	Type   GitResType `json:"type"`
	GitUrl string     `json:"git_url"`
	Name   string     `json:"name"`
}

func (file *FileInfo) String() string {
    return fmt.Sprintf("%s: %s from GitUrl: %s", string(file.Type), file.Name, file.GitUrl)
}

