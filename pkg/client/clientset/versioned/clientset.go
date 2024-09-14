/*
Copyright (C) 2022-2024 ApeCloud Co., Ltd

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

package versioned

import (
	"fmt"
	"net/http"

	appsv1 "github.com/apecloud/kubeblocks/pkg/client/clientset/versioned/typed/apps/v1"
	appsv1alpha1 "github.com/apecloud/kubeblocks/pkg/client/clientset/versioned/typed/apps/v1alpha1"
	appsv1beta1 "github.com/apecloud/kubeblocks/pkg/client/clientset/versioned/typed/apps/v1beta1"
	dataprotectionv1alpha1 "github.com/apecloud/kubeblocks/pkg/client/clientset/versioned/typed/dataprotection/v1alpha1"
	extensionsv1alpha1 "github.com/apecloud/kubeblocks/pkg/client/clientset/versioned/typed/extensions/v1alpha1"
	workloadsv1 "github.com/apecloud/kubeblocks/pkg/client/clientset/versioned/typed/workloads/v1"
	workloadsv1alpha1 "github.com/apecloud/kubeblocks/pkg/client/clientset/versioned/typed/workloads/v1alpha1"
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	AppsV1alpha1() appsv1alpha1.AppsV1alpha1Interface
	AppsV1beta1() appsv1beta1.AppsV1beta1Interface
	AppsV1() appsv1.AppsV1Interface
	DataprotectionV1alpha1() dataprotectionv1alpha1.DataprotectionV1alpha1Interface
	ExtensionsV1alpha1() extensionsv1alpha1.ExtensionsV1alpha1Interface
	WorkloadsV1alpha1() workloadsv1alpha1.WorkloadsV1alpha1Interface
	WorkloadsV1() workloadsv1.WorkloadsV1Interface
}

// Clientset contains the clients for groups.
type Clientset struct {
	*discovery.DiscoveryClient
	appsV1alpha1           *appsv1alpha1.AppsV1alpha1Client
	appsV1beta1            *appsv1beta1.AppsV1beta1Client
	appsV1                 *appsv1.AppsV1Client
	dataprotectionV1alpha1 *dataprotectionv1alpha1.DataprotectionV1alpha1Client
	extensionsV1alpha1     *extensionsv1alpha1.ExtensionsV1alpha1Client
	workloadsV1alpha1      *workloadsv1alpha1.WorkloadsV1alpha1Client
	workloadsV1            *workloadsv1.WorkloadsV1Client
}

// AppsV1alpha1 retrieves the AppsV1alpha1Client
func (c *Clientset) AppsV1alpha1() appsv1alpha1.AppsV1alpha1Interface {
	return c.appsV1alpha1
}

// AppsV1beta1 retrieves the AppsV1beta1Client
func (c *Clientset) AppsV1beta1() appsv1beta1.AppsV1beta1Interface {
	return c.appsV1beta1
}

// AppsV1 retrieves the AppsV1Client
func (c *Clientset) AppsV1() appsv1.AppsV1Interface {
	return c.appsV1
}

// DataprotectionV1alpha1 retrieves the DataprotectionV1alpha1Client
func (c *Clientset) DataprotectionV1alpha1() dataprotectionv1alpha1.DataprotectionV1alpha1Interface {
	return c.dataprotectionV1alpha1
}

// ExtensionsV1alpha1 retrieves the ExtensionsV1alpha1Client
func (c *Clientset) ExtensionsV1alpha1() extensionsv1alpha1.ExtensionsV1alpha1Interface {
	return c.extensionsV1alpha1
}

// WorkloadsV1alpha1 retrieves the WorkloadsV1alpha1Client
func (c *Clientset) WorkloadsV1alpha1() workloadsv1alpha1.WorkloadsV1alpha1Interface {
	return c.workloadsV1alpha1
}

// WorkloadsV1 retrieves the WorkloadsV1Client
func (c *Clientset) WorkloadsV1() workloadsv1.WorkloadsV1Interface {
	return c.workloadsV1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfig will generate a rate-limiter in configShallowCopy.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c

	if configShallowCopy.UserAgent == "" {
		configShallowCopy.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	// share the transport between all clients
	httpClient, err := rest.HTTPClientFor(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	return NewForConfigAndClient(&configShallowCopy, httpClient)
}

// NewForConfigAndClient creates a new Clientset for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfigAndClient will generate a rate-limiter in configShallowCopy.
func NewForConfigAndClient(c *rest.Config, httpClient *http.Client) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		if configShallowCopy.Burst <= 0 {
			return nil, fmt.Errorf("burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}

	var cs Clientset
	var err error
	cs.appsV1alpha1, err = appsv1alpha1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	cs.appsV1beta1, err = appsv1beta1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	cs.appsV1, err = appsv1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	cs.dataprotectionV1alpha1, err = dataprotectionv1alpha1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	cs.extensionsV1alpha1, err = extensionsv1alpha1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	cs.workloadsV1alpha1, err = workloadsv1alpha1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	cs.workloadsV1, err = workloadsv1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	cs, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.appsV1alpha1 = appsv1alpha1.New(c)
	cs.appsV1beta1 = appsv1beta1.New(c)
	cs.appsV1 = appsv1.New(c)
	cs.dataprotectionV1alpha1 = dataprotectionv1alpha1.New(c)
	cs.extensionsV1alpha1 = extensionsv1alpha1.New(c)
	cs.workloadsV1alpha1 = workloadsv1alpha1.New(c)
	cs.workloadsV1 = workloadsv1.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
