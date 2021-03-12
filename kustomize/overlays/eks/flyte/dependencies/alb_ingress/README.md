# :construction: Instructions to deploy ALB Ingress controller

Follow instructions here to install ALB Ingress Controller: https://docs.aws.amazon.com/eks/latest/userguide/alb-ingress.html

Replace `alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:us-east-2:590375264460:certificate/e2f04275-2dff-4118-a493-ed3ec8f41605` in ingress.yaml and ingress_grpc.yaml with your own SSL cert (that you will create by following ALB Instructions above)
