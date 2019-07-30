package test

import (
	"fmt"
	"testing"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"math/rand"
	"strings"
	"strconv"
	"github.com/stretchr/testify/assert"

)

func TestTerraformAwsS3Example(t *testing.T) {
	
	t.Parallel()

	awsRegion := "eu-central-1"

	terraformOptions := &terraform.Options{
		TerraformDir: "./module",
	}

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	bucketID := terraform.Output(t, terraformOptions, "bucket_id")

	aws.AssertS3BucketPolicyExists(t, awsRegion, bucketID)

}

func TestBucketPolicy(t *testing.T) {

	state := false

	terraformOptions := &terraform.Options{
		TerraformDir: "./module",
	}

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	expectedFileCount := rand.Intn(10000)
	awsRegion := "eu-central-1"

	for i := 1; i <= expectedFileCount; i++ {
		key := fmt.Sprintf("test-%s", strconv.Itoa(i))
		body := strings.NewReader("This is the body")
		bucket_name := "devops-turkey-demo-bucket"

		params := &s3manager.UploadInput{
			Bucket: &bucket_name,
			Key:    &key,
			Body:   body,
		}

		uploader := aws.NewS3Uploader(t, awsRegion)

		_, err := uploader.Upload(params)
		if err != nil {
			state = true
		}

	}

	assert.Equal(t, false, state)

}