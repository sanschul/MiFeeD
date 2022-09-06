# Workbalancer

The `work` folder contains the repository and results. The folder is used as working directory and should be mounted using tmpfs.

In `commits.txt` the commit hashes to be processed are stored. To process all commits, execute `git rev-list master > ../../commits.txt` in the repo folder in the working directory.