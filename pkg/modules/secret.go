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
	"github.com/golang/glog"
	"metagraf/pkg/metagraf"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"metagraf/mg/ocpclient"
)

func GenSecrets(mg *metagraf.MetaGraf) {
	for _,r := range mg.Spec.Resources {
		// Is secret generation necessary?
		if len(r.Secret) == 0 && len(r.User) == 0 {
			glog.Info("Skipping resource: ", r.Name)
			continue
		}

		// Do not create secret if it already exist!
		if secretExists(ResourceSecretName(&r)) {
			glog.Info("Skipping resource: ", r.Name)
			continue
		}

		obj := genResourceSecret(&r, mg)
		ba, err := json.Marshal(obj)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(ba))
	}
}

// Check if a named secret exsist in the current namespace.
func secretExists(name string) bool {
	cli := ocpclient.GetCoreClient()
	l, err := cli.Secrets(NameSpace).List(metav1.ListOptions{LabelSelector:"name = "+name})

	if err != nil{
		glog.Error(err)
		os.Exit(1)
	}

	if len(l.Items) > 0 {
		glog.Info("Secret ", name, " exists in namespace: ", NameSpace)
		return true
	}
	return false
}

func genResourceSecret(res *metagraf.Resource, mg *metagraf.MetaGraf) *corev1.Secret {

	objname := Name(mg)

	// Resource labels
	l := make(map[string]string)
	l["name"] = ResourceSecretName(res)
	l["app"] = objname

	// Populate v1.Secret StringData and Data
	stringdata := make(map[string]string)
	data := make(map[string][]byte)

	if len(res.User) > 0 {
		stringdata["user"] = res.User
		stringdata["password"] = "secretstring"
	}

	if len(res.Secret) > 0 && res.SecretType == "cert" {
		data[res.Secret] = []byte("Replace this")
	}

	sec := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: l["name"],
			Labels: l,
		},
		Type: "opaque",
		StringData: stringdata,
		Data: data,
	}

	return &sec
}


