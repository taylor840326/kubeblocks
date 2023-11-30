/*
Copyright (C) 2022-2023 ApeCloud Co., Ltd

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
	"strconv"

	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"

	appsv1alpha1 "github.com/apecloud/kubeblocks/apis/apps/v1alpha1"
	"github.com/apecloud/kubeblocks/pkg/constant"
	intctrlutil "github.com/apecloud/kubeblocks/pkg/controllerutil"
	viper "github.com/apecloud/kubeblocks/pkg/viperx"
)

func BuildWeSyncer(reqCtx intctrlutil.RequestCtx, synthesizeComp *SynthesizedComponent) error {
	// If it's not a built-in handler supported by Lorry, LorryContainers are not injected by default.
	haType := getWeSyncerType(synthesizeComp)
	if haType == "" {
		return nil
	}

	container := buildBasicContainer(synthesizeComp)
	weSyncerSvcHTTPPort := viper.GetInt32(constant.KBEnvWeSyncerHTTPPort)
	availablePorts, err := getAvailableContainerPorts(synthesizeComp.PodSpec.Containers, []int32{weSyncerSvcHTTPPort})
	if err != nil {
		reqCtx.Log.Info("get lorry container port failed", "error", err)
		return err
	}
	weSyncerSvcHTTPPort = availablePorts[0]

	initContainer := container.DeepCopy()
	buildWeSyncerInitContainer(synthesizeComp, initContainer)
	modifyMainContainerForWesyncer(synthesizeComp, int(weSyncerSvcHTTPPort))
	synthesizeComp.PodSpec.InitContainers = append(synthesizeComp.PodSpec.InitContainers, *initContainer)

	return nil
}

func buildWeSyncerInitContainer(component *SynthesizedComponent, container *corev1.Container) {
	container.Image = viper.GetString(constant.KBWesyncerImage)
	container.Name = constant.WesyncerInitContainerName
	container.ImagePullPolicy = corev1.PullPolicy(viper.GetString(constant.KBImagePullPolicy))
	container.Command = []string{"cp", "-r", "/bin/wesyncer", "/config", "/kubeblocks/"}
	container.StartupProbe = nil
	container.ReadinessProbe = nil
	volumeMount := corev1.VolumeMount{Name: "kubeblocks", MountPath: "/kubeblocks"}
	container.VolumeMounts = []corev1.VolumeMount{volumeMount}
}

func modifyMainContainerForWesyncer(component *SynthesizedComponent, wesyncerSvcHTTPPort int) {
	container := component.PodSpec.Containers[0]
	command := []string{"/kubeblocks/wesyncer", "run",
		"--dapr-http-port", strconv.Itoa(wesyncerSvcHTTPPort),
		"--"}
	container.Command = append(command, container.Command...)
	volumeMount := corev1.VolumeMount{Name: "kubeblocks", MountPath: "/kubeblocks"}
	container.VolumeMounts = append(container.VolumeMounts, volumeMount)
	container.Env = append(container.Env, buildContainerEnvs(component)...)

	container.Ports = append(container.Ports, corev1.ContainerPort{
		ContainerPort: int32(wesyncerSvcHTTPPort),
		Name:          constant.WesyncerHTTPPortName,
		Protocol:      "TCP",
	})
	component.PodSpec.Containers[0] = container
}

func getWeSyncerType(synthesizeComp *SynthesizedComponent) string {
	var haType string
	if synthesizeComp.LifecycleActions.RoleProbe != nil && synthesizeComp.LifecycleActions.RoleProbe.BuiltinHandler != nil {
		haType = string(*synthesizeComp.LifecycleActions.RoleProbe.BuiltinHandler)
		if slices.Contains(constant.WeSyncerSupportTypes, haType) {
			return haType
		}
	}

	if synthesizeComp.CharacterType != "" {
		haType = synthesizeComp.CharacterType
		if slices.Contains(constant.WeSyncerSupportTypes, haType) {
			return haType
		}
	}

	if *synthesizeComp.RoleArbitrator == appsv1alpha1.WesyncerRoleArbitrator && haType != "" {
		return haType
	}

	return ""
}
