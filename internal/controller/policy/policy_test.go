package policy

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/munditrade/provider-secret/apis/vault/v1alpha1"
	"github.com/munditrade/provider-secret/internal/clients"
	"github.com/munditrade/provider-secret/internal/clients/exceptions"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPolicy_Observe(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type prepareMock func(m *clients.MockPolicyManager)

	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason      string
		args        args
		want        want
		prepareMock prepareMock
	}{
		"when rules not exist this the policies should be created": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Policy{
					ObjectMeta: v1.ObjectMeta{
						Name: "the-super-policy",
					},
					Spec: v1alpha1.PolicySpec{
						ForProvider: v1alpha1.PolicyParameters{
							Rules: []v1alpha1.Rule{
								{
									Path:         "/dev",
									Capabilities: []string{"list, read"},
								},
								{
									Path:         "/stg",
									Capabilities: []string{"read"},
								},
							},
						},
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
				err: nil,
			},
			prepareMock: func(m *clients.MockPolicyManager) {
				m.EXPECT().Get(gomock.Any(), "the-super-policy").
					Return(nil, exceptions.NewNotFoundPolicy("the-super-policy")).AnyTimes()
			},
		},
		"when add new rules should be updated the policies": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Policy{
					ObjectMeta: v1.ObjectMeta{
						Name: "the-super-policy",
					},
					Spec: v1alpha1.PolicySpec{
						ForProvider: v1alpha1.PolicyParameters{
							Rules: []v1alpha1.Rule{
								{
									Path:         "/dev",
									Capabilities: []string{"list, read"},
								},
							},
						},
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: managed.ConnectionDetails{},
				},
				err: nil,
			},
			prepareMock: func(m *clients.MockPolicyManager) {
				m.EXPECT().Get(gomock.Any(), "the-super-policy").
					Return([]clients.Policy{
						{
							PathConfig: clients.PathConfig{
								Path:         "/dev",
								Capabilities: []string{"list"},
							},
						},
						{
							PathConfig: clients.PathConfig{
								Path:         "/stg",
								Capabilities: []string{"read"},
							},
						},
					}, nil).AnyTimes()
			},
		},
		"when rules change should be updated the policies": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Policy{
					ObjectMeta: v1.ObjectMeta{
						Name: "the-super-policy",
					},
					Spec: v1alpha1.PolicySpec{
						ForProvider: v1alpha1.PolicyParameters{
							Rules: []v1alpha1.Rule{
								{
									Path:         "/dev",
									Capabilities: []string{"list, read"},
								},
								{
									Path:         "/stg",
									Capabilities: []string{"read"},
								},
							},
						},
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: managed.ConnectionDetails{},
				},
				err: nil,
			},
			prepareMock: func(m *clients.MockPolicyManager) {
				m.EXPECT().Get(gomock.Any(), "the-super-policy").
					Return([]clients.Policy{
						{
							PathConfig: clients.PathConfig{
								Path:         "/dev",
								Capabilities: []string{"list"},
							},
						},
						{
							PathConfig: clients.PathConfig{
								Path:         "/stg",
								Capabilities: []string{"read"},
							},
						},
					}, nil).AnyTimes()
			},
		},
	}

	for name, tc := range cases {
		testCase := tc
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := clients.NewMockPolicyManager(ctrl)
			testCase.prepareMock(mock)
			e := external{service: mock}
			got, err := e.Observe(testCase.args.ctx, testCase.args.mg)

			if diff := cmp.Diff(testCase.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", testCase.reason, diff)
			}
			if diff := cmp.Diff(testCase.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", testCase.reason, diff)
			}
		})
	}
}

