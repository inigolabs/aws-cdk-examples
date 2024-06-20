package main

import (
	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	ecs "github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	elb "github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	log "github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	sec "github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type FargateWithALBStackProps struct {
	cdk.StackProps
}

func NewFargateWithALBStack(scope constructs.Construct, id string, props *FargateWithALBStackProps) cdk.Stack {
	var sprops cdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	// Stack
	stack := cdk.NewStack(scope, &id, &sprops)

	// VPC
	vpc := ec2.NewVpc(stack, jsii.String("VPC"), &ec2.VpcProps{
		MaxAzs: jsii.Number(2),
	})

	// Cluster
	cluster := ecs.NewCluster(stack, jsii.String("Cluster"), &ecs.ClusterProps{
		Vpc: vpc,
	})

	// Task
	taskDef := ecs.NewFargateTaskDefinition(stack, jsii.String("Task"), &ecs.FargateTaskDefinitionProps{
		MemoryLimitMiB: jsii.Number(512),
		Cpu:            jsii.Number(256),
	})

	// Logging
	logGroup := log.NewLogGroup(stack, jsii.String("LogGroup"), &log.LogGroupProps{
		RemovalPolicy: cdk.RemovalPolicy_DESTROY,
		Retention:     log.RetentionDays_ONE_DAY,
	})

	// Starwars Container
	taskDef.AddContainer(jsii.String("Starwars"), &ecs.ContainerDefinitionOptions{
		Image:        ecs.ContainerImage_FromRegistry(jsii.String("inigohub/starwars:latest"), &ecs.RepositoryImageProps{}),
		PortMappings: &[]*ecs.PortMapping{{ContainerPort: jsii.Number(8888)}},
		Logging: ecs.LogDriver_AwsLogs(&ecs.AwsLogDriverProps{
			StreamPrefix: jsii.String("Starwars"),
			LogGroup:     logGroup,
		}),
		Environment: &map[string]*string{
			"SERVICE_LISTEN_PORT": jsii.String("8888"),
		},
	})

	// Service Token from Secrets Manager
	// To create one, use UI or CLI:
	// $ aws secretsmanager create-secret --name InigoServiceToken --secret-string '{"SERVICE_TOKEN":"..."}'
	secret := sec.Secret_FromSecretNameV2(stack, jsii.String("Secret"), jsii.String("InigoServiceToken"))

	// Sidecar Container
	taskDef.AddContainer(jsii.String("Sidecar"), &ecs.ContainerDefinitionOptions{
		Image:        ecs.ContainerImage_FromRegistry(jsii.String("inigohub/sidecar:latest"), &ecs.RepositoryImageProps{}),
		PortMappings: &[]*ecs.PortMapping{{ContainerPort: jsii.Number(80)}},
		Logging: ecs.LogDriver_AwsLogs(&ecs.AwsLogDriverProps{
			StreamPrefix: jsii.String("Sidecar"),
			LogGroup:     logGroup,
		}),
		Environment: &map[string]*string{
			"INIGO_ENABLE":      jsii.String("true"),
			"INIGO_LISTEN_PORT": jsii.String("80"),
			"INIGO_EGRESS_URL":  jsii.String("http://localhost:8888/query"),
		},
		Secrets: &map[string]ecs.Secret{
			"INIGO_SERVICE_TOKEN": ecs.Secret_FromSecretsManager(secret, jsii.String("SERVICE_TOKEN")),
		},
	})

	// Service
	service := ecs.NewFargateService(stack, jsii.String("Service"), &ecs.FargateServiceProps{
		Cluster:        cluster,
		TaskDefinition: taskDef,
	})

	// Load Balancer
	lb := elb.NewApplicationLoadBalancer(stack, jsii.String("LoadBalancer"), &elb.ApplicationLoadBalancerProps{
		Vpc:            vpc,
		InternetFacing: jsii.Bool(true),
	})

	// Public Listener
	listener := lb.AddListener(jsii.String("PublicListener"), &elb.BaseApplicationListenerProps{
		Port: jsii.Number(80),
		Open: jsii.Bool(true),
	})

	// Attach Starwars to Load Balancer
	listener.AddTargets(jsii.String("StarwarsSidecar"), &elb.AddApplicationTargetsProps{
		Port: jsii.Number(80),
		Targets: &[]elb.IApplicationLoadBalancerTarget{
			service.LoadBalancerTarget(&ecs.LoadBalancerTargetOptions{
				ContainerName: jsii.String("Sidecar"),
				ContainerPort: jsii.Number(80),
			}),
		},
	})

	// Output
	cdk.NewCfnOutput(stack, jsii.String("Address"), &cdk.CfnOutputProps{Value: lb.LoadBalancerDnsName()})

	return stack
}

func main() {
	app := cdk.NewApp(nil)

	NewFargateWithALBStack(app, "InigoStarwarsExample", &FargateWithALBStackProps{
		cdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *cdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &cdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &cdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
