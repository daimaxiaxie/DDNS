package common

import (
	"fmt"
	"strconv"
)

var Manager = &PluginManager{
	Plugins: make(map[string]CloudProvider),
}

type PluginManager struct {
	Plugins map[string]CloudProvider
}

func (pm *PluginManager) Register(cloud CloudProvider) {
	pm.Plugins[hash(cloud.Info())] = cloud
}

func (pm *PluginManager) GetCloudProvider(name string, version int) (CloudProvider, error) {
	key := hash(CloudInfo{
		-1,
		int8(version),
		name,
	})
	cloud, ok := pm.Plugins[key]
	if !ok {
		return nil, fmt.Errorf("not found %s : %v cloud provider", name, version)
	}
	return cloud, nil
}

func hash(info CloudInfo) string {
	//buffer.WriteString
	return info.Name + strconv.Itoa(int(info.Version))
}
