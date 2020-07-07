package bridge

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/zmalik/kudo-bridge/bridge-controller/pkg/apis/kudobridge/v1alpha1"
	"github.com/zmalik/kudo-bridge/bridge-controller/pkg/client"
	"github.com/zmalik/kudo-bridge/bridge-controller/pkg/generated/clientset/versioned/scheme"
)

const (
	finalizerName = "finalizer.bridge.kudo.dev"
)

type Bridge struct {
	*client.Client
}

func (b *Bridge) Process(ro runtime.Object) error {
	ctx := context.TODO()
	if ro == nil {
		// Event was deleted
		return nil
	}

	bi, ok := ro.(*v1alpha1.BridgeInstance)
	if !ok {
		log.Infoln("cannot cast the object to ExternalService", ro)
	}

	// marked for deletion
	if !bi.DeletionTimestamp.IsZero() {
		if err := b.RemoveFinalizer(bi); err != nil {
			return err
		}
		return nil
	}

	err := b.validateCRD(bi)
	if err != nil {
		return err
	}

	err = setGVKFromScheme(bi)
	if err != nil {
		return fmt.Errorf("could not set GroupVerionKind for %s/%s: %v", bi.GetNamespace(), bi.GetName(), err)
	}

	_, err = b.KubeClient.AppsV1().Deployments(bi.GetNamespace()).Get(ctx, bi.GetName(), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		// create sa/role
		sa, err := b.createSARole(bi)
		if err != nil {
			log.Errorf("Error creating service account the KUDO Bridge %s/%s: %v", bi.Namespace, bi.Name, err)
			return err
		}
		// create deployment
		newDeployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bi.GetName(),
				Namespace: bi.GetNamespace(),
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: bi.APIVersion,
						Kind:       bi.Kind,
						Name:       bi.GetName(),
						UID:        bi.GetUID(),
					},
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"kudobridge.dev": bi.GetName()},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"kudobridge.dev": bi.GetName()}},
					Spec: corev1.PodSpec{
						ServiceAccountName: sa.GetName(),
						Containers: []corev1.Container{
							{
								Name:            "crd-controller",
								Image:           "zmalikshxil/kudo-crd-controller:0.0.1-alpha",
								ImagePullPolicy: corev1.PullAlways,
								Env: []corev1.EnvVar{
									{
										Name:  "GROUP_VERSION",
										Value: bi.Spec.CRDSpec.GetAPIVersion(),
									},
									{
										Name:  "KIND",
										Value: bi.Spec.CRDSpec.GetKind(),
									},
								},
								Command: []string{"/root/crd-controller"},
								Args: []string{
									fmt.Sprintf("-group-version=%s", bi.Spec.CRDSpec.GetAPIVersion()),
									fmt.Sprintf("-kind=%s", bi.Spec.CRDSpec.GetKind()),
								},
							},
						},
					},
				},
			},
		}
		dep, err := b.KubeClient.AppsV1().Deployments(bi.GetNamespace()).Create(context.TODO(), newDeployment, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("Error creating deployment for the KUDO Bridge %s/%s: %v", bi.Namespace, bi.Name, err)
			return err
		}
		log.Infof("CRD Controller deployment %s/%s created for ExternalService %s/%s", dep.GetNamespace(), dep.GetName(), bi.Namespace, bi.Name)

	} else if err != nil {
		log.Errorf("Error fetching deployment for the KUDO Bridge %s/%s: %v", bi.Namespace, bi.Name, err)
		return err
	} else {
		log.Infof("deployment %s/%s already exists, updates aren't supported yet", bi.Namespace, bi.Name)
	}
	return nil
}

func (b *Bridge) createSARole(bi *v1alpha1.BridgeInstance) (*corev1.ServiceAccount, error) {
	sa, err := b.KubeClient.CoreV1().ServiceAccounts(bi.GetNamespace()).Get(context.TODO(), bi.GetName(), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		sa = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bi.GetName(),
				Namespace: bi.GetNamespace(),
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: bi.APIVersion,
						Kind:       bi.Kind,
						Name:       bi.GetName(),
						UID:        bi.GetUID(),
					},
				},
			},
		}
		sa, err = b.KubeClient.CoreV1().ServiceAccounts(bi.GetNamespace()).Create(context.TODO(), sa, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("Error creating service account the KUDO Bridge %s/%s: %v", bi.Namespace, bi.Name, err)
			return nil, err
		}
		log.Infof("ServiceAccount %s/%s created for ExternalService %s/%s", sa.GetNamespace(), sa.GetName(), bi.Namespace, bi.Name)
	}
	return sa, b.createRBAC(bi)
}

