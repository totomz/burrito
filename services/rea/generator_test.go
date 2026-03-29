package rea

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/totomz/burrito/common"
)

func TestGetProjectRoot(t *testing.T) {
	wd, err := projectRoot(common.MustGetCwd())
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(wd) != "template-burrito" {
		t.Fatal("root project folder not found")
	}

	_, err = projectRoot("/Users/totomz/daje/sempre/forza/")
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestCloneTemplate(t *testing.T) {
	seed := ServiceSeed{
		ServiceName: fmt.Sprintf("zalice"),
	}

	t.Cleanup(func() {
		// _ = DeleteServiceDirectory(seed)
	})

	err := CreateServiceDirectory(seed)
	if err != nil {
		t.Fatal(err)
	}

	err = CloneTemplate(seed)
	if err != nil {
		t.Fatal(err)
	}

	// Try to compile the newly created service
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = seed.ServicePath()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to compile generated service: %v\nOutput: %s", err, string(output))
	}
	t.Logf("Successfully compiled generated service at %s", seed.ServicePath())
}
