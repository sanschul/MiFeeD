# Tooling

This folder contains the different componets, I used to extract the data from a repo into a sql database.

## Dataextraction

Prerequirements:
- `docker`
- `golang`

The following Docker-Container are used in the process:
- [alpine/git](https://hub.docker.com/r/alpine/git)
- [zottelchin/cpp-stats](https://hub.docker.com/r/zottelchin/cppstats) - This could be build localy with the Dockerfile [Dockerfiles/Dockerfile-cppstats](Dockerfiles/Dockerfile-cppstats) (`docker build -t zottelchin/cppstats -f Dockerfile-cppstats .`) Attention: If you change the name of the image, you have to edit the config.go file in the Workbalancer directory by changing the name!
- zottelchin/ma-combine - this container has to be build for every different reposity as it cotains the python file responsible for uploding the data into the postgres database. 
  (`docker build -t zottelchin/ma-combine -f Dockerfile-combine .`) Attention: If you change the name of the image, you have to edit the config.go file in the Workbalancer directory by changing the name!

To Process a repository the following steps have to be executed:

1. Clone the repository you want to process into `Workbalancer/work/repo`. You might need to delete the .gitKeep File in the folder in order for git to accept this as an empty folder.
2. Add the commit hashes you want to process into the `commit.txt` file in the `Workbalancer` folder. For this the git-cli could be used with the function [rev-list](https://git-scm.com/docs/git-rev-list). 
   
    During the execution, another file called done.txt is created and every processed hash is stored in there. This could be quite handy in combination with the commit.txt to restart the process without processing commits twice.

3. If not already running, spin up a postgres database for your data. I used an additional docker container. For how to set up a postgres container, see https://hub.docker.com/_/postgres
   
   Note: If the container is set up in docker on the same machine the workbalancer will be running on, consider in the deployment, that the database is accessible from inside another docker container.
4. In this database create the table for your planned analysis. 
    ```sql
    CREATE TABLE libxml2 (
    id SERIAL PRIMARY KEY,
    constants TEXT,
    expression TEXT,
    type TEXT,
    file TEXT,
    hash TEXT
    );
    ```
5. Edit the `combine.py` according to the table you just created and create the docker image `zottelchin/ma-combine`. 

6. Now you can start the workbalancer with `go run *.go`.
    This starts the execution of the defined pipeline (see `config.go`) with the specified number of threads `main.go`.
    I would strongly advise to start this inside of `tmux` or something different. 

   

-----
Below i dumped all the commands i used to execute the process processing the libxml2 repo as an example

```
cd tools/Workbalancer/work
rm -r repo
git clone https://gitlab.gnome.org/GNOME/libxml2 repo
git rev-list master > ../../commits.txt

docker exec -it postgres-hSNk psql -U postgres
 > create database masterarbeit;
 > \c masterarbeit
 > CREATE TABLE libxml2 (
  id SERIAL PRIMARY KEY,
  constants TEXT,
  expression TEXT,
  type TEXT,
  file TEXT,
  hash TEXT
);

cd ../../Dockerfiles
nano combine.py
docker build -t zottelchin/ma-combine -f Dockerfile-combine .

cd ../Workbalancer/workbalancer
go run *.go
```

## Simple Metrics
To generate simple Metrics the python script in Metric Generation can be used. To install the dependancies run `pip3 install inquirer psycopg2-binary`. Then edit the file with the correct postgres parameters (there are two lines to change!)

Then execute the script, an select your table to analyse, like in the following output:

```
PostgreSQL connection is closed
[?] Which table should be processed?: libxml2
   openldap
 > libxml2

-------------------------------------------------------------------
found 8rows
PostgreSQL connection is closed
The five commits with the highest number of features are:
44ecefc8cc299a66ac21ffec141eb261e92638da - 1
7fbd454d9f70f0f0c0a0c27a7d541fed4d038c2a - 1
04d1bedd8c3fc5d9e41d11e2d0da08a966b732d3 - 1
067986fa674f0811614dab4c4572f5f7ff483400 - 1
4b3452d17123631ec43d532b83dc182c1a638fed - 2
-------------------------------------------------------------------
Anzahl Commits mit Änderungen in nur einem Feature:      4
Anzahl Commits mit Änderungen in mehr als einem Feature: 1
Anzahl Commits mit Änderungen an Featuren gesamt:        5
-------------------------------------------------------------------
The five features seen in the highest number of commits are:
LIBXML_PUSH_ENABLED            - 1
LIBXML_HTML_ENABLED            - 5
-------------------------------------------------------------------
Anzahl Features in nur einem Commit:      1
Anzahl Features in mehr als einem Commit: 1
Anzahl aller Features:                    2

```