func (b *Bridge) createRBAC(bi *v1alpha1.BridgeInstance) error {
	_, err := b.KubeClient.RbacV1().Roles(bi.GetNamespace()).Get(context.TODO(), bi.GetName(), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		role := &v1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bi.GetName(),
				Namespace: bi.GetNamespace(),
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: bi.APIVersion,
						Kind:       bi.Kind,
						Name:       bi.GetName(),
						UID:        bi.GetUID(),
					},
				},
			},
			Rules: []v1.PolicyRule{
				{
					Verbs:         []string{"get", "watch", "list"},
					Resources:     []string{"bridgeinstances"},
					APIGroups:     []string{"kudobridge.dev"},
					ResourceNames: []string{},
				},
				{
					Verbs:         []string{"get", "watch", "list", "create", "update", "patch", "delete"},
					Resources:     []string{"operatorversions", "instances", "operators"},
					APIGroups:     []string{"kudo.dev"},
					ResourceNames: []string{},
				},
			},
		}
		role, err = b.KubeClient.RbacV1().Roles(bi.GetNamespace()).Create(context.TODO(), role, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("Error creating Role for the KUDO Bridge %s/%s: %v", bi.Namespace, bi.Name, err)
			return err
		}
		log.Infof("Role %s/%s created for ExternalService %s/%s", role.GetNamespace(), role.GetName(), bi.Namespace, bi.Name)
	} else if err != nil {
		return err
	}

	clusterRoleName := fmt.Sprintf("kudobridge-%s-%s", bi.GetNamespace(), bi.GetName())
	_, err = b.KubeClient.RbacV1().ClusterRoles().Get(context.TODO(), clusterRoleName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		clusterRole := &v1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: clusterRoleName,
				Labels: map[string]string{
					"kudobridge.dev/CRDKind":   bi.Spec.CRDSpec.GetKind(),
					"kudobridge.dev/name":      bi.GetName(),
					"kudobridge.dev/namespace": bi.GetNamespace(),
				},
			},
			Rules: []v1.PolicyRule{
				{
					Verbs:         []string{"get", "watch", "list"},
					Resources:     []string{"*"},
					APIGroups:     []string{bi.Spec.CRDSpec.GroupVersionKind().GroupVersion().Group},
					ResourceNames: []string{},
				},
				{
					Verbs:         []string{"get", "watch", "list"},
					Resources:     []string{"customresourcedefinitions"},
					APIGroups:     []string{"apiextensions.k8s.io"},
					ResourceNames: []string{},
				},
			},
		}
		clusterRole, err = b.KubeClient.RbacV1().ClusterRoles().Create(context.TODO(), clusterRole, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("Error creating ClusterRole for the KUDO Bridge %s/%s: %v", bi.Namespace, bi.Name, err)
			return err
		}
		log.Infof("ClusterRole %s created for ExternalService %s/%s", clusterRole.GetName(), bi.Namespace, bi.Name)
	} else if err != nil {
		return err
	}

	_, err = b.KubeClient.RbacV1().RoleBindings(bi.GetNamespace()).Get(context.TODO(), bi.GetName(), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		rolebinding := &v1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bi.GetName(),
				Namespace: bi.GetNamespace(),
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: bi.APIVersion,
						Kind:       bi.Kind,
						Name:       bi.GetName(),
						UID:        bi.GetUID(),
					},
				},
			},
			Subjects: []v1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      bi.GetName(),
					Namespace: bi.GetNamespace(),
				},
			},
			RoleRef: v1.RoleRef{
				APIGroup: v1.GroupName,
				Kind:     "Role",
				Name:     bi.GetName(),
			},
		}
		rolebinding, err = b.KubeClient.RbacV1().RoleBindings(bi.GetNamespace()).Create(context.TODO(), rolebinding, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("Error creating RoleBinding for the KUDO Bridge %s/%s: %v", bi.Namespace, bi.Name, err)
			return err
		}
		log.Infof("RoleBinding %s/%s created for ExternalService %s/%s", rolebinding.GetNamespace(), rolebinding.GetName(), bi.Namespace, bi.Name)
	} else if err != nil {
		return err
	}

	clusterRoleBindingName := fmt.Sprintf("kudobridge-%s-%s", bi.GetNamespace(), bi.GetName())
	_, err = b.KubeClient.RbacV1().ClusterRoleBindings().Get(context.TODO(), clusterRoleBindingName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		clusterRoleBinding := &v1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: clusterRoleBindingName,
				Labels: map[string]string{
					"kudobridge.dev/CRDKind":   bi.Spec.CRDSpec.GetKind(),
					"kudobridge.dev/name":      bi.GetName(),
					"kudobridge.dev/namespace": bi.GetNamespace(),
				},
			},
			Subjects: []v1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      bi.GetName(),
					Namespace: bi.GetNamespace(),
				},
			},
			RoleRef: v1.RoleRef{
				APIGroup: v1.GroupName,
				Kind:     "ClusterRole",
				Name:     clusterRoleName,
			},
		}
		clusterRoleBinding, err = b.KubeClient.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("Error creating ClusterRoleBinding for the KUDO Bridge %s/%s: %v", bi.Namespace, bi.Name, err)
			return err
		}
		log.Infof("ClusterRoleBinding %s created for ExternalService %s/%s", clusterRoleBinding.GetName(), bi.Namespace, bi.Name)
	} else if err != nil {
		return err
	}

	return nil
}
func (b *Bridge) validateCRD(bi *v1alpha1.BridgeInstance) error {
	_, err := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(b.Discovery)).RESTMapping(bi.Spec.CRDSpec.GroupVersionKind().GroupKind(), bi.Spec.CRDSpec.GroupVersionKind().Version)
	if err != nil {
		log.Errorf("Cannot watch the provided CRD :%v", err)
		return err
	}
	return nil
}
func (b *Bridge) RemoveFinalizer(bi *v1alpha1.BridgeInstance) error {
	if containsFinalizer(bi, finalizerName) {
		err := b.deleteClusterScopeResources(bi)
		if err != nil {
			log.Errorf("Cannot cleanup clusterScope resources :%v", err)
			return err
		}
		controllerutil.RemoveFinalizer(bi, finalizerName)
		_, err = b.Bridge.KudobridgeV1alpha1().BridgeInstances(bi.GetNamespace()).Update(context.TODO(), bi, metav1.UpdateOptions{})
		if err != nil {
			log.Errorf("Cannot remove finalizer %s from %s/%s: %v", finalizerName, bi.GetNamespace(), bi.GetName(), err)
			return err
		}
		return nil
	}
	return nil
}

