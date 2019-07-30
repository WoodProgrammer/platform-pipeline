package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/packer"
	"github.com/gruntwork-io/terratest/modules/test-structure"
)


func TestTerraformPackerExample(t *testing.T) {
	t.Parallel()

	awsRegion := "eu-central-1"

	workingDir := "./module"

	test_structure.RunTestStage(t, "build_ami", func() {
		buildAmi(t, awsRegion, workingDir)
	})

}

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
