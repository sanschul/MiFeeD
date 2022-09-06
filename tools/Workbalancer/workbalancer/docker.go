package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	cp "github.com/otiai10/copy"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func runSteps(steps []Step, commit, repo string) {
	log.Info().Str("Commit", commit).Msg("Starting Pipeline")

	// Create working Directory for Runner
	basepath, _ := filepath.Abs("../work")
	workpath := filepath.Join(basepath, commit)
	if _, err := os.Stat(workpath); !os.IsNotExist(err) {
		// delete folder if it already exists
		log.Info().Str("Commit", commit).Msg("Removed old direcotry")
		os.RemoveAll(workpath)
	}
	os.Mkdir(workpath, 0777)
	err := cp.Copy(repo, filepath.Join(workpath, "repo"))
	if err != nil {
		*errCount++
		log.Error().Str("Commit", commit).Err(err).Send()
		log.Error().Str("Commit", commit).Msg("Aborting Pipeline Execution")
		os.RemoveAll(workpath)
		return
	}

	// Run every Step
	for _, step := range steps {
		if !runStep(step, commit, workpath, basepath) {
			log.Error().Str("Commit", commit).Msg("Aborting Pipeline Execution")
			os.RemoveAll(workpath)
			return
		}
		time.Sleep(1 * time.Second)
	}
	os.RemoveAll(workpath)
	writeline("../done.txt", commit)

	*finished++
	log.Info().Int("Finished", *finished).Str("Commit", commit).Msg("Pipeline complete")
}

func runStep(s Step, commit, workpath, basepath string) bool {
	log.Info().Str("Commit", commit).Str("Step", s.Name).Msgf("Running Step: %s", s.Name)
	cname := commit + "-" + s.Name

	// create Docker-Client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal().Str("Commit", commit).Str("Step", s.Name).Err(err).Msg("Failed to connect to docker")
	}
	defer cli.Close()

	//check if container with name exists and delete it
	list, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: filters.NewArgs(filters.Arg("name", cname))})
	log.Debug().Str("list", fmt.Sprintf("%+v", list)).Send()
	if len(list) != 0 {
		cli.ContainerRemove(ctx, list[0].ID, types.ContainerRemoveOptions{Force: true})
		log.Info().Str("Commit", commit).Str("Step", s.Name).Msg("Removed old container")
	}

	// Create Container for Running Pipelinestep
	ctn, err := cli.ContainerCreate(
		ctx,
		generateContainerConfig(commit, s),
		generateHostConfig(workpath, basepath),
		nil,
		nil,
		cname)
	if err != nil {
		log.Fatal().Str("Commit", commit).Str("Step", s.Name).Err(err).Msg("Failed to create Container")
	}

	// Run Container
	//log.Info().Str("Commit", commit).Str("Step", s.Name).Msg("Running Container")
	err = cli.ContainerStart(context.Background(), ctn.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Error().Str("Commit", commit).Str("Step", s.Name).Err(err).Msg("Failed to start container")
	}

	//log execution
	go logging(ctn.ID, commit, s.Name)

	// wait two seconds so logs can be read before container is removed
	time.Sleep(time.Second * 2)

	statusCh, errCh := cli.ContainerWait(context.Background(), ctn.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			log.Error().Str("Commit", commit).Str("Step", s.Name).Err(err).Msg("Error from Container")
		}
	case status := <-statusCh:
		log.Debug().Str("Commit", commit).Str("Step", s.Name).Int64("Exit-code", status.StatusCode).Msg("Container exited")
		if status.StatusCode != 0 {
			log.Error().Str("Commit", commit).Str("Step", s.Name).Int64("Exit-code", status.StatusCode).Msg("Container exited with error")
			*errCount++
			cli.ContainerRemove(context.Background(), ctn.ID, types.ContainerRemoveOptions{})
			return false
		}
	}
	cli.ContainerRemove(context.Background(), ctn.ID, types.ContainerRemoveOptions{})

	return true
}

func logging(cid, commit, step string) {
	// create Docker-Client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal().Str("Commit", commit).Str("Step", step).Str("Container", cid).Err(err).Msg("Failed to connect to docker")
	}
	defer cli.Close()

	// wait one sec to ensure container is startet
	time.Sleep(time.Second)

	// Subscribe to Logs for Container
	reader, err := cli.ContainerLogs(context.Background(), cid, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: false,
		Follow:     true,
	})
	defer reader.Close()

	if err != nil {
		log.Error().Str("Commit", commit).Str("Step", step).Str("Container", cid).Err(err).Msg("Failed to connect to container-logs")
		return
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		log.Debug().Str("Commit", commit).Str("Step", step).Str("Container", cid).Msg(scanner.Text())
	}
}

// Copied from https://github.com/woodpecker-ci/woodpecker/blob/9e8d1a9294d1bdced7c35e613fc8816e7988537f/pipeline/frontend/yaml/compiler/script_posix.go#L12-L28
func generateScriptPosix(commands []string) string {
	var buf bytes.Buffer
	for _, command := range commands {
		escaped := fmt.Sprintf("%q", command)
		escaped = strings.Replace(escaped, "$", `\$`, -1)
		buf.WriteString(fmt.Sprintf(
			traceScript,
			escaped,
			command,
		))
	}
	script := fmt.Sprintf(
		buf.String(),
	)
	return base64.StdEncoding.EncodeToString([]byte(script))
}

func generateHostConfig(workpath, basepath string) *container.HostConfig {
	return &container.HostConfig{
		LogConfig: container.LogConfig{},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: workpath,
				Target: "/work",
			},
			{
				Type:   mount.TypeBind,
				Source: filepath.Join(basepath, "results"),
				Target: "/results",
			},
		},
	}
}

func generateContainerConfig(commit string, s Step) *container.Config {
	return &container.Config{
		Image:        s.Image,
		WorkingDir:   "/work",
		Entrypoint:   strslice.StrSlice{"/bin/sh", "-c"},
		Cmd:          strslice.StrSlice{"echo " + generateScriptPosix(s.Commands) + " | base64 -d | /bin/sh -e"},
		Env:          []string{"COMMIT=" + commit},
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}
}

var ctx = context.Background()

const traceScript = `
echo + %s
%s
`

func writeline(file, line string) {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
	if err != nil {
		log.Error().Err(err).Send()
	}
	defer f.Close()
	f.WriteString(line + "\n")
}