func (b *Bridge) deleteClusterScopeResources(bi *v1alpha1.BridgeInstance) error {
	clusterRoleName := fmt.Sprintf("kudobridge-%s-%s", bi.GetNamespace(), bi.GetName())
	err := b.KubeClient.RbacV1().ClusterRoles().Delete(context.TODO(), clusterRoleName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		log.Errorf("Cannot delete the ClusterRole %s :%s", clusterRoleName, err)
		return err
	}
	log.Infof("Deleted the ClusterRole %s", clusterRoleName)
	clusterRoleBindingName := fmt.Sprintf("kudobridge-%s-%s", bi.GetNamespace(), bi.GetName())
	err = b.KubeClient.RbacV1().ClusterRoleBindings().Delete(context.TODO(), clusterRoleBindingName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		log.Errorf("Cannot delete the ClusterRoleBinding %s :%s", clusterRoleName, err)
		return err
	}
	log.Infof("Deleted the ClusterRoleBinding %s", clusterRoleBindingName)
	return nil
}

func setGVKFromScheme(object runtime.Object) error {
	gvks, unversioned, err := scheme.Scheme.ObjectKinds(object)
	if err != nil {
		return err
	}
	if len(gvks) == 0 {
		return fmt.Errorf("no ObjectKinds available for %T", object)
	}
	if !unversioned {
		object.GetObjectKind().SetGroupVersionKind(gvks[0])
	}
	return nil
}

// ContainsFinalizer checks an Object that the provided finalizer is present.
func containsFinalizer(o metav1.Object, finalizer string) bool {
	f := o.GetFinalizers()
	for _, e := range f {
		if e == finalizer {
			return true
		}
	}
	return false
}
