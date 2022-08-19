package install

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	ErrKubectlInstall  = errors.New("kubectl is not installed")
	ErrKubeconfigNs    = errors.New("set the kubeconfig or kubeconfig does not have required permissions")
	ErrArgoCD          = errors.New("argocd is not installed")
	ErrArgoCDAuthToken = errors.New("argocd auth token has expired, login to argocd again")
	ErrGit             = errors.New("git is not installed")
	ErrKubectlPresent  = errors.New("kubectl is already installed")
	ErrKubectlFail     = errors.New("error installing kubectl")
	kubectlPath        = "/usr/local/bin/kubectl"
)

func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:               "install",
		Short:             "Install required tools for Arlon",
		Long:              "Install the missing tools that are required for Arlon",
		DisableAutoGenTag: true,
		Example:           "arlonctl install",
		RunE: func(c *cobra.Command, args []string) error {
			// Install kubectl and point it to the kubeconfig
			_, err := installKubectl()
			if err == ErrKubectlPresent {
				fmt.Println("Kubectl is already present in the path")
			} else if err != nil {
				fmt.Println("Error while installing kubectl ", err)
			} else {
				fmt.Println("Successfully Installed Kubectl")
			}
			return nil
		},
	}
	return command
}

// Check if kubectl is installed and the kubeconfig is pointing to the correct kubeconfig
func installKubectl() (bool, error) {
	_, err := exec.LookPath("kubectl")
	if err != nil {
		fmt.Println("Proceeding to install kubectl")
		errInstall := installKubectlPlatform()
		if errInstall != nil {
			fmt.Println("Error installing kubectl")
			return false, ErrKubectlFail
		} else {
			return true, nil
		}
	}
	return true, ErrKubectlPresent
}

// Check the platform and on the basis of that install kubectl
func installKubectlPlatform() error {
	osPlatform := runtime.GOOS
	switch osPlatform {
	case "darwin":
		fmt.Println("Installing kubectl on darwin system")
		err1 := downloadKubectlLatest(osPlatform)
		if err1 != nil {
			fmt.Println("Error installing the latest kubectl version")
		}
		_, err2 := exec.Command("chmod", "+x", kubectlPath).Output()
		if err2 != nil {
			fmt.Println("Error giving access to kubectl")
			return err2
		}

	case "windows":
		fmt.Println("Installing kubectl on windows system")
		err1 := downloadKubectlLatest(osPlatform)
		if err1 != nil {
			fmt.Println("Error installing the latest kubectl version")
		}
		fmt.Println("Add kubectl binary to your windows path")

	case "linux":
		fmt.Println("Installing kubectl on linux system")
		err1 := downloadKubectlLatest(osPlatform)
		if err1 != nil {
			fmt.Println("Error installing the latest kubectl version")
		}
		_, err2 := exec.Command("chmod", "+x", kubectlPath).Output()
		if err2 != nil {
			fmt.Println("Error giving access to kubectl")
			return err2
		}
	}
	return nil
}

// Downloads the latest version of kubectl
func downloadKubectlLatest(osPlatform string) error {
	latestVersion := "https://storage.googleapis.com/kubernetes-release/release/stable.txt"
	ver, err := exec.Command("curl", "-sL", latestVersion).Output()
	if err != nil {
		fmt.Println("Error fetching latest kubectl version")
		return err
	}
	var downloadKubectl string
	if osPlatform == "windows" {
		downloadKubectl = "https://storage.googleapis.com/kubernetes-release/release/" + string(ver) + "/bin/" + osPlatform + "/amd64/kubectl.exe"
	} else {
		downloadKubectl = "https://storage.googleapis.com/kubernetes-release/release/" + string(ver) + "/bin/" + osPlatform + "/amd64/kubectl"
	}
	fmt.Println(downloadKubectl)
	_, err1 := exec.Command("curl", "-o", kubectlPath, "-LO", downloadKubectl).Output()
	if err1 != nil {
		fmt.Println("Error downloading latest kubectl version")
		return err1
	}
	return nil
}
