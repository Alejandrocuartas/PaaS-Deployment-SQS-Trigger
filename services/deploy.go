package services

import (
	"PaaS-deployment-sqs-trigger/models"
	"PaaS-deployment-sqs-trigger/repositories"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/jinzhu/gorm"
)

func DeployApp(appId string) error {

	app, e := repositories.GetAppByUuid(appId)
	if e != nil {
		log.Printf("Error getting app: %v", e)
		if gorm.IsRecordNotFoundError(e) {
			return nil
		}
		return e
	}
	log.Println("app", app)

	_, e = repositories.DeployApp(app.UUID.String())
	if e != nil {
		log.Printf("Error deploying app: %v", e)
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	time.Sleep(time.Second * 20)

	sess, e := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if e != nil {
		log.Printf("Error creating session: %v", e)
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	svc := ecs.New(sess)

	listTasksInput := &ecs.ListTasksInput{
		Cluster: aws.String(app.UUID.String()),
	}
	tasksListResult, e := svc.ListTasks(listTasksInput)
	if e != nil {
		log.Printf("Error listing tasks: %v", e)
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	if len(tasksListResult.TaskArns) == 0 {
		log.Printf("No tasks found for app %s", app.UUID.String())
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	taskArn := tasksListResult.TaskArns[0]

	if taskArn == nil {
		log.Printf("No task found for app %s", app.UUID.String())
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	taskArnParts := strings.Split(*taskArn, "/")
	taskId := taskArnParts[len(taskArnParts)-1]

	describeTaskInput := &ecs.DescribeTasksInput{
		Cluster: aws.String(app.UUID.String()),
		Tasks:   []*string{aws.String(taskId)},
	}
	describeTaskResult, e := svc.DescribeTasks(describeTaskInput)
	if e != nil {
		log.Printf("Failed to describe task: %v", e)
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	if len(describeTaskResult.Failures) > 0 {
		log.Printf("Task %s failed to start", taskId)
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	if len(describeTaskResult.Tasks) == 0 {
		log.Printf("No task found for app %s", app.UUID.String())
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	task := describeTaskResult.Tasks[0]

	if task == nil {
		log.Printf("No task found for app %s", app.UUID.String())
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	taskAttachments := task.Attachments

	if len(taskAttachments) == 0 {
		log.Printf("No task attachments found for app %s", app.UUID.String())
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	eni := ""
	for _, attachmentDetail := range taskAttachments[0].Details {
		if attachmentDetail.Name != nil {
			if *attachmentDetail.Name == "networkInterfaceId" {
				eni = *attachmentDetail.Value
			}
		}
	}

	if eni == "" {
		log.Printf("No network interface found for app %s", app.UUID.String())
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	ec2Client := ec2.New(sess)

	describeNetworkInterfacesInput := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []*string{aws.String(eni)},
	}
	describeNetworkInterfacesResult, e := ec2Client.DescribeNetworkInterfaces(describeNetworkInterfacesInput)
	if e != nil {
		log.Printf("Failed to describe network interfaces: %v", e)
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	if len(describeNetworkInterfacesResult.NetworkInterfaces) == 0 {
		log.Printf("No network interfaces found for app %s", app.UUID.String())
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	ni := describeNetworkInterfacesResult.NetworkInterfaces[0]

	if ni == nil {
		log.Printf("No network interface found for app %s", app.UUID.String())
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	niAsociation := ni.Association

	if niAsociation == nil {
		log.Printf("No network interface association found for app %s", app.UUID.String())
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	publicDnsName := niAsociation.PublicDnsName

	if publicDnsName == nil {
		log.Printf("No public dns name found for app %s", app.UUID.String())
		app.Status = models.AppStatusInactive
		repositories.UpdateApp(app)
		return nil
	}

	url := fmt.Sprintf("http://%s", *publicDnsName)
	app.DeployUrl = sql.NullString{Valid: true, String: url}
	app.Status = models.AppStatusActive

	return repositories.UpdateApp(app)
}
