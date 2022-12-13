package secretpath

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/munditrade/provider-secret/apis/vault/v1alpha1"
	"github.com/munditrade/provider-secret/internal/common"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
)

const (
	engine                 = "test-engine"
	secretPathResourceName = "test-secret-path"
	ns                     = "test"
	path                   = "/dev"
)

func TestSecretPath_Observe(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type prepareMock func(m *common.MockSecretManager, reader *common.MockK8sReader)

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
		"when engine exist and path does not exist should queue to create a new path": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.SecretPath{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretPathResourceName,
						Namespace: ns,
					},
					Spec: v1alpha1.SecretPathSpec{
						ForProvider: v1alpha1.SecretPathParameters{
							Engine: engine,
							Path:   path,
						},
					},
				},
			},
			want: want{
				o:   managed.ExternalObservation{ResourceExists: false},
				err: nil,
			},
			prepareMock: func(m *common.MockSecretManager, reader *common.MockK8sReader) {
				reader.EXPECT().Get(gomock.Any(), types.NamespacedName{Namespace: ns, Name: engine}, gomock.Any()).
					DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj *v1alpha1.Engine) error {
						obj.ObjectMeta.Name = "test-engine"
						obj.Spec.ForProvider.Options = map[string]string{}
						return nil
					})
				m.EXPECT().GetSecrets(gomock.Any(), engine, path, map[string]string{}).
					Return(nil, errors.New(common.ErrNotFoundPath)).AnyTimes()
			},
		},
		"when engine exist and path exists should not queue to create a new path": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.SecretPath{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretPathResourceName,
						Namespace: ns,
					},
					Spec: v1alpha1.SecretPathSpec{
						ForProvider: v1alpha1.SecretPathParameters{
							Engine: engine,
							Path:   path,
						},
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
			prepareMock: func(m *common.MockSecretManager, reader *common.MockK8sReader) {
				reader.EXPECT().Get(gomock.Any(), types.NamespacedName{Namespace: ns, Name: engine}, gomock.Any()).
					DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj *v1alpha1.Engine) error {
						obj.ObjectMeta.Name = "test-engine"
						obj.Spec.ForProvider.Options = map[string]string{}
						return nil
					})
				m.EXPECT().GetSecrets(gomock.Any(), engine, path, map[string]string{}).
					Return(nil, nil).AnyTimes()
			},
		},
		"when get secrets fails should return an error": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.SecretPath{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretPathResourceName,
						Namespace: ns,
					},
					Spec: v1alpha1.SecretPathSpec{
						ForProvider: v1alpha1.SecretPathParameters{
							Engine: engine,
							Path:   path,
						},
					},
				},
			},
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New(errDataNotFoundInPath),
			},
			prepareMock: func(m *common.MockSecretManager, reader *common.MockK8sReader) {
				reader.EXPECT().Get(gomock.Any(), types.NamespacedName{Namespace: ns, Name: engine}, gomock.Any()).
					DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj *v1alpha1.Engine) error {
						obj.ObjectMeta.Name = engine
						obj.Spec.ForProvider.Options = map[string]string{}
						return nil
					})
				m.EXPECT().GetSecrets(gomock.Any(), engine, path, map[string]string{}).
					Return(nil, errors.New("ups")).AnyTimes()
			},
		},
		"when get get engine resource fails should fail the observe": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.SecretPath{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretPathResourceName,
						Namespace: ns,
					},
					Spec: v1alpha1.SecretPathSpec{
						ForProvider: v1alpha1.SecretPathParameters{
							Engine: engine,
							Path:   path,
						},
					},
				},
			},
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New(errEngineNotFound),
			},
			prepareMock: func(m *common.MockSecretManager, reader *common.MockK8sReader) {
				reader.EXPECT().Get(gomock.Any(), types.NamespacedName{Namespace: ns, Name: engine}, gomock.Any()).
					Return(errors.New("ups"))
			},
		},
	}
	for name, tc := range cases {
		testCase := tc
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := common.NewMockSecretManager(ctrl)
			reader := common.NewMockK8sReader(ctrl)

			testCase.prepareMock(mock, reader)

			e := external{service: mock, kubeReader: reader}
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

func TestSecretPath_Create(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type prepareMock func(m *common.MockSecretManager, reader *common.MockK8sReader)

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
		"should create a new secret path": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.SecretPath{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretPathResourceName,
						Namespace: ns,
					},
					Spec: v1alpha1.SecretPathSpec{
						ForProvider: v1alpha1.SecretPathParameters{
							Engine: engine,
							Path:   path,
						},
					},
				},
			},
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{},
				},
				err: nil,
			},
			prepareMock: func(m *common.MockSecretManager, reader *common.MockK8sReader) {
				reader.EXPECT().Get(gomock.Any(), types.NamespacedName{Namespace: ns, Name: engine}, gomock.Any()).
					DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj *v1alpha1.Engine) error {
						obj.ObjectMeta.Name = engine
						obj.Spec.ForProvider.Options = map[string]string{}
						return nil
					})
				m.EXPECT().Put(gomock.Any(), engine, path, emptyMap, map[string]string{}).
					Return(nil).AnyTimes()
			},
		},
		"when get engine fails should fail": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.SecretPath{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretPathResourceName,
						Namespace: ns,
					},
					Spec: v1alpha1.SecretPathSpec{
						ForProvider: v1alpha1.SecretPathParameters{
							Engine: engine,
							Path:   path,
						},
					},
				},
			},
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.New(errEngineNotFound),
			},
			prepareMock: func(m *common.MockSecretManager, reader *common.MockK8sReader) {
				reader.EXPECT().Get(gomock.Any(), types.NamespacedName{Namespace: ns, Name: engine}, gomock.Any()).
					Return(errors.New("ups"))
			},
		},
		"when create path fails should fail": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.SecretPath{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretPathResourceName,
						Namespace: ns,
					},
					Spec: v1alpha1.SecretPathSpec{
						ForProvider: v1alpha1.SecretPathParameters{
							Engine: engine,
							Path:   path,
						},
					},
				},
			},
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.New(errCreatingPath),
			},
			prepareMock: func(m *common.MockSecretManager, reader *common.MockK8sReader) {
				reader.EXPECT().Get(gomock.Any(), types.NamespacedName{Namespace: ns, Name: engine}, gomock.Any()).
					DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj *v1alpha1.Engine) error {
						obj.ObjectMeta.Name = engine
						obj.Spec.ForProvider.Options = map[string]string{}
						return nil
					})
				m.EXPECT().Put(gomock.Any(), engine, path, emptyMap, map[string]string{}).
					Return(errors.New("ups")).AnyTimes()
			},
		},
	}

	for name, tc := range cases {
		testCase := tc
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := common.NewMockSecretManager(ctrl)
			reader := common.NewMockK8sReader(ctrl)

			testCase.prepareMock(mock, reader)

			e := external{service: mock, kubeReader: reader}
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

func TestSecretPath_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type prepareMock func(m *common.MockSecretManager, reader *common.MockK8sReader)

	type want struct {
		err error
	}

	cases := map[string]struct {
		reason      string
		args        args
		want        want
		prepareMock prepareMock
	}{
		"should delete the path": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.SecretPath{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretPathResourceName,
						Namespace: ns,
					},
					Spec: v1alpha1.SecretPathSpec{
						ForProvider: v1alpha1.SecretPathParameters{
							Engine: engine,
							Path:   path,
						},
					},
				},
			},
			want: want{
				err: nil,
			},
			prepareMock: func(m *common.MockSecretManager, reader *common.MockK8sReader) {
				reader.EXPECT().Get(gomock.Any(), types.NamespacedName{Namespace: ns, Name: engine}, gomock.Any()).
					DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj *v1alpha1.Engine) error {
						obj.ObjectMeta.Name = engine
						obj.Spec.ForProvider.Options = map[string]string{}
						return nil
					})
				m.EXPECT().DeletePath(gomock.Any(), engine, path, map[string]string{}).
					Return(nil).AnyTimes()
			},
		},
		"should fail when delete path fails": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.SecretPath{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretPathResourceName,
						Namespace: ns,
					},
					Spec: v1alpha1.SecretPathSpec{
						ForProvider: v1alpha1.SecretPathParameters{
							Engine: engine,
							Path:   path,
						},
					},
				},
			},
			want: want{
				err: errors.New("boom"),
			},
			prepareMock: func(m *common.MockSecretManager, reader *common.MockK8sReader) {
				reader.EXPECT().Get(gomock.Any(), types.NamespacedName{Namespace: ns, Name: engine}, gomock.Any()).
					DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj *v1alpha1.Engine) error {
						obj.ObjectMeta.Name = engine
						obj.Spec.ForProvider.Options = map[string]string{}
						return nil
					})
				m.EXPECT().DeletePath(gomock.Any(), engine, path, map[string]string{}).
					Return(errors.New("boom")).AnyTimes()
			},
		},
		"should success when engine does not exist": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.SecretPath{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretPathResourceName,
						Namespace: ns,
					},
					Spec: v1alpha1.SecretPathSpec{
						ForProvider: v1alpha1.SecretPathParameters{
							Engine: engine,
							Path:   path,
						},
					},
				},
			},
			want: want{
				err: nil,
			},
			prepareMock: func(m *common.MockSecretManager, reader *common.MockK8sReader) {
				reader.EXPECT().Get(gomock.Any(), types.NamespacedName{Namespace: ns, Name: engine}, gomock.Any()).
					Return(errors.New(common.ErrNoParentReferences))
			},
		},
		"should fail when get engine fails": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.SecretPath{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretPathResourceName,
						Namespace: ns,
					},
					Spec: v1alpha1.SecretPathSpec{
						ForProvider: v1alpha1.SecretPathParameters{
							Engine: engine,
							Path:   path,
						},
					},
				},
			},
			want: want{
				err: errors.New("boom"),
			},
			prepareMock: func(m *common.MockSecretManager, reader *common.MockK8sReader) {
				reader.EXPECT().Get(gomock.Any(), types.NamespacedName{Namespace: ns, Name: engine}, gomock.Any()).
					DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj *v1alpha1.Engine) error {
						obj.ObjectMeta.Name = engine
						obj.Spec.ForProvider.Options = map[string]string{}
						return nil
					})
				m.EXPECT().DeletePath(gomock.Any(), engine, path, map[string]string{}).
					Return(errors.New("boom")).AnyTimes()
			},
		},
		"should not fail when delete secret fails when path does not exist": {
			args: args{
				ctx: context.Background(),
				mg: &v1alpha1.SecretPath{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretPathResourceName,
						Namespace: ns,
					},
					Spec: v1alpha1.SecretPathSpec{
						ForProvider: v1alpha1.SecretPathParameters{
							Engine: engine,
							Path:   path,
						},
					},
				},
			},
			want: want{
				err: nil,
			},
			prepareMock: func(m *common.MockSecretManager, reader *common.MockK8sReader) {
				reader.EXPECT().Get(gomock.Any(), types.NamespacedName{Namespace: ns, Name: engine}, gomock.Any()).
					DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj *v1alpha1.Engine) error {
						obj.ObjectMeta.Name = engine
						obj.Spec.ForProvider.Options = map[string]string{}
						return nil
					})
				m.EXPECT().DeletePath(gomock.Any(), engine, path, map[string]string{}).
					Return(errors.New(common.ErrNotFoundPath)).AnyTimes()
			},
		},
	}

	for name, tc := range cases {
		testCase := tc
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := common.NewMockSecretManager(ctrl)
			reader := common.NewMockK8sReader(ctrl)

			testCase.prepareMock(mock, reader)

			e := external{service: mock, kubeReader: reader}
			err := e.Delete(testCase.args.ctx, testCase.args.mg)

			if diff := cmp.Diff(testCase.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", testCase.reason, diff)
			}
		})
	}
}
