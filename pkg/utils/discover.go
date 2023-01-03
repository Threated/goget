package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
)

func readJsonInto(body io.ReadCloser, out any) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}

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

func DownloadBlob(urlString, apiToken, filePath string) error {
	type blob struct {
		Data []byte `json:"content"`
	}

	req, err := http.NewRequest("GET", urlString, nil)

	// add authorization header to the req
	if apiToken != "" {
		req.Header.Add("Authorization", "Bearer "+apiToken)
	}

	// Send req using http Client
	client := &http.Client{}
	res, err := client.Do(req)
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

func downloadFiles(files []FileInfo, out, baseUrl string, depth int, apiToken string, results chan Result) {
	os.MkdirAll(out, os.ModePerm)
	var wg sync.WaitGroup
	for _, file := range files {
		switch file.Type {
		case File:
			wg.Add(1)
			go func(theFile FileInfo, filePath string) {
				defer wg.Done()
				results <- Result{&theFile, DownloadBlob(theFile.GitUrl, apiToken, filePath)}
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
				downloadFiles(moreFiles, outFilePath, newBaseUrl, depth-1, apiToken, results)
			}(newFiles, path.Join(out, file.Name), newUrl)
		default:
			results <- Result{&file, fmt.Errorf("Unknown response type %s", file.Type)}
		}
	}
	wg.Wait()
}

func Download(subRepo *RepoInfo, outDir string, depth int) chan Result {
	results := make(chan Result)

	go func() {
		baseUrl := subRepo.Url()
        apiToken := subRepo.ApiToken
		if subRepo.UrlType == Blob {
			os.MkdirAll(outDir, os.ModePerm)
			name := subRepo.Path[len(subRepo.Path)-1]
			results <- Result{
				&FileInfo{
					Name:   name,
					Type:   File,
					GitUrl: baseUrl,
				},
				DownloadBlob(baseUrl, apiToken, path.Join(outDir, name)),
			}
		} else if subRepo.UrlType == Tree {
			files, err := DirInfoFromUrl(baseUrl)
			if err != nil {
				results <- Result{nil, err}
			} else {
				downloadFiles(files, outDir, baseUrl, depth, apiToken, results)
			}
		} else {
			results <- Result{
				nil,
				fmt.Errorf("Unknown git url schema expected blob or tree got %s", string(subRepo.UrlType)),
			}
		}
		close(results)
	}()

	return results
}
