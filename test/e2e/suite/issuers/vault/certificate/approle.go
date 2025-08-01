/*
Copyright 2020 The cert-manager Authors.

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

package certificate

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cert-manager/cert-manager/e2e-tests/framework"
	"github.com/cert-manager/cert-manager/e2e-tests/framework/addon"
	vaultaddon "github.com/cert-manager/cert-manager/e2e-tests/framework/addon/vault"
	"github.com/cert-manager/cert-manager/e2e-tests/framework/helper/featureset"
	"github.com/cert-manager/cert-manager/e2e-tests/framework/helper/validation"
	"github.com/cert-manager/cert-manager/e2e-tests/framework/helper/validation/certificates"
	"github.com/cert-manager/cert-manager/e2e-tests/util"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/cert-manager/cert-manager/test/unit/gen"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = framework.CertManagerDescribe("Vault Issuer Certificate (AppRole, CA without root)", func() {
	fs := featureset.NewFeatureSet(featureset.SaveRootCAToSecret)
	runVaultAppRoleTests(cmapi.IssuerKind, false, fs)
})
var _ = framework.CertManagerDescribe("Vault Issuer Certificate (AppRole, CA with root)", func() {
	fs := featureset.NewFeatureSet()
	runVaultAppRoleTests(cmapi.IssuerKind, true, fs)
})

var _ = framework.CertManagerDescribe("Vault ClusterIssuer Certificate (AppRole, CA without root)", func() {
	fs := featureset.NewFeatureSet(featureset.SaveRootCAToSecret)
	runVaultAppRoleTests(cmapi.ClusterIssuerKind, false, fs)
})
var _ = framework.CertManagerDescribe("Vault ClusterIssuer Certificate (AppRole, CA with root)", func() {
	fs := featureset.NewFeatureSet()
	runVaultAppRoleTests(cmapi.ClusterIssuerKind, true, fs)
})

func runVaultAppRoleTests(issuerKind string, testWithRoot bool, unsupportedFeatures featureset.FeatureSet) {
	f := framework.NewDefaultFramework("create-vault-certificate")

	certificateName := "test-vault-certificate"
	certificateSecretName := "test-vault-certificate"
	var vaultIssuerName string

	appRoleSecretGeneratorName := "vault-approle-secret-"
	var roleId, secretId string
	var vaultSecretName, vaultSecretNamespace string

	var setup *vaultaddon.VaultInitializer

	BeforeEach(func(testingCtx context.Context) {
		By("Configuring the Vault server")
		if issuerKind == cmapi.IssuerKind {
			vaultSecretNamespace = f.Namespace.Name
		} else {
			vaultSecretNamespace = f.Config.Addons.CertManager.ClusterResourceNamespace
		}

		setup = vaultaddon.NewVaultInitializerAppRole(
			addon.Base.Details().KubeClient,
			*addon.Vault.Details(),
			testWithRoot,
		)
		Expect(setup.Init(testingCtx)).NotTo(HaveOccurred(), "failed to init vault")
		Expect(setup.Setup(testingCtx)).NotTo(HaveOccurred(), "failed to setup vault")

		var err error
		roleId, secretId, err = setup.CreateAppRole(testingCtx)
		Expect(err).NotTo(HaveOccurred())

		sec, err := f.KubeClientSet.CoreV1().Secrets(vaultSecretNamespace).Create(testingCtx, vaultaddon.NewVaultAppRoleSecret(appRoleSecretGeneratorName, secretId), metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
		vaultSecretName = sec.Name
	})

	JustAfterEach(func(testingCtx context.Context) {
		By("Cleaning up")
		Expect(setup.Clean(testingCtx)).NotTo(HaveOccurred())

		if issuerKind == cmapi.IssuerKind {
			err := f.CertManagerClientSet.CertmanagerV1().Issuers(f.Namespace.Name).Delete(testingCtx, vaultIssuerName, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		} else {
			err := f.CertManagerClientSet.CertmanagerV1().ClusterIssuers().Delete(testingCtx, vaultIssuerName, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		}

		err := f.KubeClientSet.CoreV1().Secrets(vaultSecretNamespace).Delete(testingCtx, vaultSecretName, metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should generate a new valid certificate", func(testingCtx context.Context) {
		By("Creating an Issuer")
		vaultURL := addon.Vault.Details().URL

		certClient := f.CertManagerClientSet.CertmanagerV1().Certificates(f.Namespace.Name)

		var err error
		if issuerKind == cmapi.IssuerKind {
			vaultIssuer := gen.IssuerWithRandomName("test-vault-issuer-",
				gen.SetIssuerNamespace(f.Namespace.Name),
				gen.SetIssuerVaultURL(vaultURL),
				gen.SetIssuerVaultPath(setup.IntermediateSignPath()),
				gen.SetIssuerVaultCABundle(addon.Vault.Details().VaultCA),
				gen.SetIssuerVaultAppRoleAuth("secretkey", vaultSecretName, roleId, setup.AppRoleAuthPath()))
			iss, err := f.CertManagerClientSet.CertmanagerV1().Issuers(f.Namespace.Name).Create(testingCtx, vaultIssuer, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			vaultIssuerName = iss.Name
		} else {
			vaultIssuer := gen.ClusterIssuerWithRandomName("test-vault-issuer-",
				gen.SetIssuerVaultURL(vaultURL),
				gen.SetIssuerVaultPath(setup.IntermediateSignPath()),
				gen.SetIssuerVaultCABundle(addon.Vault.Details().VaultCA),
				gen.SetIssuerVaultAppRoleAuth("secretkey", vaultSecretName, roleId, setup.AppRoleAuthPath()))
			iss, err := f.CertManagerClientSet.CertmanagerV1().ClusterIssuers().Create(testingCtx, vaultIssuer, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			vaultIssuerName = iss.Name
		}

		By("Waiting for Issuer to become Ready")

		if issuerKind == cmapi.IssuerKind {
			err = util.WaitForIssuerCondition(testingCtx, f.CertManagerClientSet.CertmanagerV1().Issuers(f.Namespace.Name),
				vaultIssuerName,
				cmapi.IssuerCondition{
					Type:   cmapi.IssuerConditionReady,
					Status: cmmeta.ConditionTrue,
				})
		} else {
			err = util.WaitForClusterIssuerCondition(testingCtx, f.CertManagerClientSet.CertmanagerV1().ClusterIssuers(),
				vaultIssuerName,
				cmapi.IssuerCondition{
					Type:   cmapi.IssuerConditionReady,
					Status: cmmeta.ConditionTrue,
				})
		}

		Expect(err).NotTo(HaveOccurred())

		By("Creating a Certificate")
		cert, err := certClient.Create(testingCtx, util.NewCertManagerVaultCertificate(certificateName, certificateSecretName, vaultIssuerName, issuerKind, nil, nil), metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for the Certificate to be issued...")
		cert, err = f.Helper().WaitForCertificateReadyAndDoneIssuing(testingCtx, cert, time.Minute*5)
		Expect(err).NotTo(HaveOccurred())

		By("Validating the issued Certificate...")
		err = f.Helper().ValidateCertificate(cert, validation.CertificateSetForUnsupportedFeatureSet(unsupportedFeatures)...)
		Expect(err).NotTo(HaveOccurred())

	})

	cases := []struct {
		inputDuration    *metav1.Duration
		inputRenewBefore *metav1.Duration
		expectedDuration time.Duration
		label            string
		event            string
	}{
		{
			inputDuration:    &metav1.Duration{Duration: time.Hour * 24 * 35},
			inputRenewBefore: nil,
			expectedDuration: time.Hour * 24 * 35,
			label:            "valid for 35 days",
		},
		{
			inputDuration:    nil,
			inputRenewBefore: nil,
			expectedDuration: time.Hour * 24 * 90,
			label:            "valid for the default value (90 days)",
		},
		{
			inputDuration:    &metav1.Duration{Duration: time.Hour * 24 * 365},
			inputRenewBefore: nil,
			expectedDuration: time.Hour * 24 * 90,
			label:            "with Vault configured maximum TTL duration (90 days) when requested duration is greater than TTL",
		},
		{
			inputDuration:    &metav1.Duration{Duration: time.Hour * 24 * 240},
			inputRenewBefore: &metav1.Duration{Duration: time.Hour * 24 * 120},
			expectedDuration: time.Hour * 24 * 90,
			label:            "with a warning event when renewBefore is bigger than the duration",
		},
	}

	for _, v := range cases {
		It("should generate a new certificate "+v.label, func(testingCtx context.Context) {
			By("Creating an Issuer")

			var err error
			if issuerKind == cmapi.IssuerKind {
				vaultIssuer := gen.IssuerWithRandomName("test-vault-issuer-",
					gen.SetIssuerNamespace(f.Namespace.Name),
					gen.SetIssuerVaultURL(addon.Vault.Details().URL),
					gen.SetIssuerVaultPath(setup.IntermediateSignPath()),
					gen.SetIssuerVaultCABundle(addon.Vault.Details().VaultCA),
					gen.SetIssuerVaultAppRoleAuth("secretkey", vaultSecretName, roleId, setup.AppRoleAuthPath()))
				iss, err := f.CertManagerClientSet.CertmanagerV1().Issuers(f.Namespace.Name).Create(testingCtx, vaultIssuer, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())

				vaultIssuerName = iss.Name
			} else {
				vaultIssuer := gen.ClusterIssuerWithRandomName("test-vault-issuer-",
					gen.SetIssuerVaultURL(addon.Vault.Details().URL),
					gen.SetIssuerVaultPath(setup.IntermediateSignPath()),
					gen.SetIssuerVaultCABundle(addon.Vault.Details().VaultCA),
					gen.SetIssuerVaultAppRoleAuth("secretkey", vaultSecretName, roleId, setup.AppRoleAuthPath()))
				iss, err := f.CertManagerClientSet.CertmanagerV1().ClusterIssuers().Create(testingCtx, vaultIssuer, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())

				vaultIssuerName = iss.Name
			}

			By("Waiting for Issuer to become Ready")

			if issuerKind == cmapi.IssuerKind {
				err = util.WaitForIssuerCondition(testingCtx, f.CertManagerClientSet.CertmanagerV1().Issuers(f.Namespace.Name),
					vaultIssuerName,
					cmapi.IssuerCondition{
						Type:   cmapi.IssuerConditionReady,
						Status: cmmeta.ConditionTrue,
					})
			} else {
				err = util.WaitForClusterIssuerCondition(testingCtx, f.CertManagerClientSet.CertmanagerV1().ClusterIssuers(),
					vaultIssuerName,
					cmapi.IssuerCondition{
						Type:   cmapi.IssuerConditionReady,
						Status: cmmeta.ConditionTrue,
					})
			}
			Expect(err).NotTo(HaveOccurred())

			By("Creating a Certificate")
			cert, err := f.CertManagerClientSet.CertmanagerV1().Certificates(f.Namespace.Name).Create(testingCtx, util.NewCertManagerVaultCertificate(certificateName, certificateSecretName, vaultIssuerName, issuerKind, v.inputDuration, v.inputRenewBefore), metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			By("Waiting for the Certificate to be issued...")
			cert, err = f.Helper().WaitForCertificateReadyAndDoneIssuing(testingCtx, cert, time.Minute*5)
			Expect(err).NotTo(HaveOccurred())

			By("Validating the issued Certificate...")
			err = f.Helper().ValidateCertificate(cert, validation.CertificateSetForUnsupportedFeatureSet(
				// X509 certificate duration will not match requested duration in Certificate for
				// all test cases. Instead, we disable the duration check here and compare the duration
				// with our expected duration below (by calling ExpectDuration explicitly).
				unsupportedFeatures.Clone().
					Insert(featureset.DurationFeature),
			)...)
			Expect(err).NotTo(HaveOccurred())

			// Vault subtract 30 seconds to the NotBefore date.
			err = f.Helper().ValidateCertificate(cert, certificates.ExpectDuration(v.expectedDuration, time.Second*30))
			Expect(err).NotTo(HaveOccurred())
		})
	}
}
