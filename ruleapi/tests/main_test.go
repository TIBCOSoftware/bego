package tests

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	image = "none"
	code := m.Run()
	if code != 0 {
		os.Exit(code)
	}

	store = "rsconfigmp.json"
	code = m.Run()
	if code != 0 {
		os.Exit(code)
	}

	run := func(img string) int {
		image = img
		command := exec.Command("docker", "run", "-p", "6380:6379", "-d", img)
		hash, err := command.Output()
		if err != nil {
			panic(err)
		}
		Pour("6380")

		defer func() {
			command := exec.Command("docker", "stop", strings.TrimSpace(string(hash)))
			err := command.Run()
			if err != nil {
				panic(err)
			}
			command = exec.Command("docker", "rm", strings.TrimSpace(string(hash)))
			err = command.Run()
			if err != nil {
				panic(err)
			}
			Drain("6380")
		}()

		return m.Run()
	}
	store = "rsconfig.json"
	code = run("redis")
	if code != 0 {
		os.Exit(code)
	}

	store = "rsconfigp.json"
	code = run("redis")
	if code != 0 {
		os.Exit(code)
	}

	store = "rsconfig.json"
	code = run("eqalpha/keydb")
	if code != 0 {
		os.Exit(code)
	}

	store = "rsconfigp.json"
	os.Exit(run("eqalpha/keydb"))
}
