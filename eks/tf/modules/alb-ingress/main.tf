# As required by https://docs.aws.amazon.com/eks/latest/userguide/alb-ingress.html
data "http" "alb_ingress_policy" {
  url = "https://raw.githubusercontent.com/kubernetes-sigs/aws-alb-ingress-controller/v1.1.4/docs/examples/iam-policy.json"

  request_headers = {
    Accept = "application/json"
  }
}

resource "aws_iam_policy" "k8s_alb_ingress_controller" {
  name   = "ALBIngressControllerIAMPolicy"
  path   = "/"
  policy = data.http.alb_ingress_policy.body
}

resource "aws_iam_role" "eks_alb_ingress_controller" {
  name = "eks-alb-ingress-controller"
  assume_role_policy = var.assume_role_policy_string
}

# Attach the policy to the flyte operator role
resource "aws_iam_role_policy_attachment" "eks_alb_attachment" {
  role       = aws_iam_role.eks_alb_ingress_controller.name
  policy_arn = aws_iam_policy.k8s_alb_ingress_controller.arn
}

data "aws_eks_cluster" "cluster" {
  name = var.cluster_id
}

data "aws_eks_cluster_auth" "cluster" {
  name = var.cluster_id
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.cluster.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority.0.data)
  token                  = data.aws_eks_cluster_auth.cluster.token
  load_config_file       = false
  version                = "~> 1.9"
}

resource "kubernetes_cluster_role" "alb_ingress_controller" {
  metadata {
    name = "alb-ingress-controller"

    labels = {
      "app.kubernetes.io/name" : "alb-ingress-controller"
    }
  }

  rule {
    api_groups = [
      "",
      "extensions",
    ]
    resources = [
      "configmaps",
      "endpoints",
      "events",
      "ingresses",
      "ingresses/status",
      "services",
    ]
    verbs = [
      "create",
      "get",
      "list",
      "update",
      "watch",
      "patch",
    ]
  }

  rule {
    api_groups = [
      "",
      "extensions",
    ]
    resources = [
      "nodes",
      "pods",
      "secrets",
      "services",
      "namespaces",
    ]
    verbs = [
      "get",
      "list",
      "watch",
    ]
  }
}

resource "kubernetes_service_account" "alb_ingress_controller" {
  metadata {
    name      = "alb-ingress-controller"
    namespace = "kube-system"

    labels = {
      "app.kubernetes.io/name" = "alb-ingress-controller"
    }

    annotations = {
      "eks.amazonaws.com/role-arn" = aws_iam_role.eks_alb_ingress_controller.arn
    }
  }
}

resource "kubernetes_cluster_role_binding" "alb_ingress_controller" {
  metadata {
    name = "alb-ingress-controller"

    labels = {
      "app.kubernetes.io/name" = "alb-ingress-controller"
    }
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.alb_ingress_controller.metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.alb_ingress_controller.metadata[0].name
    namespace = "kube-system"
  }
}

resource "kubernetes_deployment" "alb_ingress_controller" {
  metadata {
    name      = "alb-ingress-controller"
    namespace = "kube-system"

    labels = {
      "app.kubernetes.io/name" = "alb-ingress-controller"
    }
  }

  spec {
    selector {
      match_labels = {
        "app.kubernetes.io/name" = "alb-ingress-controller"
      }
    }

    template {
      metadata {
        name      = "alb-ingress-controller"
        namespace = "kube-system"

        labels = {
          "app.kubernetes.io/name" = "alb-ingress-controller"
        }
      }

      spec {
        container {
          name  = "alb-ingress-controller"
          image = "docker.io/amazon/aws-alb-ingress-controller:v1.1.4"
          args = [
            "--ingress-class=alb",
            "--cluster-name=${var.eks_cluster_name}",
            "--aws-vpc-id=${var.eks_vpc_id}",
            "--aws-region=${var.region}",
            "--feature-gates=waf=false",
          ]
        }

        service_account_name            = kubernetes_service_account.alb_ingress_controller.metadata[0].name
        automount_service_account_token = true

        node_selector = {
          "beta.kubernetes.io/os" = "linux"
        }
      }
    }
  }
}


