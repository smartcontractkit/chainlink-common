package utils

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

type Gauntlet struct {
	exec string
}

type FlowReport []struct {
	Name string `json:"name"`
	Txs  []struct {
		Contract string `json:"contract"`
		Hash     string `json:"hash"`
		Success  bool   `json:"success"`
	}
	Data   map[string]string `json:"data"`
	StepId int               `json:"stepId"`
}

type Report struct {
	Responses []struct {
		Tx struct {
			Hash    string `json:"hash"`
			Address string `json:"address"`
		}
		Contract string `json:"contract"`
	} `json:"responses"`
	Data map[string]string `json:"data"`
}

func NewGauntlet(binPath string) (Gauntlet, error) {

	_, err := exec.Command(binPath).Output()
	if err != nil {
		return Gauntlet{}, errors.New("gauntlet installation check failed")
	}
	return Gauntlet{
		exec: binPath,
	}, nil
}

func (g Gauntlet) Flag(flag string, value string) string {
	return fmt.Sprintf("--%s=%s", flag, value)
}

func (g Gauntlet) ExecCommand(args ...string) error {
	cmd := exec.Command(g.exec, args...)
	pipe, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return err
	}
	reader := bufio.NewReader(pipe)
	line, err := reader.ReadString('\n')
	for err == nil {
		fmt.Print(line)
		line, err = reader.ReadString('\n')
	}
	return nil
}

func (g Gauntlet) ReadCommandReport() (Report, error) {
	jsonFile, err := os.Open("report.json")
	if err != nil {
		return Report{}, err
	}

	var report Report
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &report)

	return report, nil
}

func (g Gauntlet) ReadCommandFlowReport() (FlowReport, error) {
	jsonFile, err := os.Open("flow-report.json")
	if err != nil {
		return FlowReport{}, err
	}

	var report FlowReport
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return FlowReport{}, err
	}
	err = json.Unmarshal(byteValue, &report)
	if err != nil {
		return FlowReport{}, err
	}

	return report, nil
}
