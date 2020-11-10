package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type PodmanImage struct {
	Name string
}

type PodmanPod struct {
	Name  string
	Share string
}

type PodmanService struct {
	Name string
}

type PodmanContainer struct {
	Name    string
	Service *PodmanService
}

type PodmanCmd struct {
	Podman *Podman
	Args   []string
}

type Podman struct {
	path   string
	dryRun bool
}

func NewPodman(path string, dryRun bool) (*Podman, error) {
	path, err := normalizePodmanPath(path)
	if err != nil {
		return nil, err
	}

	return &Podman{path: path, dryRun: dryRun}, nil
}

func NewPodmanWithDefaults() (*Podman, error) {
	path, err := normalizePodmanPath("podman")
	if err != nil {
		return nil, err
	}

	return &Podman{path: path, dryRun: false}, nil
}

func (p *Podman) GetVersion() (string, error) {
	output, err := p.Output([]string{"--version"})
	if err != nil {
		return "", err
	}

	return strings.SplitAfter(output, "podman version ")[1], nil
}

func (p *Podman) PodCreate(pod *PodmanPod) error {
	_, err := p.Run([]string{"pod", "create", "--name=" + pod.Name, "--share=" + pod.Share}, true, 0)

	return err
}

func (p *Podman) PodRemove(pod *PodmanPod) error {
	_, err := p.Run([]string{"pod", "rm", pod.Name}, true, 0)

	return err
}

func (p *Podman) ImageGetId(image *PodmanImage) (string, error) {
	output, err := p.Output([]string{"inspect", "-t", "image", "-f", "{{.Id}}", image.Name})

	return strings.TrimSpace(output), err
}

func (p *Podman) ImagePull(image *PodmanImage) error {
	_, err := p.Run([]string{"pull", image.Name}, true, 0)

	return err
}

func (p *Podman) ImagePush(image *PodmanImage) error {
	_, err := p.Run([]string{"push", image.Name}, true, 0)

	return err
}

func (p *Podman) ContainerStop(container *PodmanContainer, args []string) error {
	cmd := []string{"stop"}
	cmd = append(cmd, args...)
	cmd = append(cmd, container.Name)

	_, err := p.Run(cmd, true, 0)

	return err
}

func (p *Podman) Logs(container *PodmanContainer, follow bool, timestamps bool, tail string) error {
	cmd := []string{"logs"}
	if follow {
		cmd = append(cmd, "-f")
	}
	if timestamps {
		cmd = append(cmd, "-t")
	}
	if tail != "" && tail != "all" {
		cmd = append(cmd, "--tail", tail)
	}

	cmd = append(cmd, container.Name)

	_, err := p.Run(cmd, true, 0)

	return err
}

func (p *Podman) Ps(all bool, quiet bool, projectName string) error {
	cmd := []string{"ps"}
	if all {
		cmd = append(cmd, "-a")
	}
	if quiet {
		cmd = append(cmd, "--format", "{{.ID}}")
	}
	if projectName != "" {
		cmd = append(cmd, "--filter", "label=io.podman.compose.project="+projectName)
	}

	_, err := p.Run(cmd, true, 0)

	return err
}

func (p *Podman) Output(args []string) (string, error) {
	fmt.Println("podman cmd: " + p.path + " " + strings.Join(args, " "))

	output, err := exec.Command(p.path, args...).Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (p *Podman) Run(args []string, wait bool, sleep int) (*exec.Cmd, error) {
	fmt.Println("podman cmd: " + p.path + " " + strings.Join(args, " "))

	if p.dryRun {
		return nil, nil
	}

	cmd := exec.Command(p.path, args...)
	stdout, _ := cmd.StdoutPipe()

	err := cmd.Start()
	if err != nil {
		return cmd, err
	}

	reader := bufio.NewReader(stdout)
	line, err := reader.ReadString('\n')
	for err == nil {
		fmt.Println(line)
		line, err = reader.ReadString('\n')
	}

	if wait {
		err = cmd.Wait()
		if err != nil {
			return cmd, err
		}
	}

	if sleep != 0 {
		time.Sleep(time.Duration(sleep))
	}

	return cmd, nil
}

type PodmanCompose struct {
	commands    map[string]*PodmanCmd
	globalArgs  []string
	projectName string
	directory   string
	pods        []*PodmanPod
	containers  []*PodmanContainer
}

func normalizePodmanPath(path string) (string, error) {
	if path == "podman" {
		return path, nil
	}

	pathStat, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if pathStat.Mode().IsRegular() {
		return filepath.Abs(path)
	}

	return "", errors.New("Binary " + path + " has not been found")
}

func main() {
	podman, err := NewPodman("/usr/bin/podman", false)
	//podman, err := NewPodmanWithDefaults()
	if err != nil {
		log.Fatal(err)
	}

	podman.Ps(false, false, "")
}
