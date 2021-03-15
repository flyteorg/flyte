# :construction: Instructions to deploy ALB Ingress controller

Follow instructions here to install ALB Ingress Controller: https://docs.aws.amazon.com/eks/latest/userguide/alb-ingress.html

Replace `alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:us-east-2:111111111111:certificate/e92fefd8-6197-4249-a524-431d611c9af6` in ingress.yaml and ingress_grpc.yaml with your own SSL cert (that you will create by following ALB Instructions above)
