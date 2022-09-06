package main

type Config struct {
	Name  string
	Steps []Step
}

type Step struct {
	Name     string
	Image    string
	Commands []string
}

var config = Config{
	Name: "Masterarbeitstest",
	Steps: []Step{
		{
			Name:  "checkout",
			Image: "alpine/git",
			Commands: []string{
				"echo \"Running Script for Commit $COMMIT $PWD\"",
				//"ls -la",
				"cd repo",
				"git -c advice.detachedHead=false checkout $COMMIT",
				"git show --pretty=\"format:\" --name-only --diff-filter=d -- *.c > ../changed-files.txt",
			},
		},
		{
			Name:  "copy-changes",
			Image: "alpine/git",
			Commands: []string{
				"mkdir -p /work/changed/source",
				"while read -r line; do mkdir -p /work/changed/source/$(dirname $line); cp /work/repo/$line /work/changed/source/$line; echo \"copied $line\";done < /work/changed-files.txt",
				"find /work/changed/source -print",
			},
		},
		{
			Name:  "cpp-stats",
			Image: "zottelchin/cppstats",
			Commands: []string{
				"cd /work", // Müssten wir schon sein
				"echo \"$(pwd)/changed\" > cppstats_input.txt",
				"cppstats --kind featurelocations",
			},
		},
		{
			Name:     "combine",
			Image:    "zottelchin/ma-combine",
			Commands: []string{"python /app/combine.py $COMMIT"},
		},
	},
}
