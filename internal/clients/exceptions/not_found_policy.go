package exceptions

import "fmt"

type NotFoundPolicy struct {
	policyName string
}

func NewNotFoundPolicy(policyName string) *NotFoundPolicy {
	return &NotFoundPolicy{policyName: policyName}
}

func (m *NotFoundPolicy) Error() string {
	return fmt.Sprintf("policy not found %s", m.policyName)
}
