package engine

import (
	"context"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/munditrade/provider-secret/apis/vault/v1alpha1"
	"github.com/munditrade/provider-secret/internal/common"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

func TestEngine_Observe(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type prepareMock func(m *common.MockSecretManager)

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
		"when engine does not exist must return false in ResourceExist": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Engine{
					ObjectMeta: v1.ObjectMeta{
						Name: "test-engine",
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:    false,
					ResourceUpToDate:  true,
					ConnectionDetails: managed.ConnectionDetails{},
				},
				err: nil,
			},
			prepareMock: func(m *common.MockSecretManager) {
				m.EXPECT().ExistEngine(gomock.Any(), "test-engine").Return(false, nil).AnyTimes()
			},
		},
		"when engine does not exist must return true in ResourceExist": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Engine{
					ObjectMeta: v1.ObjectMeta{
						Name: "test-engine",
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: managed.ConnectionDetails{},
				},
				err: nil,
			},
			prepareMock: func(m *common.MockSecretManager) {
				m.EXPECT().ExistEngine(gomock.Any(), "test-engine").Return(true, nil).AnyTimes()
			},
		},
		"when ExistEngine function returns an error should return errErrorGettingEngine err": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Engine{
					ObjectMeta: v1.ObjectMeta{
						Name: "test-engine",
					},
				},
			},
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New(errErrorGettingEngine),
			},
			prepareMock: func(m *common.MockSecretManager) {
				m.EXPECT().ExistEngine(gomock.Any(), "test-engine").Return(false, errors.New("boom")).
					AnyTimes()
			},
		},
	}

	for name, tc := range cases {
		testCase := tc
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := common.NewMockSecretManager(ctrl)
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

func TestEngine_Create(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type prepareMock func(m *common.MockSecretManager)

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
		"should create a new engine": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Engine{
					Spec: v1alpha1.EngineSpec{
						ForProvider: v1alpha1.EngineParameters{
							Storage: "kv",
							Options: map[string]string{
								"version": "1",
							},
						},
					},
					ObjectMeta: v1.ObjectMeta{
						Name: "test-engine",
					},
				},
			},
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{},
				},
				err: nil,
			},
			prepareMock: func(m *common.MockSecretManager) {
				m.EXPECT().CreateEngine(gomock.Any(), "test-engine", "kv", map[string]string{
					"version": "1",
				}).Return(nil).AnyTimes()
			},
		},
		"when engine creation fails should fail all": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Engine{
					Spec: v1alpha1.EngineSpec{
						ForProvider: v1alpha1.EngineParameters{
							Storage: "kv",
							Options: map[string]string{
								"version": "1",
							},
						},
					},
					ObjectMeta: v1.ObjectMeta{
						Name: "test-engine",
					},
				},
			},
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.New(errCreatingEngine),
			},
			prepareMock: func(m *common.MockSecretManager) {
				m.EXPECT().CreateEngine(gomock.Any(), "test-engine", "kv", map[string]string{
					"version": "1",
				}).Return(errors.New("boom")).AnyTimes()
			},
		},
	}

	for name, tc := range cases {
		testCase := tc
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := common.NewMockSecretManager(ctrl)
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

func TestEngine_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type prepareMock func(m *common.MockSecretManager)

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
		"when engine exists should be deleted": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Engine{
					ObjectMeta: v1.ObjectMeta{
						Name: "test-engine",
					},
				},
			},
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{},
				},
				err: nil,
			},
			prepareMock: func(m *common.MockSecretManager) {
				m.EXPECT().ExistEngine(gomock.Any(), "test-engine").Return(true, nil).AnyTimes()
				m.EXPECT().DeleteEngine(gomock.Any(), "test-engine").Return(nil).AnyTimes()
			},
		},
		"when engine does not exist should not return an error": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.Engine{
					ObjectMeta: v1.ObjectMeta{
						Name: "test-engine",
					},
				},
			},
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{},
				},
				err: nil,
			},
			prepareMock: func(m *common.MockSecretManager) {
				m.EXPECT().ExistEngine(gomock.Any(), "test-engine").Return(false, nil).AnyTimes()
			},
		},
	}

	for name, tc := range cases {
		testCase := tc
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := common.NewMockSecretManager(ctrl)
			testCase.prepareMock(mock)
			e := external{service: mock}
			err := e.Delete(testCase.args.ctx, testCase.args.mg)

			if diff := cmp.Diff(testCase.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", testCase.reason, diff)
			}
		})
	}
}
