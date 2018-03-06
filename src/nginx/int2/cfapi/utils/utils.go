package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func IsDir(name string) (bool, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return false, err
	}
	return fi.Mode().IsDir(), nil
}

func Zipit(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func WaitForHttpPort(port string) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			client := http.Client{
				Timeout: time.Duration(1 * time.Second),
			}
			if resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%s/healthcheck", port)); err == nil {
				resp.Body.Close()
				return nil
			}
		case <-time.After(3 * time.Second):
			return fmt.Errorf("Timed out waiting to connect to port %s", port)
		}
	}
}

func HttpGet(url string, headers map[string]string) (string, map[string][]string, error) {
	client := &http.Client{}
	if headers["NoFollow"] == "true" {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		delete(headers, "NoFollow")
	}
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	if headers["user"] != "" && headers["password"] != "" {
		req.SetBasicAuth(headers["user"], headers["password"])
		delete(headers, "user")
		delete(headers, "password")
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", map[string][]string{}, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", map[string][]string{}, err
	}
	resp.Header["StatusCode"] = []string{strconv.Itoa(resp.StatusCode)}
	return string(data), resp.Header, err
}

func HttpGetBody(url string) (string, error) {
	body, _, err := HttpGet(url, map[string]string{})
	// TODO: Non 200 ??
	// if !(len(headers["StatusCode"]) == 1 && headers["StatusCode"][0] == "200") {
	// 	return "", fmt.Errorf("non 200 status: %v", headers)
	// }
	return body, err
}

func ConfirmBuildpack(log, version string) error {
	if version == "" {
		return nil
	}
	matches := regexp.MustCompile(`Buildpack version (\S+)`).FindAllStringSubmatch(log, -1)
	if len(matches) == 0 {
		return fmt.Errorf("Wrong buildpack version. Could not find any buildpack version lines")
	}

	versions := []string{}
	for _, m := range matches {
		versions = append(versions, string(m[1]))
		if string(m[1]) == version {
			return nil
		}
	}

	return fmt.Errorf("Wrong buildpack version. Expected '%s', but these were logged: %s", version, strings.Join(versions, ", "))
}
