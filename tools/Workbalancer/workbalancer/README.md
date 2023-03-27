# workbalancer

To change the  steps executed, edit the config.go file.
To adjust the number of workers to the system, edit the main.go and change the number in workerpool.New(XXX). Take into consideration, every worker copies the complete repo, so the drive with the work directory  should at least be the size of the repository * number of workers + 1.

To start the processing run `go run *.go` in this folder. As this can take quite some time, I recommend the execution inside a tmux environment.