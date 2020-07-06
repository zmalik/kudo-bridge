module github.com/zmalik/kudo-bridge

go 1.14

require (
	github.com/Jeffail/gabs/v2 v2.5.1
	github.com/Masterminds/semver v1.5.0
	github.com/devopsfaith/flatmap v0.0.0-20200601181759-8521186182fc
	github.com/kudobuilder/kudo v0.14.0
	github.com/prometheus/common v0.4.1
	github.com/sirupsen/logrus v1.4.2
	k8s.io/api v0.17.7
	k8s.io/apiextensions-apiserver v0.17.7
	k8s.io/apimachinery v0.17.7
	k8s.io/client-go v0.17.7
	k8s.io/code-generator v0.17.7
	sigs.k8s.io/controller-runtime v0.5.7
)

replace k8s.io/code-generator v0.17.7 => github.com/kudobuilder/code-generator v0.17.4-beta.0.0.20200316162450-cc91a9201457
