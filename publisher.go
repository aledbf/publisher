package main

import (
  "errors"
  "log"
  "os"
  "regexp"
  "strconv"
  "time"

  "github.com/coreos/go-etcd/etcd"
  "github.com/fsouza/go-dockerclient"
)

func getopt(name, dfault string) string {
  value := os.Getenv(name)
  if value == "" {
    value = dfault
  }
  return value
}

func main() {
  endpoint := getopt("DOCKER_HOST", "unix:///var/run/docker.sock")

  client, err := docker.NewClient(endpoint)
  if err != nil {
    log.Fatal(err)
  }

  go listenContainers(client)

  go pollContainers(client)
}

func listenContainers(client *docker.Client) {
  var appName = regexp.MustCompile("from (*):v*")

  listener := make(chan *docker.APIEvents)
  defer func() { time.Sleep(10 * time.Millisecond); client.RemoveEventListener(listener) }()
  err := client.AddEventListener(listener)
  if err != nil {
    log.Fatal(err)
  }
  for {
    select {
    case event := <-listener:
      if event.Status == "start" {
        container, err := getContainer(client, event.ID)
        if err != nil {
          log.Fatal(err)
        }
        publishContainer(container)
      } else if event.Status == "stop" {
        filteredAppName := re.FindStringSubmatch(event.From)
        if len(filteredAppName) != 0 {
          log.Println("unpublishing service",filteredAppName[1])
          keyPath := "/deis/services/" + filteredAppName[1]
          unsetEtcd(etcd.NewClient([]string{"http://" + os.Getenv("ETCD_HOST") + ":4001"}), keyPath)
        }
      }
    }
  }
}

func getContainer(client *docker.Client, id string) (*docker.APIContainers, error) {
  containers, err := client.ListContainers(docker.ListContainersOptions{})
  if err != nil {
    return nil, err
  }

  for _, container := range containers {
    // send container to channel for processing
    if container.ID == id {
      return &container, nil
    }
  }
  return nil, errors.New("could not find container")
}

func pollContainers(client *docker.Client) {
  containers, err := client.ListContainers(docker.ListContainersOptions{})
  if err != nil {
    log.Fatal(err)
  }

  for _, container := range containers {
    // send container to channel for processing
    publishContainer(&container)
  }
}

func publishContainer(container *docker.APIContainers) {
  client := etcd.NewClient([]string{"http://" + os.Getenv("ETCD_HOST") + ":4001"})

  var publishableContainerName = regexp.MustCompile(`[a-z0-9-]+_v[1-9][0-9]*.(cmd|web).[1-9][0-9]*`)
  var publishableContainerBaseName = regexp.MustCompile(`^[a-z0-9-]+`)

  // this is where we publish to etcd
  for _, name := range container.Names {
    // HACK: remove slash from container name
    // see https://github.com/docker/docker/issues/7519
    containerName := name[1:]
    if !publishableContainerName.MatchString(containerName) {
      continue
    }
    containerBaseName := publishableContainerBaseName.FindString(containerName)
    keyPath := "/deis/services/" + containerBaseName + "/" + containerName
    for _, p := range container.Ports {
      host := os.Getenv("HOST")
      port := strconv.Itoa(int(p.PublicPort))
      setEtcd(client, keyPath, host+":"+port)
      // TODO: support multiple exposed ports
      break
    }
  }
}

func setEtcd(client *etcd.Client, key, value string) {
  _, err := client.Set(key, value)
  if err != nil {
    log.Println(err)
  }
  log.Println("set", key, "->", value)
}

func unsetEtcd(client *etcd.Client, key string) {
  _, err := client.Delete(key, true)
  if err != nil {
    log.Println(err)
  }
  log.Println("unset", key)
}
