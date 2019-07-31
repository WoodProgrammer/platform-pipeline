package test

import (
	
	"testing"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/packer"
	"github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/random"

	"time"
	"fmt"
	"strings"
)


func TestTerraformPackerExample(t *testing.T) {
	t.Parallel()

	awsRegion := "eu-central-1"

	workingDir := "./module"

	test_structure.RunTestStage(t, "build_ami", func() {
		buildAmi(t, awsRegion, workingDir)
	})

	TestTerraformSshExample(t)

}

// Build the AMI
func buildAmi(t *testing.T, awsRegion string, workingDir string) {
	packerOptions := &packer.Options{
		// The path to where the Packer template is located
		Template: "./module/main.json",
	}

	// Save the Packer Options so future test stages can use them
	test_structure.SavePackerOptions(t, workingDir, packerOptions)

	// Build the AMI
	amiId := packer.BuildAmi(t, packerOptions)

	// Save the AMI ID so future test stages can use them
	test_structure.SaveAmiId(t, workingDir, amiId)
}


//rm -f /home/ec2-user/.ssh/authorized_keys


func TestTerraformSshExample(t *testing.T) {

	packerWorkingDir := "./module"

	exampleFolder := test_structure.CopyTerraformFolderToTemp(t, "./module/", "ec2")

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, exampleFolder)
		terraform.Destroy(t, terraformOptions)

		keyPair := test_structure.LoadEc2KeyPair(t, exampleFolder)
		aws.DeleteEC2KeyPair(t, keyPair)
	})

	// Deploy the example
	test_structure.RunTestStage(t, "setup", func() {
		terraformOptions, keyPair := configureTerraformOptions(t, exampleFolder, packerWorkingDir)

		// Save the options and key pair so later test stages can use them
		test_structure.SaveTerraformOptions(t, exampleFolder, terraformOptions)
		test_structure.SaveEc2KeyPair(t, exampleFolder, keyPair)

		// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
		terraform.InitAndApply(t, terraformOptions)
	})

	// Make sure we can SSH to the public Instance directly from the public Internet and the private Instance by using
	// the public Instance as a jump host
	test_structure.RunTestStage(t, "validate", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, exampleFolder)
		keyPair := test_structure.LoadEc2KeyPair(t, exampleFolder)

		testSSHToPublicHost(t, terraformOptions, keyPair)

	})

}


func configureTerraformOptions(t *testing.T, exampleFolder string, packerWorkingDir string) (*terraform.Options, *aws.Ec2Keypair) {
	// A unique ID we can use to namespace resources so we don't clash with anything already in the AWS account or
	// tests running in parallel
	uniqueID := random.UniqueId()

	// Give this EC2 Instance and other resources in the Terraform code a name with a unique ID so it doesn't clash
	// with anything else in the AWS account.
	instanceName := fmt.Sprintf("terratest-ssh-example-%s", uniqueID)

	// Pick a random AWS region to test in. This helps ensure your code works in all regions.
	awsRegion := "eu-central-1"

	// Create an EC2 KeyPair that we can use for SSH access
	keyPairName := fmt.Sprintf("terratest-ssh-example-%s", uniqueID)
	keyPair := aws.CreateAndImportEC2KeyPair(t, awsRegion, keyPairName)

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: exampleFolder,

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"aws_region":     awsRegion,
			"instance_name":  instanceName,
			"key_pair_name":  keyPairName,
			"ami_id"		: test_structure.LoadAmiId(t, packerWorkingDir),
		},

	}

	return terraformOptions, keyPair
}


func testSSHToPublicHost(t *testing.T, terraformOptions *terraform.Options, keyPair *aws.Ec2Keypair) {
	// Run `terraform output` to get the value of an output variable
	publicInstanceIP := terraform.Output(t, terraformOptions, "public_instance_ip")

	// We're going to try to SSH to the instance IP, using the Key Pair we created earlier, and the user "ubuntu",
	// as we know the Instance is running an Ubuntu AMI that has such a user
	publicHost := ssh.Host{
		Hostname:    publicInstanceIP,
		SshKeyPair:  keyPair.KeyPair,
		SshUserName: "ec2-user",
	}

	// It can take a minute or so for the Instance to boot up, so retry a few times
	maxRetries := 2
	timeBetweenRetries := 5 * time.Second
	description := fmt.Sprintf("SSH to public host %s", publicInstanceIP)

	// Run a simple echo command on the server
	expectedText := ""
	command := fmt.Sprintf("test -f /home/ec2-user/.ssh/authorized_keys && echo exist" )

	// Verify that we can SSH to the Instance and run commands
	retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
		actualText, err := ssh.CheckSshCommandE(t, publicHost, command)

		if err != nil {
			return "", err
		}

		if strings.TrimSpace(actualText) != expectedText {
			return "", fmt.Errorf("Expected SSH command to return '%s' but got '%s'", expectedText, actualText)
		}

		return "", nil
	})


}


// Delete the AMI
func deleteAmi(t *testing.T, awsRegion string, workingDir string) {
	// Load the AMI ID and Packer Options saved by the earlier build_ami stage
	amiId := test_structure.LoadAmiId(t, workingDir)

	aws.DeleteAmi(t, awsRegion, amiId)
}
