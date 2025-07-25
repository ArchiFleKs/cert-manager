/*
Copyright The cert-manager Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	context "context"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	applyconfigurationscertmanagerv1 "github.com/cert-manager/cert-manager/pkg/client/applyconfigurations/certmanager/v1"
	scheme "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// CertificatesGetter has a method to return a CertificateInterface.
// A group's client should implement this interface.
type CertificatesGetter interface {
	Certificates(namespace string) CertificateInterface
}

// CertificateInterface has methods to work with Certificate resources.
type CertificateInterface interface {
	Create(ctx context.Context, certificate *certmanagerv1.Certificate, opts metav1.CreateOptions) (*certmanagerv1.Certificate, error)
	Update(ctx context.Context, certificate *certmanagerv1.Certificate, opts metav1.UpdateOptions) (*certmanagerv1.Certificate, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, certificate *certmanagerv1.Certificate, opts metav1.UpdateOptions) (*certmanagerv1.Certificate, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*certmanagerv1.Certificate, error)
	List(ctx context.Context, opts metav1.ListOptions) (*certmanagerv1.CertificateList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *certmanagerv1.Certificate, err error)
	Apply(ctx context.Context, certificate *applyconfigurationscertmanagerv1.CertificateApplyConfiguration, opts metav1.ApplyOptions) (result *certmanagerv1.Certificate, err error)
	// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
	ApplyStatus(ctx context.Context, certificate *applyconfigurationscertmanagerv1.CertificateApplyConfiguration, opts metav1.ApplyOptions) (result *certmanagerv1.Certificate, err error)
	CertificateExpansion
}

// certificates implements CertificateInterface
type certificates struct {
	*gentype.ClientWithListAndApply[*certmanagerv1.Certificate, *certmanagerv1.CertificateList, *applyconfigurationscertmanagerv1.CertificateApplyConfiguration]
}

// newCertificates returns a Certificates
func newCertificates(c *CertmanagerV1Client, namespace string) *certificates {
	return &certificates{
		gentype.NewClientWithListAndApply[*certmanagerv1.Certificate, *certmanagerv1.CertificateList, *applyconfigurationscertmanagerv1.CertificateApplyConfiguration](
			"certificates",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *certmanagerv1.Certificate { return &certmanagerv1.Certificate{} },
			func() *certmanagerv1.CertificateList { return &certmanagerv1.CertificateList{} },
		),
	}
}
