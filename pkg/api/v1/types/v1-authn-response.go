package types

type V1AuthnResponseUser struct {
	UserName string            `json:"username,omitempty"`
	UID      string            `json:"uid,omitempty"`
	Groups   []string          `json:"groups,omitempty"`
	Extra    map[string]string `json:"extra,omitempty"`
}

type V1AuthnResponseStatus struct {
	Authenticated bool                 `json:"authenticated"`
	User          *V1AuthnResponseUser `json:"user,omitempty"`
}

type V1AuthnResponse struct {
	APIVersion string                `json:"apiVersion"`
	Kind       string                `json:"kind"`
	Status     V1AuthnResponseStatus `json:"status"`
}
