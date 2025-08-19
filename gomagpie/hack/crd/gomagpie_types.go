// +groupName=pdok
package crd

import (
	"github.com/PDOK/gomagpie/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GoMagpieSpec struct {
	Service config.Config `json:"service"`
}

type GoMagpie struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GoMagpieSpec `json:"spec,omitempty"`
}
