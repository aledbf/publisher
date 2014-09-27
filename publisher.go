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

  etcdClient := etcd.NewClient([]string{"http://" + os.Getenv("ETCD_HOST") + ":4001"})
  // default consistency level is STRONG. Avoid reads from leader.
  etcdClient.SetConsistency("WEAK_CONSISTENCY")

  timeout := 300 * time.Second
  ttl := 600 * time.Second

  go listenContainers(client, etcdClient, ttl)

  for {
    go pollContainers(client, etcdClient, timeout, ttl)
    time.Sleep(timeout)
  }
}

func listenContainers(client *docker.Client, etcd *etcd.Client, ttl time.Duration) {
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
        publishContainer(container, etcd, ttl)
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

func pollContainers(client *docker.Client, etcd *etcd.Client, timeout time.Duration, ttl time.Duration) {
  containers, err := client.ListContainers(docker.ListContainersOptions{})
  if err != nil {
    log.Fatal(err)
  }

  for _, container := range containers {
    // send container to channel for processing
    publishContainer(&container, etcd, ttl)
  }
}

func publishContainer(container *docker.APIContainers, client *etcd.Client, ttl time.Duration) {  
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
      setEtcd(client, containerName, keyPath, host+":"+port, uint64(ttl.Seconds()))
      // TODO: support multiple exposed ports
      break
    }
  }
}

func setEtcd(client *etcd.Client, appName, key, value string, ttl uint64) {
  _, err := client.Set(key, value, ttl)
  if err != nil {
    log.Println(err)
  }
  log.Printf("[deis-publisher] mapping app %60s -> %s\n", appName, value)
}

func unsetEtcd(client *etcd.Client, key string) {
  _, err := client.Delete(key, true)
  if err != nil {
    log.Println(err)
  }
  log.Printf("[deis-publisher] removing app %s\n", key)
}