func TestPolicy_Create(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type prepareMock func(m *clients.MockPolicyManager)

	type want struct {
		o   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		reason      string
		args        args
		want        want
		prepareMock prepareMock
	}{
		"should create new policies in path": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Policy{
					ObjectMeta: v1.ObjectMeta{
						Name: "the-super-policy",
					},
					Spec: v1alpha1.PolicySpec{
						ForProvider: v1alpha1.PolicyParameters{
							Rules: []v1alpha1.Rule{
								{
									Path:         "/dev",
									Capabilities: []string{"list, read"},
								},
								{
									Path:         "/stg",
									Capabilities: []string{"read"},
								},
							},
						},
					},
				},
			},
			want: want{
				o:   managed.ExternalCreation{ConnectionDetails: managed.ConnectionDetails{}},
				err: nil,
			},
			prepareMock: func(m *clients.MockPolicyManager) {
				m.EXPECT().Put(gomock.Any(), "the-super-policy", clients.Policy{
					PathConfig: clients.PathConfig{
						Path:         "/dev",
						Capabilities: []string{"list, read"},
					},
				}).Return(nil).Times(1)
				m.EXPECT().Put(gomock.Any(), "the-super-policy", clients.Policy{
					PathConfig: clients.PathConfig{
						Path:         "/stg",
						Capabilities: []string{"read"},
					},
				}).Return(nil).Times(1)
			},
		},
		"should return an error when creation policy fails": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Policy{
					ObjectMeta: v1.ObjectMeta{
						Name: "the-super-policy",
					},
					Spec: v1alpha1.PolicySpec{
						ForProvider: v1alpha1.PolicyParameters{
							Rules: []v1alpha1.Rule{
								{
									Path:         "/dev",
									Capabilities: []string{"list, read"},
								},
								{
									Path:         "/stg",
									Capabilities: []string{"read"},
								},
							},
						},
					},
				},
			},
			want: want{
				o:   managed.ExternalCreation{ConnectionDetails: managed.ConnectionDetails{}},
				err: errors.New("ups"),
			},
			prepareMock: func(m *clients.MockPolicyManager) {
				m.EXPECT().Put(gomock.Any(), "the-super-policy", clients.Policy{
					PathConfig: clients.PathConfig{
						Path:         "/dev",
						Capabilities: []string{"list, read"},
					},
				}).Return(nil).Times(1)
				m.EXPECT().Put(gomock.Any(), "the-super-policy", clients.Policy{
					PathConfig: clients.PathConfig{
						Path:         "/stg",
						Capabilities: []string{"read"},
					},
				}).Return(errors.New("ups")).Times(1)
			},
		},
	}

	for name, tc := range cases {
		testCase := tc
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := clients.NewMockPolicyManager(ctrl)
			testCase.prepareMock(mock)
			e := external{service: mock}
			got, err := e.Create(testCase.args.ctx, testCase.args.mg)

			if diff := cmp.Diff(testCase.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", testCase.reason, diff)
			}
			if diff := cmp.Diff(testCase.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", testCase.reason, diff)
			}
		})
	}
}

