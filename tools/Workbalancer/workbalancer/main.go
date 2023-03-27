package main

import (
	"bufio"
	"github.com/gammazero/workerpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var finished *int
var errCount *int
var start time.Time

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	start = time.Now()
}

func main() {
	// clean up workdir
	log.Info().Msg("Cleaning up working directory")
	basepath, _ := filepath.Abs("../work")
	files, _ := ioutil.ReadDir(basepath)
	for _, f := range files {
		if f.IsDir() && f.Name() != "repo" && f.Name() != "results" {
			log.Debug().Str("Folder", f.Name()).Msg("Deleting folder")
			os.RemoveAll(filepath.Join(basepath, f.Name()))
		}
	}

	f := 0
	finished = &f
	e := 0
	errCount = &e
	log.Info().Msgf("Processing Pipeline %s", config.Name)
	repo, _ := filepath.Abs("../work/repo")
	commits := readCommits()
	log.Info().Int("Commitzahl", len(commits)).Msg("Commits eingelesen")

	pool := workerpool.New(4)  // Change for different number of Threads
	log.Debug().Int("Poolsize", pool.Size()).Msg("Workerpool created")

	for _, commit := range commits {
		c := commit // must be here to have the right var in scope for execution, would be always the last id otherwise. Thanks to @momar
		pool.Submit(func() {
			runSteps(config.Steps, c, repo)
		})
	}

	pool.StopWait()
	log.Info().Msg("Pipeline completed")

	// clean up workdir, cause sometime the function itself fails this task
	log.Info().Msg("Cleaning up working directory")
	files, _ = ioutil.ReadDir(basepath)
	for _, f := range files {
		if f.IsDir() && f.Name() != "repo" && f.Name() != "results" {
			log.Debug().Str("Folder", f.Name()).Msg("Deleting folder")
			os.RemoveAll(filepath.Join(basepath, f.Name()))
		}
	}

	runtime := time.Since(start)
	log.Info().Msgf("-------------------------------------------------")
	log.Info().Msgf("Runtime:          %s", runtime.Round(time.Second).String())
	log.Info().Msgf("Commits Input:    %d", len(commits))
	log.Info().Msgf("Commits Finished: %d", *finished)
	log.Info().Msgf("Errors:           %d", *errCount)
	log.Info().Msgf("Workercount       %d", pool.Size())
}

func readCommits() (commits []string) {
	f, err := os.Open("../commits.txt")
	if err != nil {
		log.Fatal().Err(err).Send()
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		commits = append(commits, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal().Err(err).Send()
	}
	return
}
