package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"time"
	"flag"
	"log"

	"github.com/dietsche/rfsnotify"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"

	"github.com/Lookyan/autohelm/kubeutils"
)


var (
	kubeConfig = flag.String("kubeconfig", GetHome() + "/.kube/config", "Absolute path to the kubeconfig file")

	name = flag.String("name", "name", "Name of your project")
	helmDir = flag.String("helmdir", "helm", "Path to dir with helm chart")
	configFile = flag.String("configfile", "values.yaml", "Path to additional config file")
	namespace = flag.String("namespace", "default", "Kubernetes namespace")
	tillerNamespace = flag.String("tiller-namespace", "kube-system", "Tiller namespace")
	del = flag.Bool("d", false, "Delete chart after autohelming (WIP)")
	threshold = flag.Int("t", 5, "Seconds to wait for rebuild")
	imageTagName = flag.String("image-tag-name", "latest", "Name of image tag variable in helm chart")

	attach = flag.Bool("attach", false, "Auto attach")
	deploymentName = flag.String("deploy", "", "Deployment name for attach")
	containerName = flag.String("container", "", "Container name for attach")
)

var lastChangeTime = time.Now()
var haveChanges = false
var currentAttachCommand *exec.Cmd


func GetHome() string {
	usr, _ := user.Current()
	dir := usr.HomeDir

	return dir
}

func GenerateTag() string {
	t := time.Now()
	return fmt.Sprintf("dev-%s", t.Format("20060102150405"))
}

func RunCommand(name string, arg ...string) error {

	cmd := exec.Command(name, arg...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	return err
}

func RunCommandInBackground(name string, arg ...string) *exec.Cmd {

	cmd := exec.Command(name, arg...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()

	return cmd
}

func Redeploy(clientset *kubernetes.Clientset) {

	if currentAttachCommand != nil && !currentAttachCommand.ProcessState.Exited() {
		currentAttachCommand.Process.Signal(os.Kill)
	}

	tag := GenerateTag()
	imageName := fmt.Sprintf("%s:%s", *name, tag)

	fmt.Printf("Building %s...\n\n", imageName)

	// docker should be configured to minikube's dockerd

	err := RunCommand("docker", "build", "-t", imageName, ".")

	if err != nil {
		fmt.Println("Can't build docker image!")
		return
	}

	os.Chdir(*helmDir)

	err = RunCommand("helm",
		"upgrade",
		"--install",
		"--debug",
		"--wait",
		"--namespace",
		*namespace,
		"--tiller-namespace",
		*tillerNamespace,
		"-f",
		*configFile,
		"--set",
		*imageTagName + "=" + tag,
		*name,
		".")

	if err != nil {
		fmt.Println("Error occured while deploy helm chart")
		return
	}

	fmt.Println("Happy autohelming!")
	os.Chdir("..")

	if *attach {
		Attach(clientset)
	}
}

func PollReBuild(clientset *kubernetes.Clientset) {
	for {
		if haveChanges == true && time.Since(lastChangeTime) > time.Duration(*threshold)*time.Second {
			fmt.Println("Rebuilding...")
			haveChanges = false
			Redeploy(clientset)
		}
	}
}

func Attach(clientset *kubernetes.Clientset) {

	if *deploymentName == "" || *containerName == "" {
		fmt.Println("You should enter deployment name and container for attach")
		return
	}

	podName, err := kubeutils.GetDeploymentPod(clientset, *namespace, *deploymentName, *containerName)

	if err != nil {
		fmt.Printf("Can't attach: %s", err)
		return
	}

	fmt.Printf("Attaching to %s\n", podName)

	currentAttachCommand = RunCommandInBackground(
		"kubectl",
		"attach",
		"--namespace",
		*namespace,
		"-c",
		*containerName,
		podName)

}

func main() {
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	//Redeploy(clientset)

	watcher, err := rfsnotify.NewWatcher()

	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case _ = <-watcher.Events:
				fmt.Println("Changes!")
				lastChangeTime = time.Now()
				haveChanges = true
			case err := <-watcher.Errors:
				fmt.Println("error:", err)
			}
		}
	}()

	go PollReBuild(clientset)

	wd , _ := os.Getwd()
	fmt.Println("\n\nListening ", wd)
	err = watcher.AddRecursive(wd)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
