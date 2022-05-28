package services

import (
	"sync"
	"time"

	keasClientSet "github.com/projectkeas/crds/pkg/client/clientset/versioned"
	keasClient "github.com/projectkeas/crds/pkg/client/informers/externalversions"
	"github.com/projectkeas/sdks-service/configuration"
)

var (
	informer keasClient.SharedInformerFactory
	lock     = &sync.Mutex{}
)

func GetInformer() keasClient.SharedInformerFactory {
	if informer != nil {
		return informer
	}

	// lock ensures that we only ever have one factory
	// the lock is after the initial check as it can be lock free for most cases
	lock.Lock()
	defer lock.Unlock()

	// there's a slight race condition between the first check and the lock,
	// so check again inside a synchronised context
	if informer != nil {
		return informer
	}

	config, err := configuration.GetKubernetesConfig()
	if err != nil {
		panic(err)
	}

	client, err := keasClientSet.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	informer = keasClient.NewSharedInformerFactory(client, 5*time.Minute)
	return informer
}
