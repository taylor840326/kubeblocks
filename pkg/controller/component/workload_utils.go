/*
Copyright (C) 2022-2024 ApeCloud Co., Ltd

This file is part of KubeBlocks project

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package component

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	workloads "github.com/apecloud/kubeblocks/apis/workloads/v1alpha1"
	"github.com/apecloud/kubeblocks/pkg/constant"
	"github.com/apecloud/kubeblocks/pkg/controller/instanceset"
	"github.com/apecloud/kubeblocks/pkg/generics"
	viper "github.com/apecloud/kubeblocks/pkg/viperx"
)

func ListOwnedWorkloads(ctx context.Context, cli client.Reader, namespace, clusterName, compName string) ([]*workloads.InstanceSet, error) {
	return listWorkloads(ctx, cli, namespace, clusterName, compName)
}

func ListOwnedPods(ctx context.Context, cli client.Reader, namespace, clusterName, compName string,
	opts ...client.ListOption) ([]*corev1.Pod, error) {
	return listPods(ctx, cli, namespace, clusterName, compName, nil, opts...)
}

func ListOwnedPodsWithRole(ctx context.Context, cli client.Reader, namespace, clusterName, compName, role string,
	opts ...client.ListOption) ([]*corev1.Pod, error) {
	roleLabel := map[string]string{constant.RoleLabelKey: role}
	return listPods(ctx, cli, namespace, clusterName, compName, roleLabel, opts...)
}

func ListOwnedPVCs(ctx context.Context, cli client.Reader, namespace, clusterName, compName string,
	opts ...client.ListOption) ([]*corev1.PersistentVolumeClaim, error) {
	labels := constant.GetComponentWellKnownLabels(clusterName, compName)
	if opts == nil {
		opts = make([]client.ListOption, 0)
	}
	opts = append(opts, inDataContext())
	return listObjWithLabelsInNamespace(ctx, cli, generics.PersistentVolumeClaimSignature, namespace, labels, opts...)
}

func ListOwnedServices(ctx context.Context, cli client.Reader, namespace, clusterName, compName string,
	opts ...client.ListOption) ([]*corev1.Service, error) {
	labels := constant.GetComponentWellKnownLabels(clusterName, compName)
	if opts == nil {
		opts = make([]client.ListOption, 0)
	}
	opts = append(opts, inDataContext())
	return listObjWithLabelsInNamespace(ctx, cli, generics.ServiceSignature, namespace, labels, opts...)
}

// GetMinReadySeconds gets the underlying workload's minReadySeconds of the component.
func GetMinReadySeconds(ctx context.Context, cli client.Client, cluster appsv1.Cluster, compName string) (minReadySeconds int32, err error) {
	var its []*workloads.InstanceSet
	its, err = listWorkloads(ctx, cli, cluster.Namespace, cluster.Name, compName)
	if err != nil {
		return
	}
	if len(its) > 0 {
		minReadySeconds = its[0].Spec.MinReadySeconds
		return
	}
	return minReadySeconds, err
}

func listWorkloads(ctx context.Context, cli client.Reader, namespace, clusterName, compName string) ([]*workloads.InstanceSet, error) {
	labels := constant.GetComponentWellKnownLabels(clusterName, compName)
	return listObjWithLabelsInNamespace(ctx, cli, generics.InstanceSetSignature, namespace, labels)
}

func listPods(ctx context.Context, cli client.Reader, namespace, clusterName, compName string,
	labels map[string]string, opts ...client.ListOption) ([]*corev1.Pod, error) {
	if labels == nil {
		labels = constant.GetComponentWellKnownLabels(clusterName, compName)
	} else {
		maps.Copy(labels, constant.GetComponentWellKnownLabels(clusterName, compName))
	}
	if opts == nil {
		opts = make([]client.ListOption, 0)
	}
	opts = append(opts, inDataContext())
	return listObjWithLabelsInNamespace(ctx, cli, generics.PodSignature, namespace, labels, opts...)
}

func listObjWithLabelsInNamespace[T generics.Object, PT generics.PObject[T], L generics.ObjList[T], PL generics.PObjList[T, L]](
	ctx context.Context, cli client.Reader, _ func(T, PT, L, PL), namespace string, labels client.MatchingLabels, opts ...client.ListOption) ([]PT, error) {
	if opts == nil {
		opts = make([]client.ListOption, 0)
	}
	opts = append(opts, []client.ListOption{labels, client.InNamespace(namespace)}...)

	var objList L
	if err := cli.List(ctx, PL(&objList), opts...); err != nil {
		return nil, err
	}

	objs := make([]PT, 0)
	items := reflect.ValueOf(&objList).Elem().FieldByName("Items").Interface().([]T)
	for i := range items {
		objs = append(objs, &items[i])
	}
	return objs, nil
}

// GenerateAllPodNames generate all pod names for a component.
func GenerateAllPodNames(
	compReplicas int32,
	instances []appsv1.InstanceTemplate,
	offlineInstances []string,
	clusterName,
	fullCompName string) ([]string, error) {
	workloadName := constant.GenerateWorkloadNamePattern(clusterName, fullCompName)
	var templates []instanceset.InstanceTemplate
	for i := range instances {
		templates = append(templates, &workloads.InstanceTemplate{
			Name:     instances[i].Name,
			Replicas: instances[i].Replicas,
		})
	}
	return instanceset.GenerateAllInstanceNames(workloadName, compReplicas, templates, offlineInstances, workloads.Ordinals{})
}

// GenerateAllPodNamesToSet generate all pod names for a component
// and return a set which key is the pod name and value is a template name.
func GenerateAllPodNamesToSet(
	compReplicas int32,
	instances []appsv1.InstanceTemplate,
	offlineInstances []string,
	clusterName,
	fullCompName string) (map[string]string, error) {
	instanceNames, err := GenerateAllPodNames(compReplicas, instances, offlineInstances, clusterName, fullCompName)
	if err != nil {
		return nil, err
	}
	// key: podName, value: templateName
	podSet := map[string]string{}
	for _, insName := range instanceNames {
		podSet[insName] = appsv1.GetInstanceTemplateName(clusterName, fullCompName, insName)
	}
	return podSet, nil
}

func GetTemplateNameAndOrdinal(workloadName, podName string) (string, int32, error) {
	podSuffix := strings.Replace(podName, workloadName+"-", "", 1)
	suffixArr := strings.Split(podSuffix, "-")
	templateName := ""
	indexStr := ""
	if len(suffixArr) == 2 {
		templateName = suffixArr[0]
		indexStr = suffixArr[1]
	} else {
		indexStr = suffixArr[0]
	}
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return "", 0, fmt.Errorf("failed to obtain pod ordinal")
	}
	return templateName, int32(index), nil
}

func PodFQDN(namespace, compName, podName string) string {
	return fmt.Sprintf("%s.%s-headless.%s.svc.%s", podName, compName, namespace, clusterDomain())
}

func clusterDomain() string {
	return viper.GetString(constant.KubernetesClusterDomainEnv)
}