func TestPolicy_Update(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type prepareMock func(m *clients.MockPolicyManager)

	type want struct {
		o   managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		reason      string
		args        args
		want        want
		prepareMock prepareMock
	}{
		"should update new policies in path": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Policy{
					ObjectMeta: v1.ObjectMeta{
						Name: "the-super-policy",
					},
					Spec: v1alpha1.PolicySpec{
						ForProvider: v1alpha1.PolicyParameters{
							Rules: []v1alpha1.Rule{
								{
									Path:         "/dev",
									Capabilities: []string{"list, read"},
								},
								{
									Path:         "/stg",
									Capabilities: []string{"read"},
								},
							},
						},
					},
				},
			},
			want: want{
				o:   managed.ExternalUpdate{ConnectionDetails: managed.ConnectionDetails{}},
				err: nil,
			},
			prepareMock: func(m *clients.MockPolicyManager) {
				m.EXPECT().Put(gomock.Any(), "the-super-policy", clients.Policy{
					PathConfig: clients.PathConfig{
						Path:         "/dev",
						Capabilities: []string{"list, read"},
					},
				}).Return(nil).Times(1)
				m.EXPECT().Put(gomock.Any(), "the-super-policy", clients.Policy{
					PathConfig: clients.PathConfig{
						Path:         "/stg",
						Capabilities: []string{"read"},
					},
				}).Return(nil).Times(1)
			},
		},
		"should return an error when update policy fails": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Policy{
					ObjectMeta: v1.ObjectMeta{
						Name: "the-super-policy",
					},
					Spec: v1alpha1.PolicySpec{
						ForProvider: v1alpha1.PolicyParameters{
							Rules: []v1alpha1.Rule{
								{
									Path:         "/dev",
									Capabilities: []string{"list, read"},
								},
								{
									Path:         "/stg",
									Capabilities: []string{"read"},
								},
							},
						},
					},
				},
			},
			want: want{
				o:   managed.ExternalUpdate{ConnectionDetails: managed.ConnectionDetails{}},
				err: errors.New("ups"),
			},
			prepareMock: func(m *clients.MockPolicyManager) {
				m.EXPECT().Put(gomock.Any(), "the-super-policy", clients.Policy{
					PathConfig: clients.PathConfig{
						Path:         "/dev",
						Capabilities: []string{"list, read"},
					},
				}).Return(nil).Times(1)
				m.EXPECT().Put(gomock.Any(), "the-super-policy", clients.Policy{
					PathConfig: clients.PathConfig{
						Path:         "/stg",
						Capabilities: []string{"read"},
					},
				}).Return(errors.New("ups")).Times(1)
			},
		},
	}

	for name, tc := range cases {
		testCase := tc
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := clients.NewMockPolicyManager(ctrl)
			testCase.prepareMock(mock)
			e := external{service: mock}
			got, err := e.Update(testCase.args.ctx, testCase.args.mg)

			if diff := cmp.Diff(testCase.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", testCase.reason, diff)
			}
			if diff := cmp.Diff(testCase.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", testCase.reason, diff)
			}
		})
	}
}

func TestPolicy_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type prepareMock func(m *clients.MockPolicyManager)

	type want struct {
		err error
	}

	cases := map[string]struct {
		reason      string
		args        args
		want        want
		prepareMock prepareMock
	}{
		"should delete policy": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Policy{
					ObjectMeta: v1.ObjectMeta{
						Name: "the-super-policy",
					},
					Spec: v1alpha1.PolicySpec{
						ForProvider: v1alpha1.PolicyParameters{
							Rules: []v1alpha1.Rule{
								{
									Path:         "/dev",
									Capabilities: []string{"list, read"},
								},
								{
									Path:         "/stg",
									Capabilities: []string{"read"},
								},
							},
						},
					},
				},
			},
			want: want{
				err: nil,
			},
			prepareMock: func(m *clients.MockPolicyManager) {
				m.EXPECT().Delete(gomock.Any(), "the-super-policy").Return(nil).Times(1)
			},
		},
		"should return an error when delete policy fails": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Policy{
					ObjectMeta: v1.ObjectMeta{
						Name: "the-super-policy",
					},
					Spec: v1alpha1.PolicySpec{
						ForProvider: v1alpha1.PolicyParameters{
							Rules: []v1alpha1.Rule{
								{
									Path:         "/dev",
									Capabilities: []string{"list, read"},
								},
								{
									Path:         "/stg",
									Capabilities: []string{"read"},
								},
							},
						},
					},
				},
			},
			want: want{
				err: errors.New("ups"),
			},
			prepareMock: func(m *clients.MockPolicyManager) {
				m.EXPECT().Delete(gomock.Any(), "the-super-policy").Return(errors.New("ups")).Times(1)
			},
		},
	}

	for name, tc := range cases {
		testCase := tc
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := clients.NewMockPolicyManager(ctrl)
			testCase.prepareMock(mock)
			e := external{service: mock}
			err := e.Delete(testCase.args.ctx, testCase.args.mg)

			if diff := cmp.Diff(testCase.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", testCase.reason, diff)
			}
		})
	}
}
