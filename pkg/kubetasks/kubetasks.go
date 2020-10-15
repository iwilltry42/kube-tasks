package kubetasks

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/iwilltry42/kube-tasks/pkg/utils"
	"github.com/iwilltry42/skbn/pkg/skbn"
)

// SimpleBackup performs backup
func SimpleBackup(ctx context.Context, namespace, selector, container, path, dst string, parallel int, tag string, bufferSize float64) (string, error) {

	log.Infoln("Backup started!")
	dstPrefix, dstPath := utils.SplitInTwo(dst, "://")
	dstPath = filepath.Join(dstPath, tag)

	log.Infoln("Getting clients")
	k8sClient, dstClient, err := skbn.GetClients("k8s", dstPrefix, "", dstPath)
	if err != nil {
		return "", err
	}

	log.Infoln("Getting pods")
	pods, err := utils.GetReadyPods(ctx, k8sClient, namespace, selector)
	if err != nil {
		return "", err
	}

	if len(pods) == 0 {
		return "", fmt.Errorf("No pods were found in namespace %s by selector %s", namespace, selector)
	}

	log.Infoln("Calculating paths. This may take a while...")
	fromToPathsAllPods, err := utils.GetFromAndToPathsFromK8s(k8sClient, pods, namespace, container, path, dstPath)
	if err != nil {
		return "", err
	}

	log.Infoln("Starting files copy to tag: " + tag)
	if err := skbn.PerformCopy(k8sClient, dstClient, "k8s", dstPrefix, fromToPathsAllPods, parallel, bufferSize); err != nil {
		return "", fmt.Errorf("[SKBN] %+v", err)
	}

	log.Infoln("All done!")
	return tag, nil
}

// WaitForPods waits for a given number of pods
func WaitForPods(ctx context.Context, namespace, selector string, desiredReplicas int) error {
	log.Infoln("Getting clients")
	k8sClient, err := skbn.GetClientToK8s()
	if err != nil {
		return err
	}

	readyPods := -1
	log.Infof("Waiting for %d ready pods", desiredReplicas)
	for readyPods != desiredReplicas {
		pods, err := utils.GetReadyPods(ctx, k8sClient, namespace, selector)
		if err != nil {
			return err
		}
		readyPods = len(pods)
		log.Infof("Currently %d/%d ready pods", readyPods, desiredReplicas)
		if readyPods == desiredReplicas {
			break
		}
		time.Sleep(10 * time.Second)
	}
	return nil
}

// Execute executes simple commands in a container
func Execute(ctx context.Context, namespace, selector, container, command string) error {
	log.Infoln("Getting clients")
	k8sClient, err := skbn.GetClientToK8s()
	if err != nil {
		return err
	}

	log.Infoln("Getting pods")
	pods, err := utils.GetReadyPods(ctx, k8sClient, namespace, selector)
	if err != nil {
		return err
	}

	commandArray := strings.Fields(command)
	stdout := new(bytes.Buffer)
	stderr, err := skbn.Exec(*k8sClient, namespace, pods[0], container, commandArray, nil, stdout)
	if len(stderr) != 0 {
		return fmt.Errorf("STDERR: " + (string)(stderr))
	}
	if err != nil {
		return err
	}

	printOutput(stdout.String(), pods[0])
	return nil
}

func printOutput(output, pod string) {
	for _, line := range strings.Split(output, "\n") {
		if line != "" {
			log.Infoln(pod, line)
		}
	}
}
