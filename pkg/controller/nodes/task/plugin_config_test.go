package task

import (
	"context"
	"strings"
	"testing"
	"time"

	config2 "github.com/flyteorg/flytestdlib/config"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/magiconair/properties/assert"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config"
)

func TestWranglePluginsAndGenerateFinalList(t *testing.T) {
	const coreContainer = "container"
	const coreOther = "other"
	const k8sContainer = "k8s-container"
	const k8sOther = "k8s-other"
	coreContainerPlugin := core.PluginEntry{ID: coreContainer}
	coreOtherPlugin := core.PluginEntry{ID: coreOther}
	k8sContainerPlugin := k8s.PluginEntry{ID: k8sContainer}
	k8sOtherPlugin := k8s.PluginEntry{ID: k8sOther}

	cpe := func(es ...core.PluginEntry) []core.PluginEntry {
		return es
	}
	kpe := func(es ...k8s.PluginEntry) []k8s.PluginEntry {
		return es
	}
	type args struct {
		cfg         *config.TaskPluginConfig
		backOffCfg  *config.BackOffConfig
		corePlugins []core.PluginEntry
		k8sPlugins  []k8s.PluginEntry
	}
	type want struct {
		final sets.String
		err   bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"config-no-plugins", args{cfg: &config.TaskPluginConfig{EnabledPlugins: []string{coreContainer}}, backOffCfg: &config.BackOffConfig{BaseSecond: 0, MaxDuration: config2.Duration{Duration: time.Second * 0}}}, want{}},
		{"no-plugins", args{cfg: &config.TaskPluginConfig{EnabledPlugins: nil}, backOffCfg: &config.BackOffConfig{BaseSecond: 0, MaxDuration: config2.Duration{Duration: time.Second * 0}}}, want{}},
		{"no-config-no-plugins", args{}, want{err: true}},
		{"no-config-plugins", args{corePlugins: cpe(coreContainerPlugin, coreOtherPlugin), k8sPlugins: kpe(k8sContainerPlugin, k8sOtherPlugin)}, want{err: true}},
		{"empty-config-plugins", args{cfg: &config.TaskPluginConfig{EnabledPlugins: []string{}}, corePlugins: cpe(coreContainerPlugin, coreOtherPlugin), k8sPlugins: kpe(k8sContainerPlugin, k8sOtherPlugin)}, want{final: sets.NewString(k8sContainer, k8sOther, coreOther, coreContainer)}},
		{"config-plugins", args{cfg: &config.TaskPluginConfig{EnabledPlugins: []string{coreContainer, k8sOther}, DefaultForTaskTypes: map[string]string{"container": coreContainer}}, corePlugins: cpe(coreContainerPlugin, coreOtherPlugin), k8sPlugins: kpe(k8sContainerPlugin, k8sOtherPlugin), backOffCfg: &config.BackOffConfig{BaseSecond: 0, MaxDuration: config2.Duration{Duration: time.Second * 0}}}, want{final: sets.NewString(k8sOther, coreContainer)}},
		{"case-differs-config-plugins", args{cfg: &config.TaskPluginConfig{EnabledPlugins: []string{strings.ToUpper(coreContainer), strings.ToUpper(k8sOther)}, DefaultForTaskTypes: map[string]string{"container": coreContainer}}, corePlugins: cpe(coreContainerPlugin, coreOtherPlugin), k8sPlugins: kpe(k8sContainerPlugin, k8sOtherPlugin), backOffCfg: &config.BackOffConfig{BaseSecond: 0, MaxDuration: config2.Duration{Duration: time.Second * 0}}}, want{final: sets.NewString(k8sOther, coreContainer)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &testPluginRegistry{
				core: tt.args.corePlugins,
				k8s:  tt.args.k8sPlugins,
			}
			got, _, err := WranglePluginsAndGenerateFinalList(context.TODO(), tt.args.cfg, pr)
			if (err != nil) != tt.want.err {
				t.Errorf("WranglePluginsAndGenerateFinalList() error = %v, wantErr %v", err, tt.want.err)
				return
			}
			s := sets.NewString()
			for _, g := range got {
				s.Insert(g.ID)
			}
			assert.Equal(t, s.List(), tt.want.final.List())
		})
	}
}
