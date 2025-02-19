package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/spf13/cobra"
)

var verbose bool


var rootCmd = &cobra.Command{
	Use:   "haas-cli",
	Short: "HAAS CLI tool",
	Long:  `A CLI tool for automating HAAS related tasks`,
}

var startCmd = &cobra.Command{
	Use:   "attach-account",
	Short: "Attach an account to the HAAS process",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runAttachAccountCheck(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var restartInstanceCmd = &cobra.Command{
	Use:   "restart-instance [vm-hash] [secret]",
	Short: "Restart a specific VM instance",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runRestartInstance(args[0], args[1]); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var createAndStartInstanceCmd = &cobra.Command{
	Use:   "create-and-start",
	Short: "Create and start a new VM instance",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runCreateAndStartInstance(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop [vm-hash]",
	Short: "Stop a specific VM instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runStopInstance(args[0]); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [vm-hash]",
	Short: "Stop and delete a specific VM instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDeleteInstance(args[0]); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var (
	instanceName string
	vmSecret     string
	// Configuration defaults
	rootfsHash = "eca1a77a489ef53ecb6f6febad0735103c85b6e0176ff707390b5fee939ea824"
	rootfsSize = "10240"
	vcpus      = "1"
	memorySize = "2048"
)
var privateKey string

func init() {
	startCmd.Flags().StringVarP(&privateKey, "private-key", "k", "", "Private key for the account (required)")
	startCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	startCmd.MarkFlagRequired("private-key")

	// Add verbose flag to start-instance command
	restartInstanceCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")

	// Add flags for create-and-start command
	createAndStartInstanceCmd.Flags().StringVarP(&instanceName, "name", "n", "", "Name for the VM instance (required)")
	createAndStartInstanceCmd.Flags().StringVarP(&vmSecret, "secret", "s", "", "Secret for the VM instance (required)")
	createAndStartInstanceCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	createAndStartInstanceCmd.MarkFlagRequired("name")
	createAndStartInstanceCmd.MarkFlagRequired("secret")

	// Add verbose flag to stop command
	stopCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")

	// Add verbose flag to delete command
	deleteCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(restartInstanceCmd)
	rootCmd.AddCommand(createAndStartInstanceCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(deleteCmd)
}

func checkAlephInstallation() error {
	cmd := exec.Command("aleph", "--help")
	if err := cmd.Run(); err != nil {
		fmt.Println("Aleph is not installed. Please run 'make' to install all dependencies.")
		return fmt.Errorf("aleph not installed")
	}
	return nil
}

func checkSevctlInstallation() error {
	cmd := exec.Command("sevctl", "-V")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sevctl is not installed. Please install sevctl before running init-session")
	}
	return nil
}

func runAlephCommand(privateKey string) error {
	session := sh.NewSession()

	// Control command output visibility
	if !verbose {
		session.SetEnv("PYTHONUNBUFFERED", "1") // Ensure Python output isn't buffered
		session.Stdout = nil
		session.Stderr = nil
	}

	// First create the account
	out, err := session.
		Command("aleph", "account", "create", "--chain", "SOL", "--private-key", privateKey).
		SetInput(fmt.Sprintf("%s-%s\n", "robotter", time.Now().Format("20060102150405"))).
		Output()
	if err != nil {
		return fmt.Errorf("failed to create account: %v", err)
	}

	// Verify account creation succeeded
	if !strings.Contains(string(out), "Private key") {
		return fmt.Errorf("account creation may have failed, unexpected output: %s", string(out))
	}

	if verbose {
		fmt.Println(string(out))
	}

	// Then list files
	if verbose {
		fmt.Println("\nListing files...")
	}
	out, err = session.Command("aleph", "file", "list").Output()
	if err != nil {
		return fmt.Errorf("failed to list files: %v", err)
	}

	// Verify file list contains expected data
	fileOutput := string(out)
	if verbose {
		fmt.Println(fileOutput)
	}

	if !strings.Contains(fileOutput, "Address:") || !strings.Contains(fileOutput, "Total Size:") {
		return fmt.Errorf("file list command didn't return expected output format")
	}

	return nil
}

func runAttachAccountCheck() error {
	// Check if aleph is installed
	if err := checkAlephInstallation(); err != nil {
		return err
	}

	// Run aleph command
	if err := runAlephCommand(privateKey); err != nil {
		return err
	}

	if err := checkSevctlInstallation(); err != nil {
		return err
	}

	return nil
}

func runRestartInstance(vmHash string, secret string) error {
	session := sh.NewSession()
	if verbose {
		session.Stdout = os.Stdout
		session.Stderr = os.Stderr
	}

	// Stop the VM
	if verbose {
		fmt.Println("Stopping VM...")
	}
	if err := session.Command("aleph", "instance", "stop", vmHash).Run(); err != nil {
		return fmt.Errorf("failed to stop VM: %v", err)
	}

	// Initialize confidential session
	time.Sleep(5 * time.Second)
	if verbose {
		fmt.Println("Initializing confidential session...")
	}

	cmd := exec.Command("aleph", "instance", "confidential", "--vm-secret", secret, vmHash)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	if verbose {
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start init command: %v", err)
	}

	// Read output to check for override prompt
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if verbose {
				fmt.Println(line)
			}
			if strings.Contains(line, "are you sure you want to override") {
				time.Sleep(1 * time.Second)
				io.WriteString(stdin, "y\n")
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("init command failed: %v", err)
	}

	fmt.Println("Instance started successfully")

	return nil
}

func runCreateAndStartInstance() error {
	if err := checkAlephInstallation(); err != nil {
		return err
	}

	if err := checkSevctlInstallation(); err != nil {
		return err
	}

	cmd := exec.Command("aleph", "instance", "confidential",
		"--vm-secret", vmSecret,
		"--payment-type", "hold",
		"--name", instanceName,
		"--rootfs", rootfsHash,
		"--skip-volume",
		"--rootfs-size", rootfsSize,
		"--vcpus", vcpus,
		"--memory", memorySize,
	)

	// Create pipes for stdin and stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	combinedOutput, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	cmd.Stderr = cmd.Stdout
	writer := bufio.NewWriter(stdin)

	if verbose {
		fmt.Println("Debug: Starting command...")
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	buf := make([]byte, 1)
	line := ""
	var vmHash string

	for {
		n, err := combinedOutput.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read error: %v", err)
		}

		if n > 0 {
			if verbose {
				fmt.Printf("%s", string(buf[:n]))
			}
			line += string(buf[:n])

			// Handle node selection after loading complete
			if strings.Contains(line, "Fetching data of") && strings.Contains(line, "100%") {
				if verbose {
					fmt.Println("\nDebug: Node list fully loaded")
				}
				time.Sleep(500 * time.Millisecond)
				writer.WriteString("\r\n")
				writer.Flush()
				line = ""
			}

			if strings.Contains(line, "Deploy on this node ?") {
				if verbose {
					fmt.Println("\nDebug: Found deployment prompt")
				}
				time.Sleep(2 * time.Second)
				writer.WriteString("y\n")
				writer.Flush()
				line = ""
			}

			// Extract VM hash from success message
			if strings.Contains(line, "Your instance") && strings.Contains(line, "has been deployed on aleph.im") {
				if verbose {
					fmt.Println("\nDebug: Instance successfully deployed")
				}
				// Extract VM hash from the line
				parts := strings.Split(line, " ")
				for i, part := range parts {
					if part == "instance" && i+1 < len(parts) {
						vmHash = parts[i+1]
						break
					}
				}
				line = ""
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %v", err)
	}

	if vmHash == "" {
		return fmt.Errorf("failed to get VM hash from output")
	}

	fmt.Printf("Instance created with hash: %s\n", vmHash)
	fmt.Println("Starting instance... Done")

	// Wait a moment before starting the instance
	time.Sleep(2 * time.Second)
	return nil
}

func runStopInstance(vmHash string) error {
	session := sh.NewSession()
	if verbose {
		session.Stdout = os.Stdout
		session.Stderr = os.Stderr
	}

	if verbose {
		fmt.Println("Stopping VM...")
	}
	if err := session.Command("aleph", "instance", "stop", vmHash).Run(); err != nil {
		return fmt.Errorf("failed to stop VM: %v", err)
	}

	fmt.Println("Instance stopped successfully")
	return nil
}

func runDeleteInstance(vmHash string) error {
	// First stop the instance
	if err := runStopInstance(vmHash); err != nil {
		return err
	}

	session := sh.NewSession()
	if verbose {
		session.Stdout = os.Stdout
		session.Stderr = os.Stderr
	}

	if verbose {
		fmt.Println("Deleting VM...")
	}
	if err := session.Command("aleph", "instance", "delete", vmHash).Run(); err != nil {
		return fmt.Errorf("failed to delete VM: %v", err)
	}

	fmt.Println("Instance deleted successfully")
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
