package kudo

import (
	"fmt"
	"os"
	"reflect"

	"github.com/Masterminds/semver"
	"github.com/kudobuilder/kudo/pkg/apis/kudo/v1beta1"
	"github.com/kudobuilder/kudo/pkg/kudoctl/packages"
	"github.com/kudobuilder/kudo/pkg/kudoctl/packages/install"
	"github.com/kudobuilder/kudo/pkg/kudoctl/packages/resolver"
	"github.com/kudobuilder/kudo/pkg/kudoctl/util/kudo"
	"github.com/kudobuilder/kudo/pkg/kudoctl/util/repo"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/zmalik/kudo-bridge/bridge-controller/pkg/apis/kudobridge/v1alpha1"
	"github.com/zmalik/kudo-bridge/crd-controller/pkg/client"
)

type KUDOClient struct {
	c *client.Client

	kc              *kudo.Client
	kudoPackageName string
	version         string
	appVersion      string
	repoURL         string

	resources *packages.Resources
}

func NewKUDOClient(k *client.Client, bi v1alpha1.BridgeInstance) (*KUDOClient, error) {
	kc, err := kudo.NewClient(os.Getenv("KUBECONFIG"), 0, true)
	if err != nil {
		return nil, err
	}
	return &KUDOClient{
		c:               k,
		kc:              kc,
		kudoPackageName: bi.Spec.KUDOOperator.Package,
		version:         bi.Spec.KUDOOperator.Version,
		appVersion:      bi.Spec.KUDOOperator.AppVersion,
		repoURL:         bi.Spec.KUDOOperator.KUDORepository,
	}, nil
}

func (k *KUDOClient) GetOVOrInstall(crd *unstructured.Unstructured) (*v1beta1.OperatorVersion, error) {
	if instance, err := k.kc.GetInstance(crd.GetName(), crd.GetNamespace()); err == nil && instance != nil {
		// already installed
		log.Infof("found the KUDO Instance %s/%s", crd.GetNamespace(), crd.GetName())
		log.Infof("fetching the KUDO Instance Operator %s/%s", instance.Spec.OperatorVersion.Name, instance.GetNamespace())
		return k.kc.GetOperatorVersion(instance.Spec.OperatorVersion.Name, instance.GetNamespace())
	}

	// install OV
	return k.InstallOV(crd)
}

func (k *KUDOClient) InstallOV(crd *unstructured.Unstructured) (*v1beta1.OperatorVersion, error) {
	repoConfig := repo.Configuration{
		URL:  k.repoURL,
		Name: "kudoBridge",
	}

	repository, err := repo.NewClient(&repoConfig)
	if err != nil {
		return nil, err
	}

	r := resolver.New(repository)
	p, err := r.Resolve(k.kudoPackageName, k.appVersion, k.version)
	if err != nil {
		return nil, err
	}

	k.resources = p.Resources
	installOpts := install.Options{
		SkipInstance:    true,
		CreateNamespace: false,
	}

	parameters := make(map[string]string)
	if err := install.Package(k.kc, crd.GetName(), crd.GetNamespace(), *k.resources, parameters, installOpts); err != nil {
		return nil, err
	}
	return k.kc.GetOperatorVersion(k.resources.OperatorVersion.GetName(), k.resources.OperatorVersion.GetNamespace())

}

func (k *KUDOClient) InstallOrUpdateInstance(crd *unstructured.Unstructured, ov *v1beta1.OperatorVersion, params map[string]string) error {
	instance, err := k.kc.GetInstance(crd.GetName(), crd.GetNamespace())
	if instance == nil && err == nil {
		// install Instance
		return k.InstallInstance(crd, ov, params)
	}
	if err != nil {
		return err
	}
	// update existing instance
	return k.upgrade(instance, crd, ov, params)
}

func (k *KUDOClient) InstallInstance(crd *unstructured.Unstructured, ov *v1beta1.OperatorVersion, params map[string]string) error {
	installOpts := install.Options{
		SkipInstance:    false,
		CreateNamespace: false,
	}
	if err := install.Package(k.kc, crd.GetName(), crd.GetNamespace(), *k.resources, params, installOpts); err != nil {
		return err
	}
	return k.MarkOwnerReference(crd)

}

func (k *KUDOClient) upgrade(instance *v1beta1.Instance, crd *unstructured.Unstructured, ov *v1beta1.OperatorVersion, params map[string]string) error {
	oldOv, err := k.kc.GetOperatorVersion(instance.Spec.OperatorVersion.Name, instance.GetNamespace())
	if err != nil {
		return err
	}
	if oldOv == nil {
		return fmt.Errorf("no OperatorVersion installed for Instance %s/%s", instance.GetNamespace(), instance.GetName())
	}

	oldVersion, err := semver.NewVersion(oldOv.Spec.Version)
	if err != nil {
		return err
	}
	newVersion, err := semver.NewVersion(ov.Spec.Version)
	if err != nil {
		return err
	}

	if newVersion.Equal(oldVersion) {
		// update the instance values only if parameters are changed
		if !reflect.DeepEqual(instance.Spec.Parameters, params) {
			log.Infof("updating instance %s/%s parameters", instance.GetNamespace(), instance.GetName())
			log.Infof("old parameters: %+v", instance.Spec.Parameters)
			log.Infof("new parameters: %+v", params)
			return k.kc.UpdateInstance(instance.GetName(), instance.GetNamespace(), nil, params, nil, false, 0)
		}
		return nil
	}

	return kudo.UpgradeOperatorVersion(k.kc, ov, instance.GetName(), instance.GetNamespace(), params)
}
func (k *KUDOClient) MarkOwnerReference(crd *unstructured.Unstructured) error {
	instance, err := k.kc.GetInstance(crd.GetName(), crd.GetNamespace())
	if err != nil {
		return err
	}
	instance.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: crd.GetAPIVersion(),
			Kind:       crd.GetKind(),
			Name:       crd.GetName(),
			UID:        crd.GetUID(),
		},
	}
	_, err = k.c.KudoClient.KudoV1beta1().Instances(crd.GetNamespace()).Update(instance)
	return err
}
