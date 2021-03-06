package main

import (
  "os"
  "os/exec"
  "strings"
  "bytes"
  "bufio"
  "fmt"
  "strconv"
  "path"
  "sort"
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

func readChoice(choices []string, prompt string) (string, error) {
  for {
    input, err := readInput(prompt)
    if err != nil {
      return "", err
    }

    valid := false
    input = strings.ToLower(strings.TrimSpace(input))
    for _, choice := range choices {
      if input == choice {
        valid = true
        break
      }
    }
 
    if !valid {
      fmt.Printf("Invalid choice.\n")
    } else {
      return input, nil
    }
  }
}

func max(x, y int) int {
  if x > y {
    return x
  }
  return y
}

func printTable(table [][]string) {
  if len(table) == 0 {
    return
  }

  widths := make([]int, len(table[0]))
  for _, row := range table {
    for i, col := range row {
      widths[i] = max(widths[i], len(col))
    }
  }

  for _, row := range table {
    for i, col := range row {
      spaces := widths[i] - len(col) + 2
      fmt.Print(col)
      for j := 0; j < spaces; j++ {
        fmt.Print(" ")
      }
    }
    fmt.Print("\n")
  }
}

func chooseContainer(containers []*Container) (*Container, error) {
  choices := make(map[string]*Container)
  for i, container := range containers {
    input := strconv.FormatInt(int64(i + 1), 36)
    choices[input] = container
  }

  var keys []string
  for k := range choices {
    keys = append(keys, k)
  }
  sort.Strings(keys)

  var table [][]string
  for _, key := range keys {
    container := choices[key]

    row := []string{
      key,
      container.Name,
      container.Id,
      container.Image,
    }

    table = append(table, row)
  }

  printTable(table)

  if choice, err := readChoice(keys, "> "); err != nil {
    return nil, err
  } else {
    return choices[choice], nil
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

func findShells(container *Container, docker string) ([]string, error) {
  var buf bytes.Buffer
  cmd := exec.Command(docker, "exec", container.Id, "cat", "/etc/shells")
  cmd.Stdout = &buf
  err := cmd.Run()

  if err != nil {
    return []string{}, err
  }

  output := buf.String()
  lines := strings.Split(strings.TrimSpace(output), "\n")

  var shells []string
  for _, line := range lines {
    line = strings.TrimSpace(line)
    if path.IsAbs(line) {
      shells = append(shells, line)
    }
  }

  return shells, nil
}

func selectContainer(query, docker string) *Container {
  var containers []*Container
  if query == "" {
    containers = runningContainers(docker)
    if len(containers) == 0 {
      fmt.Printf("There are no running containers.")
      return nil
    }
  } else {
    containers = findContainers(query, docker)
    if len(containers) == 0 {
      fmt.Printf("Could not find any containers matching '%s'.", query)
      return nil
    } else if len(containers) > 1 {
      fmt.Printf("Multiple containers found for '%s':\n", query)
    }
  }

  var container *Container
  if len(containers) == 1 {
    container = containers[0]
  } else {
    var err error
    if container, err = chooseContainer(containers); err != nil {
      return nil
    }

    fmt.Println()
  }

  return container
}

func selectShell(container *Container, docker string) string {
  shells, err := findShells(container, docker)
  if err != nil {
    fmt.Printf("Could not find shell for %s.\n", container.Id)
    return ""
  }

  ranks := map[string]int {
    "zsh": 4,
    "bash": 3,
    "csh": 2,
    "tcsh": 2,
    "ksh": 2,
    "pdksh": 2,
    "fish": 2,
    "ash": 1,
    "sh": 0,
  }

  max := -2 
  shell := ""
  for _, shellPath := range shells {
    name := path.Base(shellPath)
    if rank, found := ranks[name]; found && rank > max {
      max = rank
      shell = shellPath
    }
  }

  return shell
}

var version string
var commit string

func main() {
  var query string
  if len(os.Args) >= 2 {
    query = os.Args[1]
  }

  if query == "-v" || query == "--version" {
    fmt.Println("dsh", version, commit)
    return
  }

  docker, err := exec.LookPath("docker")
  if err != nil {
    fmt.Println("'docker' not found.")
    return
  }

  var container *Container
  if container = selectContainer(query, docker); container == nil {
    return
  }

  var shell string
  if shell = selectShell(container, docker); shell == "" {
    return
  }

  fmt.Printf("Running %s in %s (%s).\n", shell, container.Name, container.Id)
  runCommand(container, docker, shell)
}
