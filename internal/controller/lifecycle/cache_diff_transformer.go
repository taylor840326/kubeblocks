/*
Copyright ApeCloud, Inc.

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

package lifecycle

import (
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "github.com/apecloud/kubeblocks/apis/apps/v1alpha1"
	"github.com/apecloud/kubeblocks/internal/controller/graph"
	intctrlutil "github.com/apecloud/kubeblocks/internal/controllerutil"
)

// cacheDiffTransformer reads all Vertex.Obj in cache and compute the diff DAG.
type cacheDiffTransformer struct {
	cc  compoundCluster
	cli client.Client
	ctx intctrlutil.RequestCtx
}

func ownKinds() []client.ObjectList {
	return []client.ObjectList{
		&appsv1.StatefulSetList{},
		&appsv1.DeploymentList{},
		&corev1.ServiceList{},
		&corev1.SecretList{},
		&corev1.ConfigMapList{},
		&corev1.PersistentVolumeClaimList{},
		&policyv1.PodDisruptionBudgetList{},
	}
}

// read all objects owned by our cluster
func (c *cacheDiffTransformer) readCacheSnapshot() (clusterSnapshot, error) {
	scheme, _ := appsv1alpha1.SchemeBuilder.Build()

	// list what kinds of object cluster owns
	kinds := ownKinds()
	snapshot := make(clusterSnapshot)
	ml := client.MatchingLabels{intctrlutil.AppInstanceLabelKey: c.cc.cluster.GetName()}
	inNS := client.InNamespace(c.cc.cluster.Namespace)
	for _, list := range kinds {
		if err := c.cli.List(c.ctx.Ctx, list, inNS, ml); err != nil {
			return nil, err
		}
		// reflect get list.Items
		items := reflect.ValueOf(list).Elem().FieldByName("Items")
		l := items.Len()
		for i := 0; i < l; i++ {
			// get the underlying object
			object := items.Index(i).Addr().Interface().(client.Object)
			// put to snapshot if owned by our cluster
			if isOwnerOf(c.cc.cluster, object, scheme) {
				name := getGVKName(object)
				snapshot[name] = object
			}
		}
	}

	return snapshot, nil
}

func (c *cacheDiffTransformer) Transform(dag *graph.DAG) error {
	// get the old snapshot
	oldSnapshot, err := c.readCacheSnapshot()
	if err != nil {
		return err
	}

	// we have target snapshot in dag
	// now do the heavy lift:
	// compute the diff between cache and target spec and generate the plan
	newNameVertices := make(map[gvkName]graph.Vertex)
	for _, vertex := range dag.Vertices() {
		v, _ := vertex.(*lifecycleVertex)
		name := getGVKName(v.obj)
		newNameVertices[name] = vertex
	}

	oldNameSet := sets.KeySet(oldSnapshot)
	newNameSet := sets.KeySet(newNameVertices)

	// case cluster Deletion
	if !c.cc.cluster.DeletionTimestamp.IsZero() {
		// TODO: handle deletion
		return nil
	}
	// case cluster Creation or Update
	deleteSet := oldNameSet.Difference(newNameSet)
	createSet := newNameSet.Difference(oldNameSet)
	updateSet := newNameSet.Intersection(oldNameSet)
	// dag root is our cluster object
	root := dag.Root()
	if root == nil {
		return fmt.Errorf("root vertex not found: %v", dag)
	}
	for name := range deleteSet {
		v := &lifecycleVertex{
			obj:    oldSnapshot[name],
			action: actionPtr(DELETE),
		}
		dag.AddVertex(v)
		dag.Connect(root, v)
	}
	for name := range createSet {
		v, _ := newNameVertices[name].(*lifecycleVertex)
		v.action = actionPtr(CREATE)
	}
	for name := range updateSet {
		v, _ := newNameVertices[name].(*lifecycleVertex)
		v.action = actionPtr(UPDATE)
	}

	return nil
}
