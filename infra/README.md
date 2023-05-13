# Welcome to your CDK Go project!

This is a blank project for CDK development with Go.

The `cdk.json` file tells the CDK toolkit how to execute your app.

## Useful commands

 * `cdk deploy`      deploy this stack to your default AWS account/region
 * `cdk diff`        compare deployed stack with current state
 * `cdk synth`       emits the synthesized CloudFormation template
 * `go test`         run unit tests

## New project runbook
### Before creating infra: 
create a new folder named `infra` inside your project dir
`cdk init app --language=go`

### To create infra:
1. Create an ECR repository manually & note its arn.
2. Create an S3 bucket manually & note its bucket-name.
3. Buy a new domain. Go to Route53 & create a hosted zone. 
   1. Add hosted-zone's CNAME records to the domain-hosting site (like hostinger / godaddy) DNS config.
4. You will get the zone-id & domain-name to create a certificate.
   1. This certificate is helpful to get HTTPS on our webservers.
5. Copy paste the code from infra.go, & put the constant values obtained above in the right places.
6. Create an ssh key locally. Go to aws and put the public key generated this way in KeyPair for EC2. 
   1. Also put the public-key text string in the NewCfnKeyPair. 
   2. You can now ssh using this key into your EC2 instances
   3. TODO: test if this still works give we added loadbalancer?
7. Any env variables should be added in parameter store. 
   1. Get the ARN / Name of the params & update the SSM policy statement to allow reading the right params.
   2. To read the env vars, you can refer to main.go of this code.
8. Create an admin role. Go to Github.com, create your new repo.
   1. Copy the github workflow from this codebase to your new repo. This will push changes to ECR when code is merged to master.
   2. Add the AWS_ACCESS_KEY_ID & AWS_ACCESS_KEY_SECRET varibles to the github actions secret.
9. Copy the rest of infra code as it is.
10. Once the infra is deployed, get the DNS address of the loadbalancer.
    1. Go to route53 hosted-zone & add do create-record. 
    2. decide a subdomain (ex. {api}.website.com) 
    3. select record type as A record
    4. Set routes traffic to & select your load balancer.
    5. You can now call this url (api.website.com) from your the internet (& browser) 

### To deploy infra
1. if never done: `cdk bootstrap`
2. `cdk deploy`

## To deploy code changes:
0. Login to AWS via this link: https://d-9067b5bbe4.awsapps.com/start and get keys
0. run `aws-cli configure` and put your AWS keys here
1. cd infra
2. export DEPLOY_COMMIT="the commit you want to deploy"
3. cdk deploy