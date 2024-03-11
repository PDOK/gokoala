// +groupName=pdok
package crd

import (
	"github.com/PDOK/gokoala/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GoKoalaSpec struct {
	Service config.Config `json:"service"`
}

type GoKoala struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GoKoalaSpec `json:"spec,omitempty"`
}
