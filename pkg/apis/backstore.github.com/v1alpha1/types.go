package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BackupSpec struct {
	VolumeSnapshotName      string `json:"volume-snapshot-name"`
	VolumeSnapshotClassName string `json:"volume-snapshot-class-name"`
	PVC                     string `json:"pvc"`
	Namespace               string `json:"namespace"`
}

type RestoreSpec struct {
	BackupName              string `json:"backup-name"`
	VolumeSnapshotClassName string `json:"volume-snapshot-class-name"`
	Storage                 string `json:"storage,omitempty"`
}

type Backup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BackupSpec `json:"spec,omitempty"`
}

type BackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Backup `json:"backups,omitempty"`
}

type Restore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RestoreSpec `json:"spec,omitempty"`
}

type RestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Restore `json:"backups,omitempty"`
}
