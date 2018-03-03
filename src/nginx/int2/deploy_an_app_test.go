package integration_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

type App struct {
	t       *testing.T
	fixture string
	name    string
	tmpPath string
	port    string
	Stdout  bytes.Buffer
	Stderr  bytes.Buffer
}

func (a *App) Run() error {
	cmd := exec.Command("cf", "local", "run", a.name)
	cmd.Dir = a.tmpPath
	var stdout bytes.Buffer
	cmd.Stdout = io.MultiWriter(&a.Stdout, &stdout)
	cmd.Stderr = &a.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
DONE:
	for {
		select {
		case <-ticker.C:
			if line, err := a.Stdout.ReadString('\n'); err == nil {
				port := regexp.MustCompile(`Running .* on port (\d+)\.\.\.`).FindAllStringSubmatch(line, 1)
				if len(port) == 1 {
					a.port = port[0][1]
					break DONE
				} else {
					return fmt.Errorf("Could not find port in: %s", line)
				}
			}
		case <-time.After(3 * time.Second):
			return fmt.Errorf("Timed out trying to find port")
		}
	}

	for {
		select {
		case <-ticker.C:
			conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%s", a.port))
			if err == nil {
				conn.Close()
				return nil
			}
		case <-time.After(3 * time.Second):
			return fmt.Errorf("Timed out waiting to connect to port %s", a.port)
		}
	}
}

func (a *App) GetUrl(path string) (string, error) {
	if a.port == "" {
		return "", fmt.Errorf("app does not have a known running port")
	}
	return fmt.Sprintf("http://localhost:%s%s", a.port, path), nil
}

func (a *App) Get(path string, headers map[string]string) (string, map[string][]string, error) {
	url, err := a.GetUrl(path)
	if err != nil {
		return "", map[string][]string{}, err
	}
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

func (a *App) GetBody(path string) (string, error) {
	body, _, err := a.Get(path, map[string]string{})
	// TODO: Non 200 ??
	// if !(len(headers["StatusCode"]) == 1 && headers["StatusCode"][0] == "200") {
	// 	return "", fmt.Errorf("non 200 status: %v", headers)
	// }
	return body, err
}

func testObject3(t *testing.T, when spec.G, it spec.S) {
	var app *App

	it.Before(func() {
		RegisterTestingT(t)

		app = &App{
			t:       t,
			fixture: "~/workspace/nginx-buildpack/fixtures/mainline/",
			name:    "app1",
			tmpPath: "/tmp",
		}
	})

	it("do thing", func() {
		Expect(app.Run()).To(Succeed())
		Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
	})
}

// package integration_test
//
// import (
// 	"path/filepath"
//
// 	"github.com/cloudfoundry/libbuildpack/cutlass"
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// )
//
// var _ = Describe("CF Nginx Buildpack", func() {
// 	var app *cutlass.App
//
// 	AfterEach(func() {
// 		if app != nil {
// 			app.Destroy()
// 		}
// 		app = nil
// 	})
//
// 	Context("with no specified version", func() {
// 		BeforeEach(func() {
// 			app = cutlass.New(filepath.Join(bpDir, "fixtures", "unspecified_version"))
// 		})
//
// 		It("Uses latest mainline nginx", func() {
// 			PushAppAndConfirm(app)
//
// 			Eventually(app.Stdout.String).Should(ContainSubstring(`No nginx version specified - using mainline => 1.13.`))
// 			Eventually(app.Stdout.String).ShouldNot(ContainSubstring(`Requested nginx version:`))
//
// 			Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
// 			Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
// 		})
// 	})
//
// 	Context("with an nginx app specifying mainline", func() {
// 		BeforeEach(func() {
// 			app = cutlass.New(filepath.Join(bpDir, "fixtures", "mainline"))
// 		})
//
// 		It("Logs nginx buildpack version", func() {
// 			PushAppAndConfirm(app)
//
// 			Eventually(app.Stdout.String).Should(ContainSubstring(`Requested nginx version: mainline => 1.13.`))
//
// 			Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
// 			Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
// 		})
// 	})
//
// 	Context("with an nginx app specifying stable", func() {
// 		BeforeEach(func() {
// 			app = cutlass.New(filepath.Join(bpDir, "fixtures", "stable"))
// 		})
//
// 		It("Logs nginx buildpack version", func() {
// 			PushAppAndConfirm(app)
//
// 			Eventually(app.Stdout.String).Should(ContainSubstring(`Requested nginx version: stable => 1.12.`))
// 			Eventually(app.Stdout.String).Should(ContainSubstring(`Warning: usage of "stable" versions of NGINX is discouraged in most cases by the NGINX team.`))
//
// 			Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
// 			Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
// 		})
// 	})
//
// 	Context("with an nginx app specifying 1.12.x", func() {
// 		BeforeEach(func() {
// 			app = cutlass.New(filepath.Join(bpDir, "fixtures", "1_12_x"))
// 		})
//
// 		It("Logs nginx buildpack version", func() {
// 			PushAppAndConfirm(app)
//
// 			Eventually(app.Stdout.String).Should(ContainSubstring(`Requested nginx version: 1.12.x => 1.12.`))
// 			Eventually(app.Stdout.String).Should(ContainSubstring(`Warning: usage of "stable" versions of NGINX is discouraged in most cases by the NGINX team.`))
//
// 			Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
// 			Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
// 		})
// 	})
//
// 	Context("with an nginx app specifying an unknown version", func() {
// 		BeforeEach(func() {
// 			app = cutlass.New(filepath.Join(bpDir, "fixtures", "unavailable_version"))
// 		})
//
// 		It("Logs nginx buildpack version", func() {
// 			Expect(app.Push()).ToNot(Succeed())
//
// 			Eventually(app.Stdout.String).Should(ContainSubstring(`Available versions: mainline, stable, 1.12.x, 1.13.x`))
// 		})
// 	})
// })
