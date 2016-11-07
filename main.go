package main

import (
  "os"
  "os/exec"
  "strings"
  "bytes"
  "bufio"
  "fmt"
  "strconv"
)

type Container struct {
  Id string
  Image string
  Name string
}

func parseContainers(output string) []*Container {
  lines := strings.Split(output, "\n")

  var containers []*Container
  for _, line := range lines {
    line = strings.TrimSpace(line)
    if line == "" {
      continue
    }

    parts := strings.Split(line, " ")

    containers = append(containers, &Container {
      Id: parts[0],
      Image: parts[1],
      Name: parts[2],
    })
  }

  return containers
}

func runningContainers(docker string) []*Container {
  var buf bytes.Buffer
  cmd := exec.Command(docker, "ps", "--format", "{{.ID}} {{.Image}} {{.Names}}")
  cmd.Stdout = &buf
  cmd.Run()

  output := buf.String()
  return parseContainers(output)
}

func findContainersByField(field, query, docker string) []*Container {
  var buf bytes.Buffer
  cmd := exec.Command(docker, "ps", "--format", "{{.ID}} {{.Image}} {{.Names}}", "--filter", field + "=" + query)
  cmd.Stdout = &buf
  cmd.Run()

  output := buf.String()
  return parseContainers(output)
}

func findContainers(query, docker string) []*Container {
  if containers := findContainersByField("id", query, docker); len(containers) > 0 {
    return containers
  } else {
    return findContainersByField("name", query, docker)
  }
}

func readInput(prompt string) (string, error) {
  reader := bufio.NewReader(os.Stdin)
  fmt.Fprint(os.Stderr, prompt)
  return reader.ReadString('\n')
}

func readIndex(low, high int, prompt string) (int, error) {
  for {
    input, err := readInput(prompt)
    if err != nil {
      return 0, err
    }

    input = strings.TrimSpace(input)
    index, err := strconv.Atoi(input)
    if (err != nil) || (index < low) || (index > high) {
      fmt.Printf("Invalid choice.\n")
    } else {
      return index, nil
    }
  }
}

func chooseContainer(containers []*Container) (*Container, error) {
  for i, container := range containers {
    fmt.Printf("%d. %s %s %s\n", i + 1, container.Id, container.Image, container.Name)
  }

  if index, err := readIndex(1, len(containers), "> "); err != nil {
    return nil, err
  } else {
    return containers[index - 1], nil
  }
}

func findCommand(container *Container, docker, command string) (string, bool) {
  var buf bytes.Buffer
  cmd := exec.Command(docker, "exec", container.Id, "which", command)
  cmd.Stdout = &buf
  err := cmd.Run()

  if err != nil {
    return "", false
  }

  output := buf.String()
  lines := strings.Split(strings.TrimSpace(output), "\n")

  if len(lines) == 0 {
    return "", false
  } else {
    command := strings.TrimSpace(lines[0])
    if len(command) > 0 {
      return command, true
    } else {
      return "", false
    }
  }
}

func runCommand(container *Container, docker, command string) error {
  cmd := exec.Command(docker, "exec", "-it", container.Id, command)
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  cmd.Stdin = os.Stdin
  return cmd.Run()
}

func main() {
  var query string
  if len(os.Args) >= 2 {
    query = os.Args[1]
  }

  docker, err := exec.LookPath("docker")
	if err != nil {
    fmt.Println("'docker' not found.")
    return
	}

  var containers []*Container
  if query == "" {
    containers = runningContainers(docker)
    if len(containers) == 0 {
      fmt.Printf("There are no running containers.")
      return
    }
  } else {
    containers = findContainers(query, docker)
    if len(containers) == 0 {
      fmt.Printf("Could not find container matching '%s'.", query)
      return
    } else if len(containers) > 1 {
      fmt.Printf("Multiple containers found for '%s'.\n", query)
    }
  }

  var container *Container
  if len(containers) == 1 {
    container = containers[0]
  } else {
    container, err = chooseContainer(containers)
    if err != nil {
      return
    }

    fmt.Println()
  }

  var shell string
  var found bool
  if shell, found = findCommand(container, docker, "zsh"); !found {
    if shell, found = findCommand(container, docker, "bash"); !found {
      shell, found = findCommand(container, docker, "sh")
    }
  }

  if !found {
    fmt.Printf("Could not find shell for %s.\n", container.Id)
    return
  }

  fmt.Printf("Running %s in %s (%s).\n", shell, container.Name, container.Id)
  runCommand(container, docker, shell)
}
