package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)


func CreateStash(name string, image string) {
	userDir,_ := os.UserHomeDir()
	mount := fmt.Sprintf("%s:/app", filepath.Join("/home", userDir, "s3-bucket", "stashes", name))
	runCmd := exec.Command("docker", "run", "-d", "-it", "-P", "--name", name,"-v" ,mount , image)

	if err := runCmd.Start(); err != nil {
		fmt.Printf("Error creating stash: %s\n", err)
		return
	}

	if err := runCmd.Wait(); err != nil {
		fmt.Printf("Error running stash: %s\n", err)
		return
	}
	portsCmd := exec.Command("docker", "port", name)
	out, err := portsCmd.CombinedOutput()
	if err != nil {
		fmt.Println(err , out)
	}
	fmt.Println("ports",string(out))

	fmt.Printf("Container %s started successfully\n", name)
}