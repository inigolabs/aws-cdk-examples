import ec2 = require('aws-cdk-lib/aws-ec2');
import ecs = require('aws-cdk-lib/aws-ecs');
import log = require('aws-cdk-lib/aws-logs');
import elb = require('aws-cdk-lib/aws-elasticloadbalancingv2');
import sec = require('aws-cdk-lib/aws-secretsmanager');
import cdk = require('aws-cdk-lib');

class InigoStarwarsExample extends cdk.Stack {
  constructor(scope: cdk.App, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // VPC
    const vpc = new ec2.Vpc(this, 'VPC', { 
      maxAzs: 2,
    });

    // Cluster
    const cluster = new ecs.Cluster(this, 'Cluster', { 
      vpc,
    });

    // Task
    const task = new ecs.FargateTaskDefinition(this, "Task", {
      memoryLimitMiB: 512,
      cpu: 256,
    });
    
    // Starwars container
    task.addContainer("Starwars", {
      image: ecs.ContainerImage.fromRegistry("inigohub/starwars:latest"),
      portMappings: [{ containerPort: 8888 }],
      logging: new ecs.AwsLogDriver({
        streamPrefix: "Starwars",
        logRetention: log.RetentionDays.ONE_DAY,
      }),
      environment: {
        "SERVICE_LISTEN_PORT": "8888",
      },
    });

    // Service Token from Secrets Manager
    // To create one, use UI or CLI:
    // $ aws secretsmanager create-secret --name InigoServiceToken --secret-string '{"SERVICE_TOKEN":"..."}'
    const secret = sec.Secret.fromSecretNameV2(this, "Secret", "InigoServiceToken");

    // Sidecar container
    task.addContainer("Sidecar", {
      image: ecs.ContainerImage.fromRegistry("inigohub/sidecar:latest"),
      portMappings: [{ containerPort: 80 }],
      logging: new ecs.AwsLogDriver({
        streamPrefix: "Sidecar",
        logRetention: log.RetentionDays.ONE_DAY,
      }),
      environment: {
        "INIGO_ENABLE":      "true",
        "INIGO_LISTEN_PORT": "80",
        "INIGO_EGRESS_URL":  "http://localhost:8888/query",
      },
      secrets: {
        "INIGO_SERVICE_TOKEN": ecs.Secret.fromSecretsManager(secret, "SERVICE_TOKEN"),
      },
    });

    // Service
    const service = new ecs.FargateService(this, "Service", {
      cluster,
      taskDefinition: task,
    });

    // Load Balancer
    const lb = new elb.ApplicationLoadBalancer(this, "LoadBalancer", {
      vpc,
      internetFacing: true,
    });
    
    // Public Listener
    const listener = lb.addListener("PublicListener", {
      port: 80,
      open: true,
    });

    // Attach Starwars to Load Balancer
    listener.addTargets("StarwarsSidecar", {
      port: 80,
      targets: [
        service.loadBalancerTarget({
          containerName: "Sidecar",
          containerPort: 80,
        }),
      ],
    });

    // Output
    new cdk.CfnOutput(this, "Address", { value: lb.loadBalancerDnsName });
  }
}

const app = new cdk.App();

new InigoStarwarsExample(app, 'InigoStarwarsExample');

app.synth();