/*
Copyright 2018 The metaGraf Authors

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

package modules

import (
	"encoding/json"
	"fmt"
	"github.com/blang/semver"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"metagraf/pkg/metagraf"
	"strconv"
	"strings"
)

// This is a complete hack. todo: fix this shit, restructure packages
var (
	NameSpace string // Used to pass namespace from cmd to module to avoid import cycle.
	Output    bool   // Flag passing hack
	Version   string // Flag passing hack
	Verbose   bool   // Flag passing hack
	Dryrun    bool   // Flag passing hack
	Branch	  string // Flag passing hack
)

var Variables map[string]string

// Returns a name for a resource based on convention as follows.
func Name(mg *metagraf.MetaGraf) string {
	var objname string

	if len(Version) > 0 {
		sv, err := semver.Parse(mg.Spec.Version)
		if err != nil {
			return strings.ToLower(mg.Metadata.Name+"-") + Version
		} else {
			objname = strings.ToLower(mg.Metadata.Name + "v" + strconv.FormatUint(sv.Major, 10))
			return objname + "-" + Version
		}
	}

	sv, err := semver.Parse(mg.Spec.Version)
	if err != nil {
		objname = strings.ToLower(mg.Metadata.Name)
	} else {
		objname = strings.ToLower(mg.Metadata.Name + "v" + strconv.FormatUint(sv.Major, 10))
	}
	return objname
}

func MGAppName(mg *metagraf.MetaGraf) string {
	var objname string

	if len(Version) > 0 {
		sv, err := semver.Parse(mg.Spec.Version)
		if err != nil {
			return mg.Metadata.Name+"-"+ Version
		} else {
			objname = mg.Metadata.Name+"v"+strconv.FormatUint(sv.Major, 10)
			return objname + "-" + Version
		}
	}

	sv, err := semver.Parse(mg.Spec.Version)
	if err != nil {
		objname = mg.Metadata.Name
	} else {
		objname = mg.Metadata.Name+"v"+strconv.FormatUint(sv.Major, 10)
	}
	return objname
}

// Returns a name for a secret for a resource based on convention as follows.
func ResourceSecretName(r *metagraf.Resource) string {
	if len(r.User) > 0 && len(r.Secret) == 0 {
		// Implicit secret name generation
		return strings.ToLower(r.User)
	} else if len(r.User) == 0 && len(r.Secret) > 0 {
		// Explicit secret name generation
		return strings.ToLower(r.Secret)
	} else {
		return strings.ToLower(r.Name)
	}
}

func ConfigSecretName(c *metagraf.Config) string {
	return strings.ToLower(c.Name)
}

// Prepends _ to indicate externally consumed variable.
func ExternalEnvToEnvVar(e *metagraf.EnvironmentVar ) corev1.EnvVar {
	v:= EnvToEnvVar(e)
	v.Name = "_"+v.Name
	return v
}

// Applies conventions and overridden logic to an environment variable and returns a corev1.EnvVar{}
func EnvToEnvVar(e *metagraf.EnvironmentVar) corev1.EnvVar {
	//fmt.Printf("Type: %t\n", e.Required)
	if e.Required {
		value := ""

		// Handle possible override value for non required fields
		if v, t := Variables[e.Name]; t {
			value = v
		}

		if len(e.Default) > 0 {
			value = e.Default
		}

		return corev1.EnvVar{
			Name:  e.Name,
			Value: value,
		}
	}

	if len(e.Default) == 0 {
		e.Default = "$" + e.Name
	}
	// Handle possible overridden values for required fields
	if v, t := Variables[e.Name]; t {
		e.Default = v
	}

	return corev1.EnvVar{
		Name:  e.Name,
		Value: e.Default,
	}
}

func ValueFromEnv(key string) bool {
	if _, t := Variables[key]; t {
		return true
	}
	return false
}

// Marshal kubernetes resource to json
func MarshalObject(obj interface{}) {
	ba, err := json.Marshal(obj)
	if err != nil {
		glog.Error(err)
	}
	fmt.Println(string(ba))
}
