package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

const endpoint = "unix:///var/run/docker.sock"

func attachLog(repo string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	client, err := docker.NewClient(endpoint)
	if err != nil {
		return nil, err
	}
	opts := docker.AttachToContainerOptions{
		Container:    "sirdm_" + repo,
		OutputStream: &buf,
		Stdout:       true,
		Stderr:       true,
		Logs:         true,
	}
	err = client.AttachToContainer(opts)
	if err != nil {
		return nil, err
	}
	return &buf, nil
}

func restartDockerWithNewImage(r *record, force bool) {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		log.Printf("client init fail when restart container:%+v", r)
		return
	}
	Repository := r.Repository
	containerName := "sirdm_" + Repository
	remoteRepo := "127.0.0.1:5000/" + Repository

	time.Sleep(1 * time.Second) //wait a little while since we just receive the notify

	runningImage := ""
	if container, errInspectContainer := client.InspectContainer(containerName); errInspectContainer == nil {
		runningImage = container.Image
		if !container.State.Running {
			log.Printf("container not running, force start...")
			force = true
		}
	}

	err = client.PullImage(docker.PullImageOptions{
		Repository: remoteRepo,
	}, docker.AuthConfiguration{})
	if err != nil {
		log.Printf("%s:%+v", remoteRepo, err)
		return
	}

	time.Sleep(1 * time.Second) //wait a little while since we just finish pull
	image, errInspectImage := client.InspectImage(remoteRepo)
	if errInspectImage == nil {
		if !force {
			if image.ID == runningImage {
				log.Printf("same image , no need restart")
				return
			}
			log.Printf("%s != %s, need restart ...", runningImage, image.ID)
		} else {
			log.Printf("skip version check since force is true")
		}
	}

	client.StopContainer(containerName, 10)
	client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            containerName,
		RemoveVolumes: true,
	})
	pbs := make(map[docker.Port][]docker.PortBinding)
	eps := make(map[docker.Port]struct{})
	ports := strings.Split(r.Ports, ",")
	log.Printf("need open port:%s", r.Ports)
	for _, port := range ports {
		pbs[docker.Port(fmt.Sprintf("%s/tcp", port))] = []docker.PortBinding{
			docker.PortBinding{
				HostIP:   "0.0.0.0",
				HostPort: port,
			},
		}
		eps[docker.Port(fmt.Sprintf("%s/tcp", port))] = struct{}{}
	}
	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: containerName,
		Config: &docker.Config{
			Image:        remoteRepo,
			ExposedPorts: eps,
		},
		HostConfig: &docker.HostConfig{
			PortBindings: pbs,
		},
	})
	if err = client.StartContainer(container.ID, nil); err != nil {
		log.Printf("start fail! %s", err.Error())
	} else {
		log.Printf("new version of %s restarted!", Repository)
		saveRecord(&record{
			Repository: Repository,
			Version:    image.ID,
			RebootTime: time.Now(),
		})
	}
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
		if i.Name != "en0" && i.Name != "en1" {
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
