package exceptions

import "fmt"

type NotFoundPolicyPath struct {
	policyName string
}

func NewNotFoundPolicyPath(policyName string) *NotFoundPolicyPath {
	return &NotFoundPolicyPath{policyName: policyName}
}

func (m *NotFoundPolicyPath) Error() string {
	return fmt.Sprintf("not found paths for policy %s", m.policyName)
}
