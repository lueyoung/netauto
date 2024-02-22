package core

import (
	"fmt"
	"io/ioutil"
	//"log"
	//"math/rand"
	"os/exec"
	//"strconv"
	//"strings"
)

func execCommand(strCommand string) string {
	cmd := exec.Command("/bin/bash", "-c", strCommand)

	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		fmt.Println("Execute failed when Start:" + err.Error())
		return ""
	}

	out_bytes, _ := ioutil.ReadAll(stdout)
	stdout.Close()

	if err := cmd.Wait(); err != nil {
		fmt.Println("Execute failed when Wait:" + err.Error())
		return ""
	}
	str := string(out_bytes)
	return str
}

func execCmd(strCommand string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", strCommand)

	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		fmt.Println("Execute failed when Start:" + err.Error())
		return "", err
	}

	out_bytes, _ := ioutil.ReadAll(stdout)
	stdout.Close()

	if err := cmd.Wait(); err != nil {
		fmt.Println("Execute failed when Wait:" + err.Error())
		return "", err
	}
	str := string(out_bytes)
	return str, nil
}

func execCmdVerbose(strCommand string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", strCommand)

	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		fmt.Println("Execute failed when Start:" + err.Error())
		//return "", err
	}

	out_bytes, _ := ioutil.ReadAll(stdout)
	stdout.Close()

	if err := cmd.Wait(); err != nil {
		fmt.Println("Execute failed when Wait:" + err.Error())
		//return "", err
	}
	str := string(out_bytes)
	return str, nil
}

func execCmdLogged(strCommand string) (string, error) {
	//strCommand += fmt.Sprintf(" >> %v", string(kubelog))
	return execCmd(strCommand)
}
