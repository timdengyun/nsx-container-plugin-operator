/* Copyright © 2020 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: Apache-2.0 */

package sharedinfo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gopkg.in/ini.v1"

	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/vmware/nsx-container-plugin-operator/pkg/apis/operator/v1"
	operatortypes "github.com/vmware/nsx-container-plugin-operator/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var log = logf.Log.WithName("shared_info")

type SharedInfo struct {
	AdaptorName              string
	AddNodeTag               bool
	PodSecurityPolicySupport bool
	LastNodeAgentStartTime   map[string]time.Time
	NetworkConfig            *configv1.Network
	OperatorConfigMap        *corev1.ConfigMap
	OperatorNsxSecret        *corev1.Secret

	NsxNodeAgentDsSpec    *unstructured.Unstructured
	NsxNcpBootstrapDsSpec *unstructured.Unstructured
	NsxNcpDeploymentSpec  *unstructured.Unstructured
}

func New(mgr manager.Manager, operatorNamespace string) (*SharedInfo, error) {
	reader := mgr.GetAPIReader()
	configmap := &corev1.ConfigMap{}
	watchedNamespace := operatorNamespace
	if watchedNamespace == "" {
		log.Info(fmt.Sprintf("SharedInfo can only check a single namespace, defaulting to: %s",
			operatortypes.OperatorNamespace))
		watchedNamespace = operatortypes.OperatorNamespace
	}
	configmapName := types.NamespacedName{
		Namespace: watchedNamespace,
		Name:      operatortypes.ConfigMapName,
	}
	// r.client.Get cannot work at the stage, the solution of reader referred to
	// https://github.com/jaegertracing/jaeger-operator/pull/814
	err := reader.Get(context.TODO(), configmapName, configmap)
	if err != nil {
		log.Error(err, "Failed to get configmap")
		return nil, err
	}
	data := &configmap.Data
	cfg, err := ini.Load([]byte((*data)[operatortypes.ConfigMapDataKey]))
	if err != nil {
		log.Error(err, "Failed to get adaptor name")
		return nil, err
	}
	adaptorName := strings.ToLower(cfg.Section("coe").Key("adaptor").Value())
	ncpinstallName := types.NamespacedName{
		Name:      operatortypes.NcpInstallCRDName,
		Namespace: watchedNamespace,
	}
	ncpInstall := &operatorv1.NcpInstall{}
	err = reader.Get(context.TODO(), ncpinstallName, ncpInstall)
	if err != nil {
		log.Error(err, "Failed to get ncp-install")
		return nil, err
	}
	// The default value is true
	addNodeTag := true
	if ncpInstall.Spec.AddNodeTag == false {
		addNodeTag = false
	}

	podSecurityPolicySupport := isPodSecurityPolicySupport(mgr.GetConfig())

	return &SharedInfo{
		AdaptorName: adaptorName, AddNodeTag: addNodeTag,
		PodSecurityPolicySupport: podSecurityPolicySupport,
	}, nil
}

// PodSecurityPolicy resource is not supported any longer starting k8s >= v1.25.0
func isPodSecurityPolicySupport(c *rest.Config) bool {
	version125, _ := version.ParseGeneric("v1.25.0")

	clientset, err := clientset.NewForConfig(c)
	if err != nil {
		log.Error(err, "failed to create clientset")
		return false
	}

	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		log.Error(err, "failed to get server Kubernetes version")
		return false
	}

	runningVersion, err := version.ParseGeneric(serverVersion.String())
	if err != nil {
		log.Error(err, fmt.Sprintf("unexpected error parsing server Kubernetes version %s", runningVersion.String()))
		return false
	}

	log.Info(fmt.Sprintf("running server version is %s", runningVersion.String()))
	return runningVersion.LessThan(version125)
}
