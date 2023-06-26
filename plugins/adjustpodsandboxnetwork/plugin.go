/*
   Copyright The containerd Authors.

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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/containerd/nri/pkg/api"
	"github.com/containerd/nri/pkg/stub"
)

type config struct {
	CfgParam1 string `json:"cfgParam1"`
}

type plugin struct {
	stub stub.Stub
	mask stub.EventMask
}

var (
	cfg config
	log *logrus.Logger
)

func (p *plugin) Configure(config, runtime, version string) (stub.EventMask, error) {
	log.Infof("Connected to %s/%s...", runtime, version)

	if config == "" {
		return 0, nil
	}

	err := yaml.Unmarshal([]byte(config), &cfg)
	if err != nil {
		return 0, fmt.Errorf("failed to parse configuration: %w", err)
	}

	log.Info("Got configuration data %+v...", cfg)

	return 0, nil
}

func (p *plugin) Synchronize(pods []*api.PodSandbox, containers []*api.Container) ([]*api.ContainerUpdate, error) {
	log.Info("Synchronizing state with the runtime...")
	return nil, nil
}

func (p *plugin) Shutdown() {
	log.Info("Runtime shutting down...")
}

func (p *plugin) RunPodSandbox(pod *api.PodSandbox) error {
	log.Infof("Started pod %s/%s...", pod.GetNamespace(), pod.GetName())
	return nil
}

func (p *plugin) StopPodSandbox(pod *api.PodSandbox) error {
	log.Infof("Stopped pod %s/%s...", pod.GetNamespace(), pod.GetName())
	return nil
}

func (p *plugin) RemovePodSandbox(pod *api.PodSandbox) error {
	log.Infof("Removed pod %s/%s...", pod.GetNamespace(), pod.GetName())
	return nil
}

const QoSResourceNet = "net"

func (p *plugin) AdjustPodSandboxNetwork(pod *api.PodSandbox, networkconfigs []*api.NetworkConfiguration) ([]*api.NetworkConfiguration, error) {
	log.Infof("Adjust pod sandbox network for %s/%s...",  pod.GetNamespace(), pod.GetName())

	var returnconfigs []*api.NetworkConfiguration

	for _, label := range pod.Labels {
		// specific label
		if label == QoSResourceNet {
			log.Infof("Have Qos network resource %s", label)
		}
	}

	for _, annotation := range pod.Annotations {
		if annotation == QoSResourceNet {
			log.Infof("Have Qos network resource %s", annotation)
		}
	}

	returnconfigs = append(returnconfigs, &api.NetworkConfiguration {
		Bandwidth: &api.NetworkBandwidth {
			IngressRate: 1000000,
			IngressBurst: 1500000,
		},
		})

	return returnconfigs, nil
}

func (p *plugin) CreatePodSandboxNetworkConf(networkconf []*api.CreateNetworkConf) ([]*api.CreateNetworkConf, error) {
	log.Infof("Create pod sandbox network for...")

	for i, network := range networkconf {
		log.Infof("Network %d type %s...", i, network.NetworkType)
	}

	return nil, nil
}

func (p *plugin) onClose() {
	log.Infof("Connection to the runtime lost, exiting...")
	os.Exit(0)
}

func main() {
	var (
		pluginName string
		pluginIdx  string
		err        error
	)

	log = logrus.StandardLogger()
	log.SetFormatter(&logrus.TextFormatter{
		PadLevelText: true,
	})

	flag.StringVar(&pluginName, "name", "", "plugin name to register to NRI")
	flag.StringVar(&pluginIdx, "idx", "", "plugin index to register to NRI")
	flag.Parse()

	p := &plugin{}
	opts := []stub.Option{
		stub.WithOnClose(p.onClose),
	}
	if pluginName != "" {
		opts = append(opts, stub.WithPluginName(pluginName))
	}
	if pluginIdx != "" {
		opts = append(opts, stub.WithPluginIdx(pluginIdx))
	}

	if p.stub, err = stub.New(p, opts...); err != nil {
		log.Fatalf("failed to create plugin stub: %v", err)
	}

	if err = p.stub.Run(context.Background()); err != nil {
		log.Errorf("plugin exited (%v)", err)
		os.Exit(1)
	}
}
