package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

const endpoint = "unix:///var/run/docker.sock"

type event struct {
	Events []struct {
		Action string `json:"action"`
		Target struct {
			Repository string `json:"repository"`
			Tag        string `json:"tag"`
		} `json:"target"`
	} `json:"events"`
}

func restartDockerWithNewImage(ev *event) {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		log.Printf("client init fail when restart container:%+v", ev)
		return
	}
	Repository := ev.Events[0].Target.Repository
	containerName := "sirdm_" + Repository
	remoteRepo := "127.0.0.1:5000/" + Repository

	time.Sleep(1 * time.Second) //wait a little while since we just receive the notify

	runningImage := ""
	if container, errInspectContainer := client.InspectContainer(containerName); errInspectContainer == nil {
		runningImage = container.Image
	}

	err = client.PullImage(docker.PullImageOptions{
		Repository: remoteRepo,
	}, docker.AuthConfiguration{})
	if err != nil {
		log.Printf("%+v", err)
		return
	}

	time.Sleep(1 * time.Second) //wait a little while since we just finish pull

	if image, errInspectImage := client.InspectImage(remoteRepo); errInspectImage == nil {
		if image.ID == runningImage {
			log.Printf("same image , no need restart")
			return
		}
		log.Printf("%s != %s, need restart ...", runningImage, image.ID)
	}

	client.StopContainer(containerName, 10)
	client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            containerName,
		RemoveVolumes: true,
	})
	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: containerName,
		Config: &docker.Config{
			Cmd:   []string{},
			Image: remoteRepo,
		},
		HostConfig: &docker.HostConfig{},
	})
	client.StartContainer(container.ID, nil)
	log.Printf("new version of %s restarted!", Repository)
}

func handleRegistryEvent(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("can't read body")
		return
	}
	ev := event{}
	json.Unmarshal(body, &ev)
	if ev.Events[0].Action == "push" && ev.Events[0].Target.Tag == "latest" {
		log.Printf("%s %s %s, may need restart ...", ev.Events[0].Action, ev.Events[0].Target.Repository, ev.Events[0].Target.Tag)
		go restartDockerWithNewImage(&ev)
	}
}

func webTask(port int) {
	http.HandleFunc("/event", handleRegistryEvent)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func removeRegister(client *docker.Client, container *docker.Container) {
	log.Printf("wait do cleanup ...")
	client.StopContainer(container.ID, 1000)
	client.RemoveContainer(docker.RemoveContainerOptions{
		ID: container.ID,
	})
}

func getLocalIP() (ip string) {
	ifaces, _ := net.Interfaces()
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		if i.Name != "en0" {
			continue
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP.String()
			case *net.IPAddr:
				ip = v.IP.String()
			}
		}
	}
	return
}

func createCfgFile(port int) string {
	tmpfile, error := ioutil.TempFile("/tmp", "sirdm_")
	defer tmpfile.Close()
	if error != nil {
		log.Fatal("创建文件失败")
	}
	localip := getLocalIP()
	tmpfile.WriteString(fmt.Sprintf(`version: 0.1
log:
  fields:
    service: registry
storage:
  cache:
    blobdescriptor: inmemory
  filesystem:
    rootdirectory: /var/lib/registry
http:
  addr: :5000
  headers:
    X-Content-Type-Options: [nosniff]
health:
  storagedriver:
    enabled: true
    interval: 10s
    threshold: 3
notifications:
  endpoints:
    - name: monitor
      url: http://%s:%d/event
      headers:
        Authorization: [Bearer <your token, if needed>]
      timeout: 500ms
      threshold: 5
      backoff: 1s
`, localip, port))
	tmpfile.Close()
	return tmpfile.Name()
}

func startRegistry(port, webport int) (*docker.Client, *docker.Container, error) {
	log.Printf("Waiting container creating ...")
	cfgfile := createCfgFile(webport)
	log.Printf("store cfg file here: %s", cfgfile)
	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}
	if container, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: "/sirdm_registry",
		Config: &docker.Config{
			Cmd:   []string{"/etc/docker/registry/config.yml"},
			Image: "registry",
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{
				cfgfile + ":/etc/docker/registry/config.yml",
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"5000/tcp": []docker.PortBinding{
					docker.PortBinding{
						HostIP:   "",
						HostPort: fmt.Sprintf("%d", port),
					},
				},
			},
		},
	}); err != nil {
		log.Fatal("err start registry", err)
	} else {
		client.StartContainer(container.ID, nil)
		return client, container, nil
	}
	return nil, nil, fmt.Errorf("check logger")
}

func main() {
	var rport = flag.Int("rport", 5000, "registry port you want to start,it must not be used")
	var wport = flag.Int("wport", 8088, "web port you want to listen, it must not be used")
	go webTask(*wport)
	if client, container, err := startRegistry(*rport, *wport); err != nil {
		log.Fatalf("err hanppen %v", err)
	} else {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for sig := range c {
				removeRegister(client, container)
				log.Fatalf("meet sig %+v", sig)
			}
		}()
		log.Printf("registry done!")
		select {}
	}
}
