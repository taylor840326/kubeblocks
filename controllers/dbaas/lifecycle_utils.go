/*
Copyright 2022.

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

package dbaas

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/leaanthony/debme"
	"github.com/sethvargo/go-password/password"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	dbaasv1alpha1 "github.com/apecloud/kubeblocks/apis/dbaas/v1alpha1"
	intctrlutil "github.com/apecloud/kubeblocks/internal/controllerutil"
)

type createParams struct {
	clusterDefinition *dbaasv1alpha1.ClusterDefinition
	cluster           *dbaasv1alpha1.Cluster
	component         *Component
	applyObjs         *[]client.Object
	cacheCtx          *map[string]interface{}
}

const (
	dbaasPrefix = "OPENDBAAS"
)

var (
	//go:embed cue/*
	cueTemplates embed.FS
)

func (c createParams) getCacheBytesValue(key string, valueCreator func() ([]byte, error)) ([]byte, error) {
	vIf, ok := (*c.cacheCtx)[key]
	if ok {
		return vIf.([]byte), nil
	}
	v, err := valueCreator()
	if err != nil {
		return nil, err
	}
	(*c.cacheCtx)[key] = v
	return v, err
}

func (c createParams) getCacheCUETplValue(key string, valueCreator func() (*intctrlutil.CUETpl, error)) (*intctrlutil.CUETpl, error) {
	vIf, ok := (*c.cacheCtx)[key]
	if ok {
		return vIf.(*intctrlutil.CUETpl), nil
	}
	v, err := valueCreator()
	if err != nil {
		return nil, err
	}
	(*c.cacheCtx)[key] = v
	return v, err
}

func getAppVersionComponentByType(components []dbaasv1alpha1.AppVersionComponent, typeName string) *dbaasv1alpha1.AppVersionComponent {
	for _, component := range components {
		if component.Type == typeName {
			return &component
		}
	}
	return nil
}

func getClusterComponentsByType(components []dbaasv1alpha1.ClusterComponent, typeName string) []*dbaasv1alpha1.ClusterComponent {
	comps := []*dbaasv1alpha1.ClusterComponent{}
	for _, component := range components {
		if component.Type == typeName {
			comps = append(comps, &component)
		}
	}
	return comps
}

func getContainerByName(containers []corev1.Container, name string) (int, *corev1.Container) {
	for i, container := range containers {
		if container.Name == name {
			return i, &container
		}
	}
	return -1, nil
}

func toK8sVolumeClaimTemplate(template dbaasv1alpha1.ClusterComponentVolumeClaimTemplate) corev1.PersistentVolumeClaimTemplate {
	t := corev1.PersistentVolumeClaimTemplate{}
	t.ObjectMeta.Name = template.Name
	t.Spec = template.Spec
	return t
}

func toK8sVolumeClaimTemplates(templates []dbaasv1alpha1.ClusterComponentVolumeClaimTemplate) []corev1.PersistentVolumeClaimTemplate {
	ts := []corev1.PersistentVolumeClaimTemplate{}
	for _, template := range templates {
		ts = append(ts, toK8sVolumeClaimTemplate(template))
	}
	return ts
}

func mergeComponents(
	clusterDef *dbaasv1alpha1.ClusterDefinition,
	clusterDefComp *dbaasv1alpha1.ClusterDefinitionComponent,
	appVerComp *dbaasv1alpha1.AppVersionComponent,
	clusterComp *dbaasv1alpha1.ClusterComponent) *Component {
	if clusterDefComp == nil {
		return nil
	}
	component := &Component{
		ClusterDefName:  clusterDef.Name,
		ClusterType:     clusterDef.Spec.Type,
		Name:            clusterDefComp.TypeName,
		Type:            clusterDefComp.TypeName,
		MinAvailable:    clusterDefComp.MinAvailable,
		MaxAvailable:    clusterDefComp.MaxAvailable,
		DefaultReplicas: clusterDefComp.DefaultReplicas,
		Replicas:        clusterDefComp.DefaultReplicas,
		AntiAffinity:    clusterDefComp.AntiAffinity,
		ComponentType:   clusterDefComp.ComponentType,
		ConsensusSpec:   clusterDefComp.ConsensusSpec,
		PodSpec:         clusterDefComp.PodSpec,
		Service:         clusterDefComp.Service,
		ReadonlyService: clusterDefComp.ReadonlyService,
		Scripts:         clusterDefComp.Scripts,
	}
	if clusterComp != nil {
		component.Name = clusterComp.Name
	}

	if appVerComp != nil && appVerComp.PodSpec.Containers != nil {
		for _, container := range appVerComp.PodSpec.Containers {
			i, c := getContainerByName(component.PodSpec.Containers, container.Name)
			if c != nil {
				if container.Image != "" {
					component.PodSpec.Containers[i].Image = container.Image
				}
				if len(container.Command) != 0 {
					component.PodSpec.Containers[i].Command = container.Command
				}
				if len(container.Args) != 0 {
					component.PodSpec.Containers[i].Args = container.Args
				}
				if container.WorkingDir != "" {
					component.PodSpec.Containers[i].WorkingDir = container.WorkingDir
				}
				if len(container.Ports) != 0 {
					component.PodSpec.Containers[i].Ports = container.Ports
				}
				if len(container.EnvFrom) != 0 {
					component.PodSpec.Containers[i].EnvFrom = container.EnvFrom
				}
				if len(container.Env) != 0 {
					component.PodSpec.Containers[i].Env = container.Env
				}
				if container.Resources.Limits != nil || container.Resources.Requests != nil {
					component.PodSpec.Containers[i].Resources = container.Resources
				}
				if len(container.VolumeMounts) != 0 {
					component.PodSpec.Containers[i].VolumeMounts = container.VolumeMounts
				}
				if len(container.VolumeDevices) != 0 {
					component.PodSpec.Containers[i].VolumeDevices = container.VolumeDevices
				}
				if container.LivenessProbe != nil {
					component.PodSpec.Containers[i].LivenessProbe = container.LivenessProbe
				}
				if container.ReadinessProbe != nil {
					component.PodSpec.Containers[i].ReadinessProbe = container.ReadinessProbe
				}
				if container.StartupProbe != nil {
					component.PodSpec.Containers[i].StartupProbe = container.StartupProbe
				}
				if container.Lifecycle != nil {
					component.PodSpec.Containers[i].Lifecycle = container.Lifecycle
				}
				if container.TerminationMessagePath != "" {
					component.PodSpec.Containers[i].TerminationMessagePath = container.TerminationMessagePath
				}
				if container.TerminationMessagePolicy != "" {
					component.PodSpec.Containers[i].TerminationMessagePolicy = container.TerminationMessagePolicy
				}
				if container.ImagePullPolicy != "" {
					component.PodSpec.Containers[i].ImagePullPolicy = container.ImagePullPolicy
				}
				if container.SecurityContext != nil {
					component.PodSpec.Containers[i].SecurityContext = container.SecurityContext
				}
			} else {
				component.PodSpec.Containers = append(component.PodSpec.Containers, container)
			}
		}
	}
	if clusterComp != nil {
		component.Name = clusterComp.Name

		// respect user's declaration
		if clusterComp.Replicas > 0 {
			component.Replicas = clusterComp.Replicas
		}

		if clusterComp.VolumeClaimTemplates != nil {
			component.VolumeClaimTemplates = toK8sVolumeClaimTemplates(clusterComp.VolumeClaimTemplates)
		}
		if clusterComp.Resources.Requests != nil || clusterComp.Resources.Limits != nil {
			component.PodSpec.Containers[0].Resources = clusterComp.Resources
		}

		// respect user's declaration
		if clusterComp.Service.Ports != nil {
			component.Service = clusterComp.Service
		}
	}
	if component.VolumeClaimTemplates == nil {
		for i := range component.PodSpec.Containers {
			component.PodSpec.Containers[i].VolumeMounts = nil
		}
	}
	return component
}

func buildClusterCreationTasks(
	clusterDefinition *dbaasv1alpha1.ClusterDefinition,
	appVersion *dbaasv1alpha1.AppVersion,
	cluster *dbaasv1alpha1.Cluster) (*intctrlutil.Task, error) {
	rootTask := intctrlutil.NewTask()

	applyObjs := make([]client.Object, 0, 3)
	cacheCtx := map[string]interface{}{}

	prepareSecretsTask := intctrlutil.NewTask()
	prepareSecretsTask.ExecFunction = prepareSecretObjs
	params := createParams{
		cluster:           cluster,
		clusterDefinition: clusterDefinition,
		applyObjs:         &applyObjs,
		cacheCtx:          &cacheCtx,
	}
	prepareSecretsTask.Context["exec"] = &params
	rootTask.SubTasks = append(rootTask.SubTasks, prepareSecretsTask)

	buildTask := func(component *Component) {
		componentTask := intctrlutil.NewTask()
		componentTask.ExecFunction = prepareComponentObjs
		iParams := params
		iParams.component = component
		componentTask.Context["exec"] = &iParams
		rootTask.SubTasks = append(rootTask.SubTasks, componentTask)
	}

	components := clusterDefinition.Spec.Components
	useDefaultComp := len(cluster.Spec.Components) == 0
	for _, component := range components {
		componentName := component.TypeName
		appVersionComponent := getAppVersionComponentByType(appVersion.Spec.Components, componentName)

		if useDefaultComp {
			buildTask(mergeComponents(clusterDefinition, &component, appVersionComponent, nil))
		} else {
			clusterComps := getClusterComponentsByType(cluster.Spec.Components, componentName)
			for _, clusterComp := range clusterComps {
				buildTask(mergeComponents(clusterDefinition, &component, appVersionComponent, clusterComp))
			}
		}
	}

	createObjsTask := intctrlutil.NewTask()
	createObjsTask.ExecFunction = checkedCreateObjs
	createObjsTask.Context["exec"] = &params
	rootTask.SubTasks = append(rootTask.SubTasks, createObjsTask)
	return &rootTask, nil
}

func checkedCreateObjs(ctx context.Context, cli client.Client, obj interface{}) error {
	params, ok := obj.(*createParams)
	if !ok {
		return fmt.Errorf("invalid arg")
	}

	if err := createOrReplaceResources(ctx, cli, params.cluster, *params.applyObjs); err != nil {
		return err
	}
	return nil
}

func prepareSecretObjs(ctx context.Context, cli client.Client, obj interface{}) error {
	params, ok := obj.(*createParams)
	if !ok {
		return fmt.Errorf("invalid arg")
	}

	secret, err := buildSecret(*params)
	if err != nil {
		return err
	}
	// must make sure secret resources are created before others
	*params.applyObjs = append(*params.applyObjs, secret)
	return nil
}

func prepareComponentObjs(ctx context.Context, cli client.Client, obj interface{}) error {
	params, ok := obj.(*createParams)
	if !ok {
		return fmt.Errorf("invalid arg")
	}

	switch params.component.ComponentType {
	case dbaasv1alpha1.Stateless:
		sts, err := buildDeploy(*params)
		if err != nil {
			return err
		}
		*params.applyObjs = append(*params.applyObjs, sts)
	case dbaasv1alpha1.Stateful:
		sts, err := buildSts(*params)
		if err != nil {
			return err
		}
		*params.applyObjs = append(*params.applyObjs, sts)

		svcs, err := buildHeadlessSvcs(*params, sts)
		if err != nil {
			return err
		}
		*params.applyObjs = append(*params.applyObjs, svcs...)
	case dbaasv1alpha1.Consensus:
		css, err := buildConsensusSet(*params)
		if err != nil {
			return err
		}
		css.GetLabels()
		*params.applyObjs = append(*params.applyObjs, css)

		svcs, err := buildHeadlessSvcs(*params, css)
		if err != nil {
			return err
		}
		*params.applyObjs = append(*params.applyObjs, svcs...)

	}

	pdb, err := buildPDB(*params)
	if err != nil {
		return err
	}
	*params.applyObjs = append(*params.applyObjs, pdb)

	if params.component.Service.Ports != nil {
		svc, err := buildSvc(*params)
		if err != nil {
			return err
		}
		if params.component.ComponentType == dbaasv1alpha1.Consensus {
			addSelectorLabels(svc, params.component, dbaasv1alpha1.ReadWrite)
		}
		*params.applyObjs = append(*params.applyObjs, svc)
	}

	if params.component.ReadonlyService.Ports != nil &&
		params.component.ComponentType == dbaasv1alpha1.Consensus {
		svc, err := buildSvc(*params)
		if err != nil {
			return err
		}
		addSelectorLabels(svc, params.component, dbaasv1alpha1.Readonly)
		*params.applyObjs = append(*params.applyObjs, svc)
	}

	return nil
}

func addSelectorLabels(service *corev1.Service, component *Component, accessMode dbaasv1alpha1.AccessMode) {
	addSelector := func(service *corev1.Service, member dbaasv1alpha1.ConsensusMember, accessMode dbaasv1alpha1.AccessMode) {
		if member.AccessMode == accessMode {
			service.Spec.Selector[consensusSetRoleLabelKey] = member.Name
		}
	}

	addSelector(service, component.ConsensusSpec.Leader, accessMode)
	addSelector(service, component.ConsensusSpec.Learner, accessMode)

	for _, member := range component.ConsensusSpec.Followers {
		addSelector(service, member, accessMode)
	}
}

func createOrReplaceResources(ctx context.Context,
	cli client.Client,
	cluster *dbaasv1alpha1.Cluster,
	objs []client.Object) error {
	scheme, _ := dbaasv1alpha1.SchemeBuilder.Build()
	for _, obj := range objs {
		if err := controllerutil.SetOwnerReference(cluster, obj, scheme); err != nil {
			return err
		}
		if err := cli.Create(ctx, obj); err == nil {
			continue
		} else if !apierrors.IsAlreadyExists(err) {
			return err
		}

		if !controllerutil.ContainsFinalizer(obj, dbClusterFinalizerName) {
			controllerutil.AddFinalizer(obj, dbClusterFinalizerName)
		}

		// Secret kind objects should only be applied once
		_, ok := obj.(*corev1.Secret)
		if ok {
			continue
		}
		// -

		key := client.ObjectKey{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
		}
		stsProto, ok := obj.(*appsv1.StatefulSet)
		if ok {
			stsObj := &appsv1.StatefulSet{}
			if err := cli.Get(ctx, key, stsObj); err != nil {
				return err
			}
			templateDiff := cmp.Diff(stsObj.Spec.Template, stsProto.Spec.Template)
			stsObj.Spec.Template = stsProto.Spec.Template
			stsObj.Spec.Replicas = stsProto.Spec.Replicas
			stsObj.Spec.UpdateStrategy = stsProto.Spec.UpdateStrategy
			if err := cli.Update(ctx, stsObj); err != nil {
				return err
			}
			// handle ConsensusSet Update
			if templateDiff != "" {
				_, err := handleConsensusSetUpdate(ctx, cli, cluster, stsObj)
				if err != nil {
					return err
				}
			}
			// check stsObj.Spec.VolumeClaimTemplates storage
			// request size and find attached PVC and patch request
			// storage size
			for _, vct := range stsObj.Spec.VolumeClaimTemplates {
				var vctProto *corev1.PersistentVolumeClaim
				for _, i := range stsProto.Spec.VolumeClaimTemplates {
					if i.Name == vct.Name {
						vctProto = &i
						break
					}
				}

				// REVIEW: how could VCT proto is nil?
				if vctProto == nil {
					continue
				}

				for i := *stsObj.Spec.Replicas - 1; i >= 0; i-- {
					pvc := &corev1.PersistentVolumeClaim{}
					pvcKey := types.NamespacedName{
						Namespace: key.Namespace,
						Name:      fmt.Sprintf("%s-%s-%d", vct.Name, stsObj.Name, i),
					}
					if err := cli.Get(ctx, pvcKey, pvc); err != nil {
						return err
					}
					if pvc.Spec.Resources.Requests[corev1.ResourceStorage] == vctProto.Spec.Resources.Requests[corev1.ResourceStorage] {
						continue
					}
					patch := client.MergeFrom(pvc.DeepCopy())
					pvc.Spec.Resources.Requests[corev1.ResourceStorage] = vctProto.Spec.Resources.Requests[corev1.ResourceStorage]
					if err := cli.Patch(ctx, pvc, patch); err != nil {
						return err
					}
				}
			}
			continue
		}
		deployProto, ok := obj.(*appsv1.Deployment)
		if ok {
			deployObj := &appsv1.Deployment{}
			if err := cli.Get(ctx, key, deployObj); err != nil {
				return err
			}
			deployObj.Spec = deployProto.Spec
			if err := cli.Update(ctx, deployObj); err != nil {
				return err
			}
			continue
		}
		svcProto, ok := obj.(*corev1.Service)
		if ok {
			svcObj := &corev1.Service{}
			if err := cli.Get(ctx, key, svcObj); err != nil {
				return err
			}
			svcObj.Spec = svcProto.Spec
			if err := cli.Update(ctx, svcObj); err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func handleConsensusSetUpdate(ctx context.Context, cli client.Client, cluster *dbaasv1alpha1.Cluster, stsObj *appsv1.StatefulSet) (bool, error) {
	// get typeName from stsObj.name
	typeName := getComponentTypeName(*cluster, *stsObj)

	// get component from ClusterDefinition by typeName
	component, err := getComponent(ctx, cli, cluster, typeName)
	if err != nil {
		return false, err
	}

	if component.ComponentType != dbaasv1alpha1.Consensus {
		return true, nil
	}

	// get podList owned by stsObj
	podList := &corev1.PodList{}
	if err := cli.List(ctx, podList,
		&client.ListOptions{Namespace: stsObj.Namespace},
		client.MatchingLabelsSelector{Selector: labels.Everything()}); err != nil {
		return false, err
	}
	pods := make([]corev1.Pod, 0)
	for _, pod := range podList.Items {
		if isMemberOf(stsObj, &pod) {
			pods = append(pods, pod)
		}
	}

	// get pod label and name, compute plan
	plan := generateUpdatePlan(ctx, cli, stsObj, pods, component)
	// execute plan
	return plan.walkOneStep()
}

func generateUpdatePlan(ctx context.Context, cli client.Client, stsObj *appsv1.StatefulSet, pods []corev1.Pod, component dbaasv1alpha1.ClusterDefinitionComponent) *Plan {
	plan := &Plan{}
	plan.Start = &Step{}
	plan.WalkFunc = func(obj interface{}) (bool, error) {
		pod := obj.(corev1.Pod)
		spec := stsObj.Spec.Template.Spec
		if cmp.Equal(pod.Spec, spec) {
			return false, nil
		}
		if err := cli.Delete(ctx, &pod); err != nil {
			return false, nil
		}

		return true, nil
	}

	// now all are followers
	leader := component.ConsensusSpec.Leader.Name
	learner := component.ConsensusSpec.Learner.Name
	noneFollowers := make(map[string]string)
	readonlyFollowers := make(map[string]string)
	readWriteFollowers := make(map[string]string)
	exist := "EXIST"
	for _, follower := range component.ConsensusSpec.Followers {
		switch follower.AccessMode {
		case dbaasv1alpha1.None:
			noneFollowers[follower.Name] = exist
		case dbaasv1alpha1.Readonly:
			readonlyFollowers[follower.Name] = exist
		case dbaasv1alpha1.ReadWrite:
			readWriteFollowers[follower.Name] = exist
		}
	}

	// make a Serial pod list
	sort.SliceStable(pods, func(i, j int) bool {
		roleI := pods[i].Labels[consensusSetRoleLabelKey]
		roleJ := pods[j].Labels[consensusSetRoleLabelKey]
		if roleI == learner {
			return true
		}
		if roleJ == learner {
			return false
		}
		if roleI == leader {
			return false
		}
		if roleJ == leader {
			return true
		}
		if noneFollowers[roleI] == exist {
			return true
		}
		if noneFollowers[roleJ] == exist {
			return false
		}
		if readonlyFollowers[roleI] == exist {
			return true
		}
		if readonlyFollowers[roleJ] == exist {
			return false
		}
		if readWriteFollowers[roleI] == exist {
			return true
		}

		return false
	})

	// generate plan by updateStrategy
	switch component.ConsensusSpec.UpdateStrategy {
	case dbaasv1alpha1.Serial:
		// learner -> followers(none->readonly->readwrite) -> leader
		start := plan.Start
		for _, pod := range pods {
			nextStep := &Step{}
			nextStep.Obj = pod
			start.NextSteps = append(start.NextSteps, nextStep)
			start = nextStep
		}
	case dbaasv1alpha1.Parallel:
		// leader & followers & learner
		start := plan.Start
		for _, pod := range pods {
			nextStep := &Step{}
			nextStep.Obj = pod
			start.NextSteps = append(start.NextSteps, nextStep)
		}
	case dbaasv1alpha1.BestEffortParallel:
		// learner & 1/2 followers -> 1/2 followers -> leader
		start := plan.Start
		// append learner
		index := 0
		for _, pod := range pods {
			if pod.Labels[consensusSetRoleLabelKey] != learner {
				break
			}
			nextStep := &Step{}
			nextStep.Obj = pod
			start.NextSteps = append(start.NextSteps, nextStep)
			index++
		}
		if len(start.NextSteps) > 0 {
			start = start.NextSteps[0]
		}
		// append 1/2 followers
		podList := pods[index:]
		end := (len(podList) - 1) / 2
		for i := 0; i < end; i++ {
			nextStep := &Step{}
			nextStep.Obj = podList[i]
			start.NextSteps = append(start.NextSteps, nextStep)
		}

		if len(start.NextSteps) > 0 {
			start = start.NextSteps[0]
		}
		// append the other 1/2 followers
		podList = podList[end:]
		end = len(podList) - 1
		for i := 0; i < end; i++ {
			nextStep := &Step{}
			nextStep.Obj = podList[i]
			start.NextSteps = append(start.NextSteps, nextStep)
		}

		if len(start.NextSteps) > 0 {
			start = start.NextSteps[0]
		}
		// append leader
		podList = podList[end:]
		for _, pod := range podList {
			nextStep := &Step{}
			nextStep.Obj = pod
			start.NextSteps = append(start.NextSteps, nextStep)
		}
	}

	return plan
}

func getComponent(ctx context.Context, cli client.Client, cluster *dbaasv1alpha1.Cluster, typeName string) (dbaasv1alpha1.ClusterDefinitionComponent, error) {
	clusterDef := &dbaasv1alpha1.ClusterDefinition{}
	if err := cli.Get(ctx, client.ObjectKey{Name: cluster.Spec.ClusterDefRef}, clusterDef); err != nil {
		return dbaasv1alpha1.ClusterDefinitionComponent{}, err
	}

	for _, component := range clusterDef.Spec.Components {
		if component.TypeName == typeName {
			return component, nil
		}
	}

	return dbaasv1alpha1.ClusterDefinitionComponent{}, errors.New("componentDef not found: " + typeName)
}

func getComponentTypeName(cluster dbaasv1alpha1.Cluster, stsObj appsv1.StatefulSet) string {
	names := strings.Split(stsObj.Name, "-")
	name := names[len(names)-1]
	for _, component := range cluster.Spec.Components {
		if name == component.Name {
			return component.Type
		}
	}

	return name
}

func buildHeadlessSvcs(params createParams, sts *appsv1.StatefulSet) ([]client.Object, error) {
	stsPodLabels := sts.Spec.Template.Labels
	replicas := *sts.Spec.Replicas
	svcs := make([]client.Object, replicas)
	for i := 0; i < int(replicas); i++ {
		pod := &corev1.Pod{}
		pod.ObjectMeta.Name = fmt.Sprintf("%s-%d", sts.GetName(), i)
		pod.ObjectMeta.Namespace = sts.Namespace
		pod.ObjectMeta.Labels = map[string]string{
			statefulSetPodNameLabelKey: pod.ObjectMeta.Name,
			appNameLabelKey:            stsPodLabels[appNameLabelKey],
			appInstanceLabelKey:        stsPodLabels[appInstanceLabelKey],
			appComponentLabelKey:       stsPodLabels[appNameLabelKey],
		}
		pod.Spec.Containers = sts.Spec.Template.Spec.Containers

		svc, err := buildHeadlessService(params, pod)
		if err != nil {
			return nil, err
		}
		svcs[i] = svc
	}
	return svcs, nil
}

func buildSvc(params createParams) (*corev1.Service, error) {
	cueFS, _ := debme.FS(cueTemplates, "cue")

	cueTpl, err := params.getCacheCUETplValue("service_template.cue", func() (*intctrlutil.CUETpl, error) {
		return intctrlutil.NewCUETplFromBytes(cueFS.ReadFile("service_template.cue"))
	})
	if err != nil {
		return nil, err
	}

	cueValue := intctrlutil.NewCUEBuilder(*cueTpl)
	clusterStrByte, err := params.getCacheBytesValue("cluster", func() ([]byte, error) {
		return json.Marshal(params.cluster)
	})
	if err != nil {
		return nil, err
	}
	if err = cueValue.Fill("cluster", clusterStrByte); err != nil {
		return nil, err
	}

	componentStrByte, err := json.Marshal(params.component)
	if err != nil {
		return nil, err
	}
	if err = cueValue.Fill("component", componentStrByte); err != nil {
		return nil, err
	}

	svcStrByte, err := cueValue.Lookup("service")
	if err != nil {
		return nil, err
	}

	svc := corev1.Service{}
	if err = json.Unmarshal(svcStrByte, &svc); err != nil {
		return nil, err
	}

	return &svc, nil
}

func randomString(length int) string {
	res, _ := password.Generate(length, 0, 0, false, false)
	return res
}

func buildSecret(params createParams) (*corev1.Secret, error) {
	cueFS, _ := debme.FS(cueTemplates, "cue")

	cueTpl, err := params.getCacheCUETplValue("secret_template.cue", func() (*intctrlutil.CUETpl, error) {
		return intctrlutil.NewCUETplFromBytes(cueFS.ReadFile("secret_template.cue"))
	})
	if err != nil {
		return nil, err
	}

	cueValue := intctrlutil.NewCUEBuilder(*cueTpl)
	clusterDefinitionStrByte, err := params.getCacheBytesValue("clusterDefinition", func() ([]byte, error) {
		return json.Marshal(params.clusterDefinition)
	})
	if err != nil {
		return nil, err
	}

	if err = cueValue.Fill("clusterdefinition", clusterDefinitionStrByte); err != nil {
		return nil, err
	}

	clusterStrByte, err := params.getCacheBytesValue("cluster", func() ([]byte, error) {
		return json.Marshal(params.cluster)
	})
	if err != nil {
		return nil, err
	}

	if err = cueValue.Fill("cluster", clusterStrByte); err != nil {
		return nil, err
	}

	if err = cueValue.FillRaw("secret.stringData.password", randomString(8)); err != nil {
		return nil, err
	}

	secretStrByte, err := cueValue.Lookup("secret")
	if err != nil {
		return nil, err
	}

	secret := corev1.Secret{}
	if err = json.Unmarshal(secretStrByte, &secret); err != nil {
		return nil, err
	}

	return &secret, nil
}

func buildSts(params createParams) (*appsv1.StatefulSet, error) {
	cueFS, _ := debme.FS(cueTemplates, "cue")

	cueTpl, err := params.getCacheCUETplValue("statefulset_template.cue", func() (*intctrlutil.CUETpl, error) {
		return intctrlutil.NewCUETplFromBytes(cueFS.ReadFile("statefulset_template.cue"))
	})
	if err != nil {
		return nil, err
	}

	cueValue := intctrlutil.NewCUEBuilder(*cueTpl)
	clusterStrByte, err := params.getCacheBytesValue("cluster", func() ([]byte, error) {
		return json.Marshal(params.cluster)
	})
	if err != nil {
		return nil, err
	}
	if err = cueValue.Fill("cluster", clusterStrByte); err != nil {
		return nil, err
	}

	componentStrByte, err := json.Marshal(params.component)
	if err != nil {
		return nil, err
	}
	if err = cueValue.Fill("component", componentStrByte); err != nil {
		return nil, err
	}

	stsStrByte, err := cueValue.Lookup("statefulset")
	if err != nil {
		return nil, err
	}

	sts := appsv1.StatefulSet{}
	if err = json.Unmarshal(stsStrByte, &sts); err != nil {
		return nil, err
	}

	stsStrByte = injectEnv(stsStrByte, dbaasPrefix+"_MY_SECRET_NAME", params.cluster.Name)

	if err = json.Unmarshal(stsStrByte, &sts); err != nil {
		return nil, err
	}

	prefix := dbaasPrefix + "_" + strings.ToUpper(params.component.Type) + "_" + strings.ToUpper(params.component.Name) + "_"
	replicas := int(*sts.Spec.Replicas)
	for i := range sts.Spec.Template.Spec.Containers {
		// inject self scope env
		c := &sts.Spec.Template.Spec.Containers[i]
		c.Env = append(c.Env, corev1.EnvVar{
			Name: dbaasPrefix + "_MY_POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		})
		// inject roleGroup scope env
		c.Env = append(c.Env, corev1.EnvVar{
			Name:      prefix + "N",
			Value:     strconv.Itoa(replicas),
			ValueFrom: nil,
		})
		for j := 0; j < replicas; j++ {
			c.Env = append(c.Env, corev1.EnvVar{
				Name:      prefix + strconv.Itoa(j) + "_HOSTNAME",
				Value:     sts.Name + "-" + strconv.Itoa(j),
				ValueFrom: nil,
			})
		}
	}
	return &sts, nil
}

// buildConsensusSet build on a stateful set
func buildConsensusSet(params createParams) (*appsv1.StatefulSet, error) {
	sts, err := buildSts(params)
	if err != nil {
		return sts, err
	}

	sts.Spec.UpdateStrategy.Type = appsv1.OnDeleteStatefulSetStrategyType
	return sts, err
}

func buildDeploy(params createParams) (*appsv1.Deployment, error) {
	cueFS, _ := debme.FS(cueTemplates, "cue")

	cueTpl, err := params.getCacheCUETplValue("deployment_template.cue", func() (*intctrlutil.CUETpl, error) {
		return intctrlutil.NewCUETplFromBytes(cueFS.ReadFile("deployment_template.cue"))
	})
	if err != nil {
		return nil, err
	}

	cueValue := intctrlutil.NewCUEBuilder(*cueTpl)
	clusterStrByte, err := params.getCacheBytesValue("cluster", func() ([]byte, error) {
		return json.Marshal(params.cluster)
	})
	if err != nil {
		return nil, err
	}
	if err = cueValue.Fill("cluster", clusterStrByte); err != nil {
		return nil, err
	}

	componentStrByte, err := json.Marshal(params.component)
	if err != nil {
		return nil, err
	}
	if err = cueValue.Fill("component", componentStrByte); err != nil {
		return nil, err
	}

	stsStrByte, err := cueValue.Lookup("deployment")
	if err != nil {
		return nil, err
	}

	deploy := appsv1.Deployment{}
	if err = json.Unmarshal(stsStrByte, &deploy); err != nil {
		return nil, err
	}

	stsStrByte = injectEnv(stsStrByte, dbaasPrefix+"_MY_SECRET_NAME", params.cluster.Name)

	if err = json.Unmarshal(stsStrByte, &deploy); err != nil {
		return nil, err
	}

	// TODO: inject environment

	return &deploy, nil
}

func buildHeadlessService(params createParams, pod *corev1.Pod) (*corev1.Service, error) {
	cueFS, _ := debme.FS(cueTemplates, "cue")

	cueTpl, err := params.getCacheCUETplValue("headless_service_template.cue", func() (*intctrlutil.CUETpl, error) {
		return intctrlutil.NewCUETplFromBytes(cueFS.ReadFile("headless_service_template.cue"))
	})
	if err != nil {
		return nil, err
	}

	cueValue := intctrlutil.NewCUEBuilder(*cueTpl)

	podStrByte, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}

	if err = cueValue.Fill("pod", podStrByte); err != nil {
		return nil, err
	}

	svcStrByte, err := cueValue.Lookup("service")
	if err != nil {
		return nil, err
	}
	svc := corev1.Service{}
	if err = json.Unmarshal(svcStrByte, &svc); err != nil {
		return nil, err
	}

	scheme, _ := dbaasv1alpha1.SchemeBuilder.Build()
	if err = controllerutil.SetOwnerReference(params.cluster, &svc, scheme); err != nil {
		return nil, err
	}

	return &svc, nil
}

func buildPDB(params createParams) (*policyv1.PodDisruptionBudget, error) {
	cueFS, _ := debme.FS(cueTemplates, "cue")

	cueTpl, err := params.getCacheCUETplValue("pdb_template.cue", func() (*intctrlutil.CUETpl, error) {
		return intctrlutil.NewCUETplFromBytes(cueFS.ReadFile("pdb_template.cue"))
	})
	if err != nil {
		return nil, err
	}

	cueValue := intctrlutil.NewCUEBuilder(*cueTpl)

	clusterStrByte, err := params.getCacheBytesValue("cluster", func() ([]byte, error) {
		return json.Marshal(params.cluster)
	})
	if err != nil {
		return nil, err
	}
	if err = cueValue.Fill("cluster", clusterStrByte); err != nil {
		return nil, err
	}

	componentStrByte, err := json.Marshal(params.component)
	if err != nil {
		return nil, err
	}
	if err = cueValue.Fill("component", componentStrByte); err != nil {
		return nil, err
	}

	pdbStrByte, err := cueValue.Lookup("pdb")
	if err != nil {
		return nil, err
	}

	pdb := policyv1.PodDisruptionBudget{}
	if err = json.Unmarshal(pdbStrByte, &pdb); err != nil {
		return nil, err
	}

	return &pdb, nil
}

func injectEnv(strByte []byte, key string, value string) []byte {
	str := string(strByte)
	str = strings.ReplaceAll(str, "$("+key+")", value)
	return []byte(str)
}
