# aws-cdk-examples

> **Stable examples. Should successfully build out of the box.**
>
> These examples are built on Construct Libraries marked as "Stable" and do not have any infrastructure prerequisites to build.

## Overview

This repository contains a collection of samples demonstrating the use of the AWS Cloud Development Kit (CDK) to configure and deploy services related to the Inigo project.

The AWS CDK is a powerful framework for defining cloud infrastructure in code and deploying it through AWS CloudFormation. It provides high-level components that preconfigure cloud resources with proven defaults, so you can build cloud applications without needing to be an expert in cloud architecture.

In this repository, you'll find examples of how to use the AWS CDK to deploy Inigo-related services. Inigo is a project that focuses on providing robust, scalable solutions for modern cloud-based applications. These samples will guide you through the process of deploying these services, providing a hands-on approach to learning.

## Deploying

- Navigate to desired directory.
- Authenticate to your AWS account from your terminal.
- Create a Service Token either via Inigo User Interface or its CLI.
- Create a Secret using the Service Token, either via AWS User Interface or using the AWS CLI by executing:

    ``` bash
    aws secretsmanager create-secret --name InigoServiceToken --secret-string '{"SERVICE_TOKEN":"INSERT SERVICE TOKEN HERE"}'
    ```

- `cdk bootstrap` to deploy the CDK toolkit stack into an AWS environment.
- `cdk deploy` to deploy the stack to the AWS account you're authenticated to.

## Testing

CDK will output the ALB endpoint after a successful deployment. Copy / paste it into the browser to check that the application is running.

- e.g. `http://InigoS-LoadB-xxxx-xxxxx.us-west-2.elb.amazonaws.com`
- _Note: Be sure to prefix the ALB endpoint with `http://` as your browser may initially force `https`, which will not work as we are not issuing or installing an SSL certificate in this sample._
