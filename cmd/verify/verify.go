package verify

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	ErrKubectlInstall  = errors.New("kubectl is not installed")
	ErrClusterInfo     = errors.New("set the kubeconfig or kubeconfig does not have required access")
	ErrArgoCD          = errors.New("argocd is not installed")
	ErrArgoCDAuthToken = errors.New("argocd auth token has expired, login to argocd again")
	ErrGit             = errors.New("git is not installed")
	ErrArlonNs         = errors.New("arlon is not installed")
	ErrCapi            = errors.New("capi services are not installed")
	ErrCapiCP          = errors.New("error fetching the capi cloudproviders")
)

func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:               "verify",
		Short:             "Verify if arlonctl can be run",
		Long:              "Verify if required kubectl,argocd,git access is present before profiles and bundles are created",
		DisableAutoGenTag: true,
		Example:           "arlonctl verify",
		RunE: func(c *cobra.Command, args []string) error {
			// Verify kubectl status
			kubectlStatus, err := verifyKubectl()
			if err != nil {
				fmt.Println("Error while verifying kubectl status: ", err)
			} else {
				fmt.Println("Successfully verified kubectl status")
			}

			// Verify argocd status
			argoStatus, err1 := verifyArgoCD()
			if err1 != nil {
				fmt.Println("Error while verifying argocd status: ", err1)
			} else {
				fmt.Println("Successfully verified argocd status")
			}

			// Verify git status
			gitStatus, err2 := verifyGit()
			if err2 != nil {
				fmt.Println("Error while verifying git status: ", err2)
			} else {
				fmt.Println("Successfully verified git status")
			}

			// Verify capi status
			capiStatus, err4 := verifyCapi()
			if err4 != nil {
				fmt.Println("Error while verifying capi status: ", err2)
			} else {
				fmt.Println("Successfully verified capi status")
			}

			// Verify arlon status
			arlonStatus, err3 := verifyArlon()
			if err3 != nil {
				fmt.Println("Error while verifying arlon status: ", err2)
			} else {
				fmt.Println("Successfully verified arlon status")
			}

			fmt.Println()
			if kubectlStatus && argoStatus && gitStatus && arlonStatus && capiStatus {
				fmt.Println("All requirements are installed")
			} else {
				fmt.Println("The check for Arlon prerequisites failed. Please install the missing tool(s).")
			}

			return nil
		},
	}
	return command
}

// Check if kubectl is installed and the kubeconfig is pointing to the correct kubeconfig
func verifyKubectl() (bool, error) {
	_, err := exec.LookPath("kubectl")
	if err != nil {
		return false, ErrKubectlInstall
	}

	//Check if kubeconfig is correct and kubectl commands are functional
	_, errClusterInfo := exec.Command("kubectl", "cluster-info").Output()
	if errClusterInfo != nil {
		return false, ErrClusterInfo
	}

	return true, nil
}

// Check if argocd cli is installed and the account has admin access
func verifyArgoCD() (bool, error) {
	// Check if argocd is installed
	_, err := exec.LookPath("argocd")
	if err != nil {
		return false, ErrArgoCD
	}

	//Check if argocd has access and auth-token is not expired
	_, errAcc := exec.Command("argocd", "account", "list").Output()
	if errAcc != nil {
		return false, ErrArgoCDAuthToken
	}
	return true, nil
}

// Check if git cli is installed
func verifyGit() (bool, error) {
	_, err := exec.LookPath("git")
	if err != nil {
		return false, ErrGit
	}
	return true, nil
}

// Verify if the arlon services are running
func verifyArlon() (bool, error) {
	ns := exec.Command("kubectl", "get", "ns")
	grepArlon := exec.Command("grep", "arlon")

	// Get ns's stdout and attach it to grep's stdin.
	pipe, _ := ns.StdoutPipe()
	defer pipe.Close()

	grepArlon.Stdin = pipe
	ns.Start()

	// Get the output of entire command
	_, errArlon := grepArlon.Output()
	if errArlon != nil {
		return false, ErrArlonNs
	}
	return true, nil
}

// Verify if capi services are running
func verifyCapi() (bool, error) {
	apiVersion := exec.Command("kubectl", "api-versions")
	grepCapiApiVersion := exec.Command("grep", "infrastructure.cluster.x-k8s.io")

	// Get apiVersion stdout and attach it to grepCapiApiVersion's stdin.
	pipe, _ := apiVersion.StdoutPipe()
	defer pipe.Close()

	grepCapiApiVersion.Stdin = pipe
	apiVersion.Start()

	// Get the output of entire command
	_, errCapi := grepCapiApiVersion.Output()
	if errCapi != nil {
		return false, ErrCapi
	}

	// Check for capa-system namespace
	errCapaCP := checkCapaCloudProvider()
	if errCapaCP != nil {
		//Check for capz-system incase capa-system is not present
		errCapaZ := checkCapzCloudProvider()
		if errCapaZ != nil {
			return false, errCapaCP
		}
	}

	return true, nil

}

// Function to check capa-system namespace
func checkCapaCloudProvider() error {
	ns := exec.Command("kubectl", "get", "ns")
	grepCapa := exec.Command("grep", "capa-system")

	// Get ns's stdout and attach it to grep's stdin.
	pipe, _ := ns.StdoutPipe()
	defer pipe.Close()

	grepCapa.Stdin = pipe
	ns.Start()

	// Get the output of entire command
	_, errCapa := grepCapa.Output()
	if errCapa != nil {
		return errCapa
	}
	return nil

}

// Function to check capz-system namespace
func checkCapzCloudProvider() error {
	ns := exec.Command("kubectl", "get", "ns")
	grepCapz := exec.Command("grep", "capz-system")

	// Get ns's stdout and attach it to grep's stdin.
	pipe, _ := ns.StdoutPipe()
	defer pipe.Close()

	grepCapz.Stdin = pipe
	ns.Start()

	// Get the output of entire command
	_, errCapz := grepCapz.Output()
	if errCapz != nil {
		return errCapz
	}
	return nil

}
