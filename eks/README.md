
Steps:
- You need to make sure kubectl and aws-cli are installed and working
- Aws-cli permissions must be referenced and terraform must be installed and running
- Run Terraform files. (There seems to be a race condition in one of the IAM role creation steps - you may need to run it twice.)
- Copy or update the kubectl config file and switch to that context.
- Create the webhook
  - Create ECR repo for the webhook
  - Build the image and push
  - Run the make cluster-up command with the right image
- Create the example 2048 game on the EKS IAM page linked above. Keep in mind that even after an address shows up in the ingress, it may take a while to provision.
- Delete the game
- Create the spare datacatalog reference in the db.
- Follow the [Installation portion](https://github.com/aws/amazon-eks-pod-identity-webhook/blob/95808cffe6d801822dae122f2f2c87a258d70bb8/README.md#installation) of the webhook readme.  You will need to make sure to use your own AWS account number, and will also need to build your own image and upload it to your ECR, which will probably require you to create that repository in your ECR.
- Go through all the overlays in the `kustomize/overlays/eks` folder and make sure all the service accounts and RDS addresses reference yours. (Do a grep for `111222333456` and `456123e6ivib`).
- Install Flyte with `kubectl apply -f deployment/eks/flyte_generated.yaml`

This is the webhook used to inject IAM role credentials into pods.
https://github.com/aws/amazon-eks-pod-identity-webhook

This is how you get pods to use the proper roles. (This is the KIAM replacement.)
https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts-technical-overview.html
The implementation of these steps is done for you in the `alb-ingress` submodule.


