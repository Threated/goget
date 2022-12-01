package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
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

func readJsonInto(body io.ReadCloser, out any) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	// println(string(data))
	return json.Unmarshal(data, out)
}

func DirInfoFromUrl(urlString string) ([]FileInfo, error) {

	res, err := http.Get(urlString)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Got status %d", res.StatusCode)
	}

	defer res.Body.Close()
	var resJson []FileInfo
	if err = readJsonInto(res.Body, &resJson); err != nil {
		return nil, err
	}

	return resJson, nil
}

type blob struct {
	Data []byte `json:"content"`
}

func DownloadBlob(urlString string, filePath string) error {
	res, err := http.Get(urlString)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("Got status %d", res.StatusCode)
	}

	defer res.Body.Close()
	var fileBlob blob
	if err = readJsonInto(res.Body, &fileBlob); err != nil {
		return err
	}
	if file, err := os.Create(filePath); err != nil {
		return err
	} else {
		_, err = io.Copy(file, bytes.NewReader(fileBlob.Data))
		return err
	}
}

type Result struct {
	Context *FileInfo
	Err     error
}

func DownloadFiles(files []FileInfo, out, baseUrl string, depth int, results chan Result) {
    os.MkdirAll(out, os.ModePerm)
	var wg sync.WaitGroup
	for _, file := range files {
		switch file.Type {
		case File:
			wg.Add(1)
			go func(theFile FileInfo, filePath string) {
				defer wg.Done()
				results <- Result{&theFile, DownloadBlob(theFile.GitUrl, filePath)}
			}(file, path.Join(out, file.Name))
		case Dir:
            if depth == 0 {
                continue
            }
			newUrl, err := url.JoinPath(baseUrl, file.Name)
			if err != nil {
				results <- Result{&file, err}
				continue
			}
			newFiles, err := DirInfoFromUrl(newUrl)
			if err != nil {
				results <- Result{&file, err}
				continue
			}
			wg.Add(1)
			go func(moreFiles []FileInfo, outFilePath, newBaseUrl string) {
				defer wg.Done()
				DownloadFiles(moreFiles, outFilePath, newBaseUrl, depth - 1, results)
			}(newFiles, path.Join(out, file.Name), newUrl)
		default:
			results <- Result{&file, fmt.Errorf("Unknown response type %s", file.Type)}
		}
	}
	wg.Wait()
}

func Download(subRepo *RepoInfo, outDir string, depth int, results chan Result) {
	baseUrl := subRepo.Url()

	if subRepo.UrlType == Blob {
        name := subRepo.Path[len(subRepo.Path)-1]
		results <- Result {
            &FileInfo {
                Name: name,
                Type: File,
                GitUrl: baseUrl,
            },
            DownloadBlob(baseUrl, path.Join(outDir, name)),
        }
	} else if subRepo.UrlType == Tree {
		files, err := DirInfoFromUrl(baseUrl)
		if err != nil {
			results <- Result{nil, err}
		} else {
			DownloadFiles(files, outDir, baseUrl, depth, results)
		}
	} else {
		results <- Result {
            nil,
            fmt.Errorf("Unknown git url schema expected blob or tree got %s", string(subRepo.UrlType)),
        }
	}
	close(results)
}
