package kubeapiauth

const (
	K8sAPIVersionV1       = "authentication.k8s.io/v1"
	K8sAPIVersionV1beta1  = "authentication.k8s.io/v1beta1"
	DefaultK8sAPIVersion  = K8sAPIVersionV1beta1
	DefaultAuthnKind      = "TokenReview"
	DefaultNamespace      = "cattle-system"
	DefaultListenHostPort = "127.0.0.1:6440"
)
