package types

type V1AuthnRequestSpec struct {
	Token string `json:"token"`
}

type V1AuthnRequest struct {
	APIVersion string             `json:"apiVersion"`
	Kind       string             `json:"kind"`
	Spec       V1AuthnRequestSpec `json:"spec"`
}
