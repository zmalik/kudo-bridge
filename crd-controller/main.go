package main

import (
	"context"
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/zmalik/kudo-bridge/crd-controller/pkg/client"
	"github.com/zmalik/kudo-bridge/crd-controller/pkg/watcher"
)

var (
	groupVersion string
	kind         string
	namespace    string
)

func main() {
	log.Infof("bootstrapping CRD Watcher...")
	clientSet, err := client.GetKubeClient()
	if err != nil {
		log.Fatalf("failed to get kube client: %v", err)
		return
	}

	if groupVersion == "" || kind == "" {
		log.Fatalf("missing groupversion of kind to watch [groupVersion=%s] [kind=%s]", groupVersion, kind)
		return
	}
	cont := watcher.NewController(clientSet, groupVersion, kind, namespace)
	cont.Run(context.Background())
}

func init() {
	flag.StringVar(&groupVersion, "group-version", "", "groupversion to watch")
	flag.StringVar(&kind, "kind", "", "kind to watch")
	flag.StringVar(&namespace, "ns", "", "namespace to watch")
	flag.Parse()

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
}
