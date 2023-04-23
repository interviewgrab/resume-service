package main

import (
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	elbv2 "github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"log"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type InfraStackProps struct {
	awscdk.StackProps
}

func NewInfraStack(scope constructs.Construct, id string, props *InfraStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)
	latestMergedCommit := os.Getenv("DEPLOY_COMMIT")
	if latestMergedCommit == "" {
		log.Fatal("Cannot find commit to deploy")
	}
	fmt.Printf("Deploying commit: %s\n", latestMergedCommit)

	// ssh key
	key := awsec2.NewCfnKeyPair(stack, jsii.String("CDKKeyPair"), &awsec2.CfnKeyPairProps{
		KeyName:           jsii.String("cdk-test-keypair"),
		PublicKeyMaterial: jsii.String("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCJ/jxSh28sbc9sckhNrP22BcbA6wgWSLAQKRoZ8WmeRAf8IYsysRKgnqamJdbulkUmUqrNbVk/t9swrTK9Tdp00/gO1sKKshlHwVTlg9YdSC4Uk21k/radyQKuuzCqyEbZFG1HAZVt4Z2Q9zrmg/qnqFckbRb4zeyeSAjFn0CDExjuS4upC0c+HoIF9lbvbodsJQVSP25qJdVbanm7aN/VZllz6ivNBrFJ1tXTUgOR4tyvhRX6R9B4uJP1gl0DZ5i81ElHeb5M4sYGxjcsONAY+TFyYVdHEf6bo0zZR5IUhnn6S1QAEhReG1Wxi3KuNHHVnDdlwWXlNdUWBJaXW1ob"),
	})

	// service ECR
	repo := awsecr.Repository_FromRepositoryName(stack, jsii.String("resume-service-ecr"), jsii.String("resume-service"))

	// service S3
	s3 := awss3.Bucket_FromBucketName(stack, jsii.String("resume-service-bucket"), jsii.String("resume-service-filestore"))

	// hosted zone
	zone := awsroute53.HostedZone_FromHostedZoneId(stack, jsii.String("resume-service-zone"), jsii.String("Z07921103DN76C22NES28"))

	// backend api certificate for subdomain
	certificate := awscertificatemanager.NewCertificate(stack, jsii.String("certificate"), &awscertificatemanager.CertificateProps{
		DomainName: jsii.String("api.interviewgrab.tech"),
		Validation: awscertificatemanager.CertificateValidation_FromDns(zone),
	})

	// common infra
	vpc := awsec2.NewVpc(stack, jsii.String("resume-service-vpc"), &awsec2.VpcProps{
		MaxAzs:             jsii.Number(1),
		NatGateways:        jsii.Number(0),
		EnableDnsSupport:   jsii.Bool(true),
		EnableDnsHostnames: jsii.Bool(true),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				CidrMask:   jsii.Number(24),
				Name:       jsii.String("ingress"),
				SubnetType: awsec2.SubnetType_PUBLIC,
			},
		},
	})

	sg := awsec2.NewSecurityGroup(stack, jsii.String("resume-service-sg"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		AllowAllOutbound: jsii.Bool(true),
	})

	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(80)), jsii.String("Allow http access"), jsii.Bool(false))
	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(443)), jsii.String("Allow https access"), jsii.Bool(false))
	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(22)), jsii.String("Allow ssh access"), jsii.Bool(false))

	// define execution role for ECS. Allow reading from s3
	executionRolePolicy := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: &[]*string{
			jsii.String("ec2:DescribeTags"),
			jsii.String("ecs:CreateCluster"),
			jsii.String("ecs:DeregisterContainerInstance"),
			jsii.String("ecs:DiscoverPollEndpoint"),
			jsii.String("ecs:Poll"),
			jsii.String("ecs:RegisterContainerInstance"),
			jsii.String("ecs:StartTelemetrySession"),
			jsii.String("ecs:UpdateContainerInstancesState"),
			jsii.String("ecs:Submit*"),
			jsii.String("ecr:GetAuthorizationToken"),
			jsii.String("ecr:BatchCheckLayerAvailability"),
			jsii.String("ecr:GetDownloadUrlForLayer"),
			jsii.String("ecr:BatchGetImage"),
			jsii.String("ec2-instance-connect:SendSSHPublicKey"),
			jsii.String("logs:CreateLogStream"),
			jsii.String("logs:PutLogEvents"),
			jsii.String("ecs:TagResource"),
		},
		Resources: &[]*string{jsii.String("*")},
	})

	executionRole := awsiam.NewRole(stack, jsii.String("resume-service-role"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ecs-tasks.amazonaws.com"), nil),
	})
	executionRole.AddToPolicy(executionRolePolicy)
	s3.GrantReadWrite(executionRole, nil)

	// allow aws ecs to read params from AWS system manager param store
	builder := NewParamBuilder(*stack.Account(), *stack.Region())

	ssmPolicyStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: &[]*string{
			jsii.String("ssm:GetParameter"),
		},
		Resources: &[]*string{
			builder.Arn("MONGO_URI"),
			builder.Arn("OPENAI_API_KEY"),
			builder.Arn("SENDER_EMAIL"),
			builder.Arn("SENDER_PASS"),
		},
	})
	executionRole.AddToPolicy(ssmPolicyStatement)

	// start defining cluster
	cluster := awsecs.NewCluster(stack, jsii.String("resume-service-cluster"), &awsecs.ClusterProps{
		Vpc: vpc,
	})
	capacity := cluster.AddCapacity(jsii.String("resume-service-capacity"), &awsecs.AddCapacityOptions{
		InstanceType:    awsec2.InstanceType_Of(awsec2.InstanceClass_T4G, awsec2.InstanceSize_SMALL),
		MachineImage:    awsecs.EcsOptimizedImage_AmazonLinux2(awsecs.AmiHardwareType_ARM, &awsecs.EcsOptimizedImageOptions{CachedInContext: jsii.Bool(true)}),
		DesiredCapacity: jsii.Number(1),
		KeyName:         key.KeyName(),
	})
	capacity.AddSecurityGroup(sg)
	// TODO separate task role
	taskDef := awsecs.NewEc2TaskDefinition(stack, jsii.String("resume-service-task"), &awsecs.Ec2TaskDefinitionProps{
		NetworkMode:   awsecs.NetworkMode_BRIDGE,
		ExecutionRole: executionRole,
		TaskRole:      executionRole,
	})
	taskDef.AddContainer(jsii.String("resume-service-container"), &awsecs.ContainerDefinitionOptions{
		Image:          awsecs.ContainerImage_FromEcrRepository(repo, jsii.String(latestMergedCommit)),
		MemoryLimitMiB: jsii.Number(512),
		Cpu:            jsii.Number(1),
		PortMappings: &[]*awsecs.PortMapping{
			{
				ContainerPort: jsii.Number(8080),
				HostPort:      jsii.Number(443),
			},
		},
	})

	service := awsecs.NewEc2Service(stack, jsii.String("resume-service-ec2service"), &awsecs.Ec2ServiceProps{
		Cluster:        cluster,
		TaskDefinition: taskDef,
	})

	// setup loadbalancer for the ECS service
	loadBalancer := elbv2.NewNetworkLoadBalancer(stack, jsii.String("loadBalancer"), &elbv2.NetworkLoadBalancerProps{
		Vpc:            vpc,
		InternetFacing: jsii.Bool(true),
	})
	tgtProps := elbv2.AddNetworkTargetsProps{
		Protocol: elbv2.Protocol_TCP,
		Port:     jsii.Number(443),
		Targets:  &[]elbv2.INetworkLoadBalancerTarget{service},
	}
	loadBalancer.AddListener(jsii.String("resume-service-http-listener"), &elbv2.BaseNetworkListenerProps{
		Port: jsii.Number(80),
	}).AddTargets(jsii.String("resume-service-http-tgt"), &tgtProps)
	loadBalancer.AddListener(jsii.String("resume-service-https-listener"), &elbv2.BaseNetworkListenerProps{
		Port:         jsii.Number(443),
		Certificates: &[]elbv2.IListenerCertificate{certificate},
	}).AddTargets(jsii.String("resume-service-https-tgt"), &tgtProps)

	// expose load balancer via route53 A record for api.interviewgrab.tech
	awsroute53.NewARecord(stack, jsii.String("interviewgrab-api-ARecord"), &awsroute53.ARecordProps{
		Zone:   zone,
		Target: awsroute53.RecordTarget_FromAlias(awsroute53targets.NewLoadBalancerTarget(loadBalancer)),
	})
	return stack
}

type ParamBuilder struct {
	account string
	region  string
}

func NewParamBuilder(account string, region string) ParamBuilder {
	return ParamBuilder{account: account, region: region}
}

func (b *ParamBuilder) Arn(param string) *string {
	return jsii.String(fmt.Sprintf("arn:aws:ssm:%s:%s:parameter/%s", b.region, b.account, param))
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewInfraStack(app, "InfraStack", &InfraStